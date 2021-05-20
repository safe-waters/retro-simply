package broker

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/safe-waters/retro-simply/backend/pkg/client"
	"github.com/safe-waters/retro-simply/backend/pkg/data"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

var tr = otel.Tracer("pkg/broker")

type Message struct {
	State *data.State
	// Since redis' pubsub protocol does not have headers like the
	// HTTP protocol, use the span context to set the same headers that
	// would be in an HTTP request. Specifically, the 'traceparent' header
	// that contains the trace ID and span ID:
	// https://github.com/open-telemetry/opentelemetry-go/blob/d616df61f5d163589228c5ff3be4aa5415f5a884/propagation/trace_context_test.go#L38
	// https://www.w3.org/TR/trace-context/#traceparent-header
	Header http.Header
}

type PubSuber interface {
	Publish(ctx context.Context, channel string, message interface{}) client.Err
	Subscribe(ctx context.Context, channels ...string) client.PubSubChannel
}

type B struct{ ps PubSuber }

func New(ps PubSuber) *B { return &B{ps: ps} }

func (b *B) Publish(ctx context.Context, rId string, s *data.State) error {
	ctx, span := tr.Start(ctx, "broker publish")
	defer span.End()

	m := &Message{State: s, Header: http.Header{}}

	var pr propagation.TraceContext
	pr.Inject(ctx, propagation.HeaderCarrier(m.Header))

	byt, err := json.Marshal(m)
	if err != nil {
		span.RecordError(err)
		return err
	}

	if err = b.ps.Publish(ctx, rId, byt).Err(); err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}

func (b *B) Subscribe(
	ctx context.Context,
	rId string,
) (<-chan *Message, error) {
	ctx, span := tr.Start(ctx, "broker subscribe")
	defer span.End()

	p := b.ps.Subscribe(ctx, rId)

	// Ensure subscription is created before returning channel
	_, err := p.Receive(ctx)
	if err != nil {
		span.RecordError(err)

		_ = p.Close()

		return nil, err
	}

	pCh := p.Channel()
	mCh := make(chan *Message)

	go func() {
		ctx, span := tr.Start(ctx, "broker listening")
		defer span.End()

		defer close(mCh)
		defer p.Close()

		for {
			select {
			case rawMsg := <-pCh:
				m := &Message{}
				err := json.Unmarshal([]byte(rawMsg.Payload), m)
				if err != nil {
					span.RecordError(err)

					continue
				}

				select {
				case mCh <- m:
				case <-ctx.Done():
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return mCh, nil
}
