package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"go.opentelemetry.io/otel"
)

var jTr = otel.Tracer("pkg/auth/jwt")

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
	_, span := jTr.Start(ctx, "auth set token")
	defer span.End()

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	signedT, err := t.SignedString(j.secret)
	if err != nil {
		span.RecordError(err)
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
	_, span := jTr.Start(ctx, "auth validate token")
	defer span.End()

	ck, err := r.Cookie("token")
	if err != nil {
		span.RecordError(err)
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
		span.RecordError(err)
		return err
	}

	if !t.Valid {
		err := errors.New("invalid token")
		span.RecordError(err)

		return err
	}

	c, ok := t.Claims.(*Claims)
	if !ok {
		err := errors.New("invalid claims")
		span.RecordError(err)

		return err
	}

	if c.RoomId != cc.RoomId {
		err := fmt.Errorf(
			"claims id: '%s' does not match room id: '%s'",
			c.RoomId,
			cc.RoomId,
		)
		span.RecordError(err)

		return err
	}

	return nil
}
