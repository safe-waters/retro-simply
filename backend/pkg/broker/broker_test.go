package broker

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/safe-waters/retro-simply/backend/pkg/client"
	"github.com/safe-waters/retro-simply/backend/pkg/data"
)

var baseState = `{
    "ws": null,
    "connected": false,
    "errorMessage": "",
    "roomId": "test",
    "columns": [
        {
            "id": "0",
            "title": "Good",
            "cardStyle": {
                "backgroundColor": "bg-danger"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "0",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        },
        {
            "id": "1",
            "title": "Bad",
            "cardStyle": {
                "backgroundColor": "bg-primary"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "1",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        },
        {
            "id": "3",
            "title": "Actions",
            "cardStyle": {
                "backgroundColor": "bg-success"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "3",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        }
    ]
}`

type mockErr struct{}

func (m *mockErr) Err() error { return nil }

type mockPubSubClient struct{ mps *mockPubSubChannel }

func newMockPubSubClient() *mockPubSubClient {
	return &mockPubSubClient{mps: newMockPubSubChannel()}
}

func (m *mockPubSubClient) Subscribe(
	ctx context.Context,
	channels ...string,
) client.PubSubChannel {
	return m.mps
}

func (m *mockPubSubClient) Publish(
	ctx context.Context,
	channel string,
	message interface{},
) client.Err {
	byt := message.([]byte)

	go func() {
		m.mps.msgCh <- &redis.Message{Payload: string(byt)}
	}()

	return &mockErr{}
}

type mockPubSubChannel struct {
	msgCh    chan *redis.Message
	closeSpy int
}

func newMockPubSubChannel() *mockPubSubChannel {
	return &mockPubSubChannel{msgCh: make(chan *redis.Message)}
}

func (m *mockPubSubChannel) Receive(ctx context.Context) (interface{}, error) {
	return nil, nil
}

func (m *mockPubSubChannel) Channel() <-chan *redis.Message { return m.msgCh }

func (m *mockPubSubChannel) Close() error {
	m.closeSpy++
	return nil
}

func TestBroker(t *testing.T) {
	m := newMockPubSubClient()
	b := New(m)

	var es data.State
	if err := json.Unmarshal([]byte(baseState), &es); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	rId := "test"

	sCh, err := b.Subscribe(ctx, rId)
	if err != nil {
		t.Fatal(err)
	}

	b.Publish(ctx, rId, &es)

	gs := <-sCh
	expectState(t, &es, gs.State)

	cancel()

	expectCloseStateChannel(t, sCh)
	expectClosePubSubChannel(t, 1, m.mps.closeSpy)
}

func expectState(t *testing.T, expected, got interface{}) {
	t.Helper()

	if !reflect.DeepEqual(expected, got) {
		pe := prettify(t, expected)
		pg := prettify(t, got)
		t.Fatalf("expected: '%s', got: '%s'", pe, pg)
	}
}

func expectCloseStateChannel(t *testing.T, sCh <-chan *Message) {
	select {
	case _, ok := <-sCh:
		if ok {
			t.Fatal("expected state channel to close")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("expected state channel to close")
	}
}

func expectClosePubSubChannel(t *testing.T, expected, got int) {
	if expected != got {
		t.Fatalf("expected: '%d' calls to Close, got: '%d'", expected, got)
	}
}

func prettify(t *testing.T, v interface{}) interface{} {
	t.Helper()

	prettyV, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		t.Fatal(err)
	}

	return string(prettyV)
}
