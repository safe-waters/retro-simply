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
	"go.opentelemetry.io/otel/codes"
)

var _ http.Handler = (*Registration)(nil)

var regTr = otel.Tracer("pkg/handlers/registration")

type PasswordHashStorer interface {
	HashedPassword(ctx context.Context, rId string) (string, error)
	StoreHashedPassword(ctx context.Context, rId string, h string) error
}

type TokenSetter interface {
	SetToken(ctx context.Context, w http.ResponseWriter, c *auth.Claims) error
}

type PasswordHashComparer interface {
	HashPassword(p string) (string, error)
	CompareHashAndPassword(h, p string) error
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
	ctx, span := regTr.Start(r.Context(), "ServeHTTP")
	defer span.End()

	r = r.WithContext(ctx)

	p := strings.TrimPrefix(r.URL.Path, rg.route)
	switch p {
	case "create":
		rg.create(w, r)
	case "join":
		rg.join(w, r)
	default:
		err := fmt.Errorf("'%s' not found", p)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		http.NotFound(w, r)
	}
}

func (rg *Registration) create(w http.ResponseWriter, r *http.Request) {
	_, span := regTr.Start(r.Context(), "create")
	defer span.End()

	rm, err := rg.decodeRoom(w, r)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return
	}

	h, err := rg.phc.HashPassword(rm.Password)
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

	if err := rg.phs.StoreHashedPassword(r.Context(), rm.Id, h); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

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

	if err := rg.setToken(rm.Id, r, w); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return
	}

	w.Header().Set(
		"Content-Location",
		fmt.Sprintf("/retrospective?roomId=%s", rm.Id),
	)
	w.WriteHeader(http.StatusCreated)
}

func (rg *Registration) join(w http.ResponseWriter, r *http.Request) {
	_, span := regTr.Start(r.Context(), "join")
	defer span.End()

	room, err := rg.decodeRoom(w, r)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return
	}

	h, err := rg.phs.HashedPassword(r.Context(), room.Id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

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

	if err := rg.phc.CompareHashAndPassword(h, room.Password); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := rg.setToken(room.Id, r, w); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return
	}

	w.Header().Set(
		"Content-Location",
		fmt.Sprintf("/retrospective?roomId=%s", room.Id),
	)
	w.WriteHeader(http.StatusOK)
}

func (rg *Registration) decodeRoom(
	w http.ResponseWriter,
	r *http.Request,
) (*data.Room, error) {
	d := json.NewDecoder(r.Body)

	var room data.Room
	if err := d.Decode(&room); err != nil {
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
	roomId string,
	r *http.Request,
	w http.ResponseWriter,
) error {
	c := auth.NewClaims(roomId, time.Now().UTC().Add(time.Hour*24*7))
	if err := rg.ts.SetToken(r.Context(), w, c); err != nil {
		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
		return err
	}

	return nil
}
