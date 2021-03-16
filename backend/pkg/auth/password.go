package auth

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type PasswordManager struct{}

func NewPasswordManager() *PasswordManager { return &PasswordManager{} }

func (pm *PasswordManager) HashPassword(p string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(h), nil
}

func (pm *PasswordManager) CompareHashAndPassword(h, p string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(h), []byte(p)); err != nil {
		return errors.New("incorrect password")
	}

	return nil
}
