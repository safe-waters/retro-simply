package broker

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/safe-waters/retro-simply/backend/pkg/client"
	"github.com/safe-waters/retro-simply/backend/pkg/data"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
)

type Message struct {
	State *data.State
	// Since redis' pubsub protocol does not have headers like the
	// HTTP protocol, use the span context to set the same headers that
	// would be in an HTTP request. Specifically, the 'traceparent' header
	// that contains the trace ID and span ID:
	// https://github.com/open-telemetry/opentelemetry-go/blob/d616df61f5d163589228c5ff3be4aa5415f5a884/propagation/trace_context_test.go#L38
	// https://www.w3.org/TR/trace-context/#traceparent-header
	Headers http.Header
}

var tr = otel.Tracer("pkg/broker/broker")

type PubSuber interface {
	Publish(ctx context.Context, channel string, message interface{}) client.Err
	Subscribe(ctx context.Context, channels ...string) client.PubSubChannel
}

type B struct{ ps PubSuber }

func New(ps PubSuber) *B { return &B{ps: ps} }

func (b *B) Publish(ctx context.Context, rId string, s *data.State) error {
	ctx, span := tr.Start(ctx, "broker publish")
	defer span.End()

	rs := &Message{State: s, Headers: map[string][]string{}}

	var pr propagation.TraceContext
	pr.Inject(ctx, propagation.HeaderCarrier(rs.Headers))

	byt, err := json.Marshal(rs)
	if err != nil {
		return err
	}

	if err = b.ps.Publish(ctx, rId, byt).Err(); err != nil {
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
		span.SetStatus(codes.Error, err.Error())

		_ = p.Close()

		return nil, err
	}

	pCh := p.Channel()
	sCh := make(chan *Message)

	go func() {
		ctx, span := tr.Start(ctx, "broker listening")
		defer span.End()

		defer close(sCh)
		defer p.Close()

		for {
			select {
			case msg := <-pCh:
				s := &Message{}
				err := json.Unmarshal([]byte(msg.Payload), s)
				if err != nil {
					span.RecordError(err)

					continue
				}

				select {
				case sCh <- s:
				case <-ctx.Done():
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return sCh, nil
}
