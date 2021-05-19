package auth

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel"
	"golang.org/x/crypto/bcrypt"
)

var pTr = otel.Tracer("pkg/auth/password")

type PasswordManager struct{}

func NewPasswordManager() *PasswordManager { return &PasswordManager{} }

func (pm *PasswordManager) HashPassword(ctx context.Context, p string) (string, error) {
	_, span := pTr.Start(ctx, "auth hash password")
	defer span.End()

	h, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	if err != nil {
		span.RecordError(err)
		return "", err
	}

	return string(h), nil
}

func (pm *PasswordManager) CompareHashAndPassword(ctx context.Context, h, p string) error {
	_, span := pTr.Start(ctx, "auth compare hash and password")
	defer span.End()

	if err := bcrypt.CompareHashAndPassword([]byte(h), []byte(p)); err != nil {
		err := errors.New("incorrect password")
		span.RecordError(err)

		return err
	}

	return nil
}
