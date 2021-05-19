package auth

import (
	"context"
	"fmt"
	"testing"
)

var password = "test"

func TestHashPassword(t *testing.T) {
	t.Parallel()

	hashPassword(t, password)
}

func TestCompareAndHashPassword(t *testing.T) {
	t.Parallel()

	p, h := hashPassword(t, password)
	if err := p.CompareHashAndPassword(
		context.Background(),
		h,
		password,
	); err != nil {
		t.Fatal(err)
	}
}

func TestInvalidCompareAndHashPassword(t *testing.T) {
	t.Parallel()

	p, h := hashPassword(t, password)
	if err := p.CompareHashAndPassword(
		context.Background(),
		h,
		fmt.Sprintf("wrong%s", password),
	); err == nil {
		t.FailNow()
	}
}

func hashPassword(t *testing.T, password string) (*PasswordManager, string) {
	p := NewPasswordManager()

	h, err := p.HashPassword(context.Background(), "test")
	if err != nil {
		t.Fatal(err)
	}

	if len(h) == 0 {
		t.FailNow()
	}

	return p, h
}
