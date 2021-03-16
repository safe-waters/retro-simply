package auth_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/safe-waters/retro-simply/backend/pkg/auth"
)

var (
	future  = time.Now().UTC().Add(time.Hour * 24 * 7)
	expired = time.Now().UTC().Add(time.Hour * -1)
	secret  = []byte("secret")
	rId     = "test"
)

func TestSetToken(t *testing.T) {
	t.Parallel()

	res, _, _ := setToken(t, future)
	expectCookie(t, res, future)
}

func TestValidateToken(t *testing.T) {
	t.Parallel()

	res, j, c := setToken(t, future)
	expectCookie(t, res, future)

	cc := auth.NewComparisonClaims(c.RoomId)
	ck := res.Result().Cookies()[0]
	if err := validateToken(t, ck, j, cc); err != nil {
		t.Fatal(err)
	}
}

func TestValidateInvalidComparisonClaims(t *testing.T) {
	t.Parallel()

	res, j, c := setToken(t, future)
	expectCookie(t, res, future)

	cc := auth.NewComparisonClaims(fmt.Sprintf("wrong%s", c.RoomId))
	ck := res.Result().Cookies()[0]
	if err := validateToken(t, ck, j, cc); err == nil {
		t.FailNow()
	}
}

func TestValidateInvalidSignature(t *testing.T) {
	t.Parallel()

	res, j, c := setToken(t, future)
	expectCookie(t, res, future)

	cc := auth.NewComparisonClaims(c.RoomId)
	ck := res.Result().Cookies()[0]

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	signedTok, err := tok.SignedString([]byte(fmt.Sprintf("wrong%s", secret)))
	if err != nil {
		t.Fatal(err)
	}

	ck.Value = signedTok
	if err := validateToken(t, ck, j, cc); err == nil {
		t.FailNow()
	}
}

func TestValidateExpiredToken(t *testing.T) {
	t.Parallel()

	res, j, c := setToken(t, expired)
	expectCookie(t, res, expired)

	cc := auth.NewComparisonClaims(c.RoomId)
	ck := res.Result().Cookies()[0]
	if err := validateToken(t, ck, j, cc); err == nil {
		t.FailNow()
	}
}

func setToken(
	t *testing.T,
	expiration time.Time,
) (*httptest.ResponseRecorder, *auth.JWT, *auth.Claims) {
	t.Helper()

	j := auth.NewJWT([]byte(secret))
	res := httptest.NewRecorder()

	c := auth.NewClaims(rId, expiration)
	j.SetToken(context.Background(), res, c)

	return res, j, c
}

func validateToken(
	t *testing.T,
	cookie *http.Cookie,
	jwtAuth *auth.JWT,
	comparisonClaims *auth.ComparisonClaims,
) error {
	t.Helper()

	req, err := http.NewRequest("", "", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.AddCookie(cookie)

	return jwtAuth.ValidateToken(context.Background(), req, comparisonClaims)
}

func expectCookie(
	t *testing.T,
	res *httptest.ResponseRecorder,
	expiration time.Time,
) {
	t.Helper()

	const (
		numCk      = 1
		ckName     = "token"
		ckHttpOnly = true
		ckSameSite = http.SameSiteStrictMode
		ckPath     = "/"
		ckSecure   = true
	)

	if len(res.Result().Cookies()) != numCk {
		t.Fatalf(
			"expected %d cookie, got: %d",
			numCk,
			len(res.Result().Cookies()),
		)
	}

	ck := res.Result().Cookies()[0]
	if ck.Name != ckName {
		t.Fatalf("expected cookie name %s, got: %s", ckName, ck.Name)
	}

	if ck.HttpOnly != ckHttpOnly {
		t.Fatalf("expected HttpOnly to be %t, got: %t", ckHttpOnly, ck.HttpOnly)
	}

	if ck.SameSite != ckSameSite {
		t.Fatalf(
			"expected SameSite policy %v, got: %v",
			ckSameSite,
			ck.SameSite,
		)
	}

	if ck.Path != ckPath {
		t.Fatalf("expected path %s, got %s", ckPath, ck.Path)
	}

	if ck.Secure != ckSecure {
		t.Fatalf("expected secure %t,  got: %t", ckSecure, ck.Secure)
	}

	if !ck.Expires.Round(time.Minute).Equal(expiration.Round(time.Minute)) {
		t.Fatalf(
			"expected expiration %s, got: %s",
			expiration.Round(time.Minute),
			ck.Expires.Round(time.Minute),
		)
	}
}
