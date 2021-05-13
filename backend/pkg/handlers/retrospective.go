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

	r = r.WithContext(ctx)

	u, ok := user.FromContext(r.Context())
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

	ctx = user.WithContext(context.Background(), u)
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
	ctx, cancel := context.WithCancel(ctx)
	ctx, span := retTr.Start(ctx, "run")
	defer span.End()

	go func(ctx context.Context) {
		_, span := retTr.Start(ctx, "wait")
		defer span.End()

		span.AddEvent("client started")

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
	defer span.End()

	span.AddEvent("read loop started")

	defer func() {
		span.AddEvent("read loop ended")
		close(c.rDone)
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
				span.AddEvent("could not read state")
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
	defer span.End()

	span.AddEvent("write loop started")

	defer func() {
		span.AddEvent("write loop ended")
		close(c.wDone)
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
