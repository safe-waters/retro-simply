package handlers

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"net/http"
// 	"net/http/httptest"
// 	"reflect"
// 	"strings"
// 	"testing"

// 	"github.com/gorilla/websocket"

// 	"github.com/safe-waters/retro-simply/backend/pkg/data"
// 	"github.com/safe-waters/retro-simply/backend/pkg/user"
// )

// const baseState = `{
//     "ws": null,
//     "connected": false,
//     "errorMessage": "",
//     "roomId": "%s",
//     "columns": [
//         {
//             "id": "0",
//             "title": "Good",
//             "cardStyle": {
//                 "backgroundColor": "bg-danger"
//             },
//             "groups": [
//                 {
//                     "id": "default",
//                     "columnId": "0",
//                     "isEditable": false,
//                     "title": "ungrouped cards",
//                     "retroCards": []
//                 }
//             ]
//         },
//         {
//             "id": "1",
//             "title": "Bad",
//             "cardStyle": {
//                 "backgroundColor": "bg-primary"
//             },
//             "groups": [
//                 {
//                     "id": "default",
//                     "columnId": "1",
//                     "isEditable": false,
//                     "title": "ungrouped cards",
//                     "retroCards": []
//                 }
//             ]
//         },
//         {
//             "id": "3",
//             "title": "Actions",
//             "cardStyle": {
//                 "backgroundColor": "bg-success"
//             },
//             "groups": [
//                 {
//                     "id": "default",
//                     "columnId": "3",
//                     "isEditable": false,
//                     "title": "ungrouped cards",
//                     "retroCards": []
//                 }
//             ]
//         }
//     ]
// }`

// type mockStateStore struct{}

// func newMockStateStore() *mockStateStore { return &mockStateStore{} }

// func (m *mockStateStore) State(
// 	ctx context.Context,
// 	rId string,
// ) (*data.State, error) {
// 	var s data.State
// 	if err := json.Unmarshal(
// 		[]byte(fmt.Sprintf(baseState, rId)),
// 		&s,
// 	); err != nil {
// 		return nil, err
// 	}

// 	return &s, nil
// }

// type mockBroker struct {
// 	ch chan *data.State
// }

// func newMockBroker() *mockBroker {
// 	ch := make(chan *data.State)
// 	return &mockBroker{
// 		ch: ch,
// 	}
// }

// func (m *mockBroker) Subscribe(
// 	ctx context.Context,
// 	rId string,
// ) (<-chan *data.State, error) {
// 	return m.ch, nil
// }

// func (m *mockBroker) Publish(
// 	ctx context.Context,
// 	rId string,
// 	s *data.State,
// ) error {
// 	go func() {
// 		m.ch <- s
// 	}()

// 	return nil
// }

// func mockUserMiddleware(rId string) func(next http.Handler) http.Handler {
// 	return func(next http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			u, _ := user.FromContext(r.Context())
// 			u.RoomId = rId
// 			ctx := user.WithContext(r.Context(), u)
// 			next.ServeHTTP(w, r.WithContext(ctx))
// 		})
// 	}
// }

// func TestRetrospective(t *testing.T) {
// 	ms := newMockStateStore()
// 	mb := newMockBroker()
// 	mq := newMockBroker()

// 	retRoute := "/api/v1/retrospectives/"
// 	rId := "test"
// 	ret := mockUserMiddleware(rId)(NewRetrospective(ms, mb, mq, rId))

// 	r := http.NewServeMux()
// 	r.Handle(retRoute, ret)

// 	s := httptest.NewServer(r)
// 	defer s.Close()

// 	u := fmt.Sprintf(
// 		"ws%s%s",
// 		strings.TrimPrefix(s.URL, "http"),
// 		fmt.Sprintf("%s%s", retRoute, rId),
// 	)

// 	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer ws.Close()

// 	const numMessages = 10

// 	var stateToSend data.State

// 	if err := json.Unmarshal(
// 		[]byte(fmt.Sprintf(baseState, rId)),
// 		&stateToSend,
// 	); err != nil {
// 		t.Fatal(err)
// 	}

// 	for i := 0; i < numMessages; i++ {
// 		if err := ws.WriteJSON(&stateToSend); err != nil {
// 			t.Fatal(err)
// 		}
// 	}

// 	for i := 0; i < numMessages+1; i++ {
// 		var stateToReceive data.State
// 		if err := ws.ReadJSON(&stateToReceive); err != nil {
// 			t.Fatal(err)
// 		}

// 		expectState(t, &stateToSend, &stateToReceive)
// 	}
// }

// func expectState(t *testing.T, expected interface{}, got interface{}) {
// 	t.Helper()

// 	if !reflect.DeepEqual(expected, got) {
// 		pe := prettify(t, expected)
// 		pg := prettify(t, got)
// 		t.Fatalf("expected: '%s', got: '%s'", pe, pg)
// 	}
// }

// func prettify(t *testing.T, v interface{}) interface{} {
// 	t.Helper()

// 	prettyV, err := json.MarshalIndent(v, "", "\t")
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	return string(prettyV)
// }
