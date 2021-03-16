package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"sync"
	"testing"

	"github.com/safe-waters/retro-simply/backend/pkg/auth"
	"github.com/safe-waters/retro-simply/backend/pkg/data"
	"github.com/safe-waters/retro-simply/backend/pkg/store"
)

type mockPasswordStore struct {
	data map[string]interface{}
	mu   *sync.Mutex
}

func newMockPasswordStore() *mockPasswordStore {
	return &mockPasswordStore{
		data: map[string]interface{}{},
		mu:   &sync.Mutex{},
	}
}

func (m *mockPasswordStore) State(
	ctx context.Context,
	rId string,
) (*data.State, error) {
	return nil, nil
}

func (m *mockPasswordStore) StoreState(
	ctx context.Context,
	s *data.State,
) (*data.State, error) {
	return nil, nil
}

func (m *mockPasswordStore) HashedPassword(
	ctx context.Context,
	rId string,
) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	p, ok := m.data[rId]
	if !ok {
		return "", store.DataDoesNotExistError{
			Err: errors.New("room does not exist"),
		}
	}

	pStr := p.(string)

	return pStr, nil
}

func (m *mockPasswordStore) StoreHashedPassword(
	ctx context.Context,
	rId string,
	h string,
) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.data[rId]; ok {
		return store.DataAlreadyExistsError{
			Err: errors.New("room already exists"),
		}
	}

	m.data[rId] = h

	return nil
}

type erroneousMockStoreHashedPassword struct{ PasswordHashStorer }

func (e *erroneousMockStoreHashedPassword) StoreHashedPassword(
	ctx context.Context,
	rId string,
	h string,
) error {
	return errors.New("")
}

type erroneousMockGetHashedPassword struct{ PasswordHashStorer }

func (e *erroneousMockGetHashedPassword) HashedPassword(
	ctx context.Context,
	rId string,
) (string, error) {
	return "", errors.New("")
}

type erroneousMockHashPassword struct{ PasswordHashComparer }

func (e *erroneousMockHashPassword) HashPassword(p string) (string, error) {
	return "", errors.New("")
}

type erroneousMockCompareHashAndPassword struct{ PasswordHashComparer }

func (e *erroneousMockCompareHashAndPassword) CompareHashAndPassword(
	h,
	p string,
) error {
	return errors.New("")
}

type erroneousMockSetToken struct{ *auth.JWT }

func (e *erroneousMockSetToken) SetToken(
	ctx context.Context,
	w http.ResponseWriter,
	c *auth.Claims,
) error {
	return errors.New("")
}

func TestCreateRoom(t *testing.T) {
	t.Parallel()

	b := map[string]string{"id": "test", "password": "test"}
	phc := auth.NewPasswordManager()
	ts := auth.NewJWT([]byte("secret"))
	phs := newMockPasswordStore()

	res := postRequest(t, "create", b, phc, ts, phs)
	expectRegistration(t, res, http.StatusCreated)
}

func TestCreateDuplicateRoom(t *testing.T) {
	t.Parallel()

	b := map[string]string{"id": "test", "password": "test"}
	phc := auth.NewPasswordManager()
	ts := auth.NewJWT([]byte("secret"))
	phs := newMockPasswordStore()

	res := postRequest(t, "create", b, phc, ts, phs)
	expectRegistration(t, res, http.StatusCreated)

	res = postRequest(t, "create", b, phc, ts, phs)
	expectRegistration(t, res, http.StatusBadRequest)
}

func TestCreateWithInvalidPassword(t *testing.T) {
	t.Parallel()

	b := map[string]string{"id": "test", "password": ""}
	phc := auth.NewPasswordManager()
	ts := auth.NewJWT([]byte("secret"))
	phs := newMockPasswordStore()

	res := postRequest(t, "create", b, phc, ts, phs)
	expectRegistration(t, res, http.StatusBadRequest)
}

func TestCreateWithInvalidRoomId(t *testing.T) {
	t.Parallel()

	b := map[string]string{"id": "", "password": "test"}
	phc := auth.NewPasswordManager()
	ts := auth.NewJWT([]byte("secret"))
	phs := newMockPasswordStore()

	res := postRequest(t, "create", b, phc, ts, phs)
	expectRegistration(t, res, http.StatusBadRequest)
}

func TestCreateWithWrongKeys(t *testing.T) {
	t.Parallel()

	b := map[string]string{"missing": ""}
	phc := auth.NewPasswordManager()
	ts := auth.NewJWT([]byte("secret"))
	phs := newMockPasswordStore()

	res := postRequest(t, "create", b, phc, ts, phs)
	expectRegistration(t, res, http.StatusBadRequest)
}

func TestCreateUnableToHashPassword(t *testing.T) {
	t.Parallel()

	b := map[string]string{"id": "test", "password": "test"}
	phc := &erroneousMockHashPassword{auth.NewPasswordManager()}
	ts := auth.NewJWT([]byte("secret"))
	phs := newMockPasswordStore()

	res := postRequest(t, "create", b, phc, ts, phs)
	expectRegistration(t, res, http.StatusInternalServerError)
}

func TestCreateUnableToStorePassword(t *testing.T) {
	t.Parallel()

	b := map[string]string{"id": "test", "password": "test"}
	phc := auth.NewPasswordManager()
	ts := auth.NewJWT([]byte("secret"))
	phs := &erroneousMockStoreHashedPassword{newMockPasswordStore()}

	res := postRequest(t, "create", b, phc, ts, phs)
	expectRegistration(t, res, http.StatusInternalServerError)
}

func TestCreateUnableToSetToken(t *testing.T) {
	t.Parallel()

	b := map[string]string{"id": "test", "password": "test"}
	phc := auth.NewPasswordManager()
	ts := &erroneousMockSetToken{auth.NewJWT([]byte("secret"))}
	phs := newMockPasswordStore()

	res := postRequest(t, "create", b, phc, ts, phs)
	expectRegistration(t, res, http.StatusInternalServerError)
}

func TestJoinRoom(t *testing.T) {
	t.Parallel()

	b := map[string]string{"id": "test", "password": "test"}
	phc := auth.NewPasswordManager()
	ts := auth.NewJWT([]byte("secret"))
	phs := newMockPasswordStore()

	res := postRequest(t, "create", b, phc, ts, phs)
	expectRegistration(t, res, http.StatusCreated)

	res = postRequest(t, "join", b, phc, ts, phs)
	expectRegistration(t, res, http.StatusOK)
}

func TestJoinRoomDoesNotExist(t *testing.T) {
	t.Parallel()

	b := map[string]string{"id": "test", "password": "test"}
	phc := auth.NewPasswordManager()
	ts := auth.NewJWT([]byte("secret"))
	phs := newMockPasswordStore()

	res := postRequest(t, "join", b, phc, ts, phs)
	expectRegistration(t, res, http.StatusBadRequest)
}

func TestUnableToGetHashedPassword(t *testing.T) {
	t.Parallel()

	b := map[string]string{"id": "test", "password": "test"}
	phc := auth.NewPasswordManager()
	ts := auth.NewJWT([]byte("secret"))
	phs := &erroneousMockGetHashedPassword{newMockPasswordStore()}

	res := postRequest(t, "create", b, phc, ts, phs)
	expectRegistration(t, res, http.StatusCreated)

	res = postRequest(t, "join", b, phc, ts, phs)
	expectRegistration(t, res, http.StatusInternalServerError)
}

func TestJoinWithInvalidPassword(t *testing.T) {
	t.Parallel()

	b := map[string]string{"id": "test", "password": ""}
	phc := auth.NewPasswordManager()
	ts := auth.NewJWT([]byte("secret"))
	phs := newMockPasswordStore()

	res := postRequest(t, "join", b, phc, ts, phs)
	expectRegistration(t, res, http.StatusBadRequest)
}

func TestJoinWithInvalidRoomId(t *testing.T) {
	t.Parallel()

	b := map[string]string{"id": "", "password": "test"}
	phc := auth.NewPasswordManager()
	ts := auth.NewJWT([]byte("secret"))
	phs := newMockPasswordStore()

	res := postRequest(t, "join", b, phc, ts, phs)
	expectRegistration(t, res, http.StatusBadRequest)
}

func TestJoinWithWrongKeys(t *testing.T) {
	t.Parallel()

	b := map[string]string{"missing": ""}
	phc := auth.NewPasswordManager()
	ts := auth.NewJWT([]byte("secret"))
	phs := newMockPasswordStore()

	res := postRequest(t, "join", b, phc, ts, phs)
	expectRegistration(t, res, http.StatusBadRequest)
}

func TestJoinRoomUnableToCompareHashAndPassword(t *testing.T) {
	t.Parallel()

	b := map[string]string{"id": "test", "password": "test"}
	phc := &erroneousMockCompareHashAndPassword{auth.NewPasswordManager()}
	ts := auth.NewJWT([]byte("secret"))
	phs := newMockPasswordStore()

	res := postRequest(t, "create", b, phc, ts, phs)
	expectRegistration(t, res, http.StatusCreated)

	res = postRequest(t, "join", b, phc, ts, phs)
	expectRegistration(t, res, http.StatusBadRequest)
}

func TestJoinUnableToSetToken(t *testing.T) {
	t.Parallel()

	b := map[string]string{"id": "test", "password": "test"}
	phc := auth.NewPasswordManager()
	ts := auth.NewJWT([]byte("secret"))
	phs := newMockPasswordStore()

	res := postRequest(t, "create", b, phc, ts, phs)
	expectRegistration(t, res, http.StatusCreated)

	ets := &erroneousMockSetToken{auth.NewJWT([]byte("secret"))}

	res = postRequest(t, "join", b, phc, ets, phs)
	expectRegistration(t, res, http.StatusInternalServerError)
}

func TestInvalidRoute(t *testing.T) {
	t.Parallel()

	b := map[string]string{"id": "test", "password": "test"}
	phc := auth.NewPasswordManager()
	ts := auth.NewJWT([]byte("secret"))
	phs := newMockPasswordStore()

	res := postRequest(t, "/wrong", b, phc, ts, phs)
	expectRegistration(t, res, http.StatusNotFound)
}

func expectRegistration(t *testing.T, res *httptest.ResponseRecorder, code int) {
	t.Helper()

	resCode := res.Result().StatusCode

	if resCode != code {
		t.Fatalf("expected status code: %d, got: %d", code, resCode)
	}

	if code != http.StatusCreated && code != http.StatusOK {
		const numCk = 0

		if len(res.Result().Cookies()) != numCk {
			t.Fatalf(
				"expected %d cookies, got: %d",
				numCk,
				len(res.Result().Cookies()),
			)
		}

		const numHeaders = 0
		if len(res.Result().Header["Content-Location"]) != numHeaders {
			t.Fatalf(
				"expected: %d Content-Location headers, got: %d",
				numHeaders,
				len(res.Result().Header["Content-Location"]),
			)
		}

		return
	}

	const numCk = 1

	if len(res.Result().Cookies()) != numCk {
		t.Fatalf(
			"expected %d cookie, got: %d",
			numCk,
			len(res.Result().Cookies()),
		)
	}

	const numH = 1

	h := res.Result().Header["Content-Location"]
	if len(h) != numH {
		t.Fatalf(
			"expected %d Content-Location headers, got: %d",
			numH,
			len(h),
		)
	}

	route, err := regexp.Compile("^/retrospective\\?roomId=[a-zA-Z0-9_-]+$")
	if err != nil {
		t.Fatal(err)
	}

	if !route.MatchString(h[0]) {
		t.Fatalf(
			"expected Content-Location header to match regex %s, got: %s",
			route,
			h[0],
		)
	}
}

func postRequest(
	t *testing.T,
	route string,
	body map[string]string,
	phc PasswordHashComparer,
	ts TokenSetter,
	phs PasswordHashStorer,
) *httptest.ResponseRecorder {
	t.Helper()

	byt, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}

	regRoute := "/api/v1/registration/"
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s%s", regRoute, route),
		bytes.NewReader(byt),
	)
	if err != nil {
		t.Fatal(err)
	}

	r := NewRegistration(regRoute, phs, ts, phc)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	return res
}
