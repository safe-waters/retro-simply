package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/safe-waters/retro-simply/backend/pkg/data"
	"github.com/safe-waters/retro-simply/backend/pkg/store"
	"github.com/safe-waters/retro-simply/backend/pkg/user"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var _ http.Handler = (*Retrospective)(nil)

var retTr = otel.Tracer("pkg/handlers/retrospective")

type Stater interface {
	State(ctx context.Context, rId string) (*data.State, error)
}

type Puber interface {
	Publish(ctx context.Context, rId string, s *data.State) error
}

type PubSuber interface {
	Subscribe(ctx context.Context, rId string) (<-chan *data.State, error)
	Puber
}

type Retrospective struct {
	st   Stater
	ps   PubSuber
	p    Puber
	pKey string
}

func NewRetrospective(
	st Stater,
	ps PubSuber,
	p Puber,
	pKey string,
) *Retrospective {
	return &Retrospective{
		st:   st,
		ps:   ps,
		p:    p,
		pKey: pKey,
	}
}

func (rt *Retrospective) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := retTr.Start(r.Context(), "ServeHTTP")
	defer span.End()

	u, ok := user.FromContext(ctx)
	if !ok || u.RoomId == "" {
		err := fmt.Errorf("user '%v' incorrectly set", u)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)

		return
	}

	wsc, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)

		return
	}

	c := newClient(wsc, rt.ps, rt.p, rt.st, rt.pKey)

	go c.run(ctx, u.RoomId)
}

const (
	wWait   = 10 * time.Second
	pWait   = 60 * time.Second
	pPeriod = (pWait * 9) / 10
)

type wsConn interface {
	WriteJSON(v interface{}) error
	WriteMessage(messageType int, data []byte) error
	ReadJSON(v interface{}) error
	Close() error
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
	SetPongHandler(h func(string) error)
}

type client struct {
	wsc   wsConn
	ps    PubSuber
	p     Puber
	st    Stater
	pKey  string
	wDone chan struct{}
	rDone chan struct{}
}

func newClient(
	wsc wsConn,
	ps PubSuber,
	p Puber,
	st Stater,
	pKey string,
) *client {
	return &client{
		wsc:   wsc,
		ps:    ps,
		p:     p,
		st:    st,
		pKey:  pKey,
		wDone: make(chan struct{}),
		rDone: make(chan struct{}),
	}
}

func (c *client) run(ctx context.Context, rId string) {
	span := trace.SpanFromContext(ctx)
	ctx = trace.ContextWithSpan(context.Background(), span)

	ctx, span = retTr.Start(ctx, "run")
	ctx, cancel := context.WithCancel(ctx)

	span.AddEvent("client started")

	go func(ctx context.Context) {
		defer span.End()

		<-c.wDone
		<-c.rDone

		cancel()
		c.wsc.Close()

		span.AddEvent("client ended")
	}(ctx)

	br, err := c.ps.Subscribe(ctx, rId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		close(c.wDone)
		close(c.rDone)

		return
	}

	s, err := c.st.State(ctx, rId)
	if err != nil {
		switch err.(type) {
		case store.DataDoesNotExistError:
			span.AddEvent("data does not exist")

			go c.readMessages(ctx, rId)
			go c.writeMessages(ctx, br)

			return
		default:
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())

			close(c.wDone)
			close(c.rDone)

			return
		}
	}

	if err := c.wsc.WriteJSON(s); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		close(c.wDone)
		close(c.rDone)

		return
	}

	go c.readMessages(ctx, rId)
	go c.writeMessages(ctx, br)
}

func (c *client) readMessages(ctx context.Context, rId string) {
	ctx, span := retTr.Start(ctx, "readMessages")

	span.AddEvent("read loop started")

	defer func() {
		close(c.rDone)

		span.AddEvent("read loop ended")
		span.End()
	}()

	_ = c.wsc.SetReadDeadline(time.Now().Add(pWait))
	c.wsc.SetPongHandler(func(string) error {
		_ = c.wsc.SetReadDeadline(time.Now().Add(pWait))
		return nil
	})

	for {
		select {
		case <-c.wDone:
			return
		case <-ctx.Done():
			return
		default:
			var s data.State

			if err := c.wsc.ReadJSON(&s); err != nil {
				span.RecordError(err)
				return
			}

			if s.RoomId != rId {
				err := errors.New("read state contains the wrong roomId")
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())

				return
			}

			if err := c.ps.Publish(ctx, rId, &s); err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())

				return
			}

			if err := c.p.Publish(ctx, c.pKey, &s); err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())

				return
			}
		}
	}
}

func (c *client) writeMessages(ctx context.Context, br <-chan *data.State) {
	ctx, span := retTr.Start(ctx, "writeMessages")

	span.AddEvent("write loop started")

	defer func() {
		close(c.wDone)

		span.AddEvent("write loop ended")
		span.End()
	}()

	t := time.NewTicker(pPeriod)
	defer t.Stop()

	for {
		select {
		case s, ok := <-br:
			if !ok {
				err := errors.New("broadcast closed")
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())

				return
			}

			_ = c.wsc.SetWriteDeadline(time.Now().Add(wWait))
			if err := c.wsc.WriteJSON(s); err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())

				return
			}
		case <-t.C:
			_ = c.wsc.SetWriteDeadline(time.Now().Add(wWait))
			if err := c.wsc.WriteMessage(
				websocket.PingMessage, nil,
			); err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())

				return
			}
		case <-ctx.Done():
			return
		case <-c.rDone:
			return
		}
	}
}
