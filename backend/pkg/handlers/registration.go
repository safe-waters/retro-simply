package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/safe-waters/retro-simply/backend/pkg/auth"
	"github.com/safe-waters/retro-simply/backend/pkg/data"
	"github.com/safe-waters/retro-simply/backend/pkg/store"
	"go.opentelemetry.io/otel"
)

var regTr = otel.Tracer("pkg/handlers/registration")

var _ http.Handler = (*Registration)(nil)

type PasswordHashStorer interface {
	HashedPassword(ctx context.Context, rId string) (string, error)
	StoreHashedPassword(ctx context.Context, rId, h string) error
}

type TokenSetter interface {
	SetToken(ctx context.Context, w http.ResponseWriter, c *auth.Claims) error
}

type PasswordHashComparer interface {
	HashPassword(ctx context.Context, p string) (string, error)
	CompareHashAndPassword(ctx context.Context, h, p string) error
}

type Registration struct {
	route string
	phs   PasswordHashStorer
	ts    TokenSetter
	phc   PasswordHashComparer
}

func NewRegistration(
	route string,
	phs PasswordHashStorer,
	ts TokenSetter,
	phc PasswordHashComparer,
) *Registration {
	return &Registration{
		route: route,
		phs:   phs,
		ts:    ts,
		phc:   phc,
	}
}

func (rg *Registration) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := regTr.Start(r.Context(), "handlers serve http")
	defer span.End()

	p := strings.TrimPrefix(r.URL.Path, rg.route)
	switch p {
	case "create":
		rg.create(ctx, w, r)
	case "join":
		rg.join(ctx, w, r)
	default:
		err := fmt.Errorf("'%s' not found", p)
		span.RecordError(err)

		http.NotFound(w, r)
	}
}

func (rg *Registration) create(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
) {
	ctx, span := regTr.Start(ctx, "handlers create")
	defer span.End()

	rm, err := rg.decodeRoom(ctx, w, r)
	if err != nil {
		span.RecordError(err)
		return
	}

	h, err := rg.phc.HashPassword(ctx, rm.Password)
	if err != nil {
		span.RecordError(err)
		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)

		return
	}

	if err := rg.phs.StoreHashedPassword(ctx, rm.Id, h); err != nil {
		span.RecordError(err)

		switch err.(type) {
		case store.DataAlreadyExistsError:
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		default:
			http.Error(
				w,
				http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError,
			)
			return
		}
	}

	if err := rg.setToken(ctx, rm.Id, r, w); err != nil {
		span.RecordError(err)
		return
	}

	w.Header().Set(
		"Content-Location",
		fmt.Sprintf("/retrospective?roomId=%s", rm.Id),
	)
	w.WriteHeader(http.StatusCreated)
}

func (rg *Registration) join(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
) {
	ctx, span := regTr.Start(ctx, "handlers join")
	defer span.End()

	room, err := rg.decodeRoom(ctx, w, r)
	if err != nil {
		span.RecordError(err)
		return
	}

	h, err := rg.phs.HashedPassword(ctx, room.Id)
	if err != nil {
		span.RecordError(err)

		switch err.(type) {
		case store.DataDoesNotExistError:
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		default:
			http.Error(
				w,
				http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError,
			)
			return
		}
	}

	if err := rg.phc.CompareHashAndPassword(ctx, h, room.Password); err != nil {
		span.RecordError(err)
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	if err := rg.setToken(ctx, room.Id, r, w); err != nil {
		span.RecordError(err)
		return
	}

	w.Header().Set(
		"Content-Location",
		fmt.Sprintf("/retrospective?roomId=%s", room.Id),
	)
	w.WriteHeader(http.StatusOK)
}

func (rg *Registration) decodeRoom(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
) (*data.Room, error) {
	_, span := regTr.Start(ctx, "handlers decode room")
	defer span.End()

	d := json.NewDecoder(r.Body)

	var room data.Room
	if err := d.Decode(&room); err != nil {
		span.RecordError(err)

		var msg string
		switch err.(type) {
		case data.PasswordInvalidError, data.RoomIdInvalidError:
			msg = err.Error()
		default:
			msg = http.StatusText(http.StatusBadRequest)
		}

		http.Error(w, msg, http.StatusBadRequest)

		return nil, err
	}

	return &room, nil
}

func (rg *Registration) setToken(
	ctx context.Context,
	roomId string,
	r *http.Request,
	w http.ResponseWriter,
) error {
	ctx, span := regTr.Start(ctx, "handlers set token")
	defer span.End()

	c := auth.NewClaims(roomId, time.Now().UTC().Add(time.Hour*24*7))
	if err := rg.ts.SetToken(ctx, w, c); err != nil {
		span.RecordError(err)
		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)

		return err
	}

	return nil
}
