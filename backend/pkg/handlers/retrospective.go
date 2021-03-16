package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/safe-waters/retro-simply/backend/pkg/data"
	"github.com/safe-waters/retro-simply/backend/pkg/logger"
	"github.com/safe-waters/retro-simply/backend/pkg/store"
	"github.com/safe-waters/retro-simply/backend/pkg/user"
)

var _ http.Handler = (*Retrospective)(nil)

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
	u, ok := user.FromContext(r.Context())
	if !ok || u.RoomId == "" {
		logger.Error(r.Context(), fmt.Errorf("user '%v' incorrectly set", u))

		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)

		return
	}

	wsc, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
	if err != nil {
		logger.Error(r.Context(), err)

		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)

		return
	}

	c := newClient(wsc, rt.ps, rt.p, rt.st, rt.pKey)

	ctx := user.WithContext(context.Background(), u)
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
	logger.Info(ctx, "client started")

	ctx, cancel := context.WithCancel(ctx)

	go func() {
		<-c.wDone
		<-c.rDone

		cancel()
		c.wsc.Close()

		logger.Info(ctx, "client ended")
	}()

	br, err := c.ps.Subscribe(ctx, rId)
	if err != nil {
		logger.Error(ctx, err)

		close(c.wDone)
		close(c.rDone)

		return
	}

	s, err := c.st.State(ctx, rId)
	if err != nil {
		switch err.(type) {
		case store.DataDoesNotExistError:
			go c.readMessages(ctx, rId)
			go c.writeMessages(ctx, br)

			return
		default:
			logger.Error(ctx, err)

			close(c.wDone)
			close(c.rDone)

			return
		}
	}

	if err := c.wsc.WriteJSON(s); err != nil {
		logger.Error(ctx, err)

		close(c.wDone)
		close(c.rDone)

		return
	}

	go c.readMessages(ctx, rId)
	go c.writeMessages(ctx, br)
}

func (c *client) readMessages(ctx context.Context, rId string) {
	logger.Info(ctx, "read loop started")
	defer func() {
		logger.Info(ctx, "read loop ended")
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
				logger.Info(ctx, "could not read state")
				return
			}

			if s.RoomId != rId {
				logger.Error(
					ctx,
					errors.New("read state contains the wrong roomId"),
				)
				return
			}

			if err := c.ps.Publish(ctx, rId, &s); err != nil {
				logger.Error(ctx, err)
				return
			}

			if err := c.p.Publish(ctx, c.pKey, &s); err != nil {
				logger.Error(ctx, err)
				return
			}
		}
	}
}

func (c *client) writeMessages(ctx context.Context, br <-chan *data.State) {
	logger.Info(ctx, "write loop started")
	defer func() {
		logger.Info(ctx, "write loop ended")
		close(c.wDone)
	}()

	t := time.NewTicker(pPeriod)
	defer t.Stop()

	for {
		select {
		case s, ok := <-br:
			if !ok {
				logger.Error(ctx, errors.New("broadcast closed"))
				return
			}

			_ = c.wsc.SetWriteDeadline(time.Now().Add(wWait))
			if err := c.wsc.WriteJSON(s); err != nil {
				logger.Error(ctx, err)
				return
			}
		case <-t.C:
			_ = c.wsc.SetWriteDeadline(time.Now().Add(wWait))
			if err := c.wsc.WriteMessage(
				websocket.PingMessage, nil,
			); err != nil {
				logger.Error(ctx, err)
				return
			}
		case <-ctx.Done():
			return
		case <-c.rDone:
			return
		}
	}
}
