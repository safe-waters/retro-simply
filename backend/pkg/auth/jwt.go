package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type JWT struct{ secret []byte }

func NewJWT(secret []byte) *JWT { return &JWT{secret: secret} }

type ComparisonClaims struct {
	RoomId string `json:"roomId"`
}

func NewComparisonClaims(rId string) *ComparisonClaims {
	return &ComparisonClaims{RoomId: rId}
}

type Claims struct {
	*ComparisonClaims
	*jwt.StandardClaims
}

func NewClaims(rId string, exp time.Time) *Claims {
	stdC := &jwt.StandardClaims{ExpiresAt: exp.Unix()}
	cc := &ComparisonClaims{RoomId: rId}

	return &Claims{
		ComparisonClaims: cc,
		StandardClaims:   stdC,
	}
}

func (j *JWT) SetToken(
	ctx context.Context,
	w http.ResponseWriter,
	c *Claims,
) error {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	signedT, err := t.SignedString(j.secret)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    signedT,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Unix(c.ExpiresAt, 0),
		Path:     "/",
	})

	return nil
}

func (j *JWT) ValidateToken(
	ctx context.Context,
	r *http.Request,
	cc *ComparisonClaims,
) error {
	ck, err := r.Cookie("token")
	if err != nil {
		return err
	}

	signedCk := ck.Value
	c := &Claims{}

	t, err := jwt.ParseWithClaims(
		signedCk,
		c,
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf(
					"invalid signing method '%T'", token.Method,
				)
			}

			return j.secret, nil
		},
	)
	if err != nil {
		return err
	}

	if !t.Valid {
		return errors.New("invalid token")
	}

	c, ok := t.Claims.(*Claims)
	if !ok {
		return errors.New("invalid claims")
	}

	if c.RoomId != cc.RoomId {
		return fmt.Errorf(
			"claims id: '%s' does not match room id: '%s'",
			c.RoomId,
			cc.RoomId,
		)
	}

	return nil
}