package broker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/safe-waters/retro-simply/backend/pkg/client"
	"github.com/safe-waters/retro-simply/backend/pkg/data"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

type RemoteState struct {
	State   *data.State
	TraceID [16]byte
	SpanID  [8]byte
	Remote  bool
}

var tr = otel.Tracer("pkg/broker/broker")

type PubSuber interface {
	Publish(ctx context.Context, channel string, message interface{}) client.Err
	Subscribe(ctx context.Context, channels ...string) client.PubSubChannel
}

type B struct{ ps PubSuber }

func New(ps PubSuber) *B { return &B{ps: ps} }

func (b *B) Subscribe(
	ctx context.Context,
	rId string,
) (<-chan *data.State, error) {
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
	sCh := make(chan *data.State)

	go func() {
		ctx, span := tr.Start(ctx, "broker listening")
		defer span.End()

		defer close(sCh)
		defer p.Close()

		for {
			select {
			case msg := <-pCh:
				s := &data.State{}
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

func (b *B) Publish(ctx context.Context, rId string, s *data.State) error {
	ctx, span := tr.Start(ctx, "broker publish")
	defer span.End()

	byt, err := json.Marshal(s)
	if err != nil {
		return err
	}

	if err = b.ps.Publish(ctx, rId, byt).Err(); err != nil {
		return err
	}

	return nil
}

func (b *B) RemotePublish(ctx context.Context, rId string, s *data.State) error {
	ctx, span := tr.Start(ctx, "broker remote publish")
	defer span.End()

	sctx := span.SpanContext()

	rs := &RemoteState{
		State:   s,
		TraceID: sctx.TraceID(),
		SpanID:  sctx.SpanID(),
		Remote:  true,
	}
	byt, err := json.Marshal(rs)
	if err != nil {
		return err
	}
	fmt.Println("THIS IS THE REMOTE PUBLISH", string(byt))

	if err = b.ps.Publish(ctx, rId, byt).Err(); err != nil {
		return err
	}

	return nil
}

func (b *B) RemoteSubscribe(
	ctx context.Context,
	rId string,
) (<-chan *RemoteState, error) {
	ctx, span := tr.Start(ctx, "broker remote subscribe")
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
	sCh := make(chan *RemoteState)

	go func() {
		ctx, span := tr.Start(ctx, "remote broker listening")
		defer span.End()

		defer close(sCh)
		defer p.Close()

		for {
			select {
			case msg := <-pCh:
				s := &RemoteState{}
				err := json.Unmarshal([]byte(msg.Payload), s)
				if err != nil {
					span.RecordError(err)
					fmt.Println("ERROR OCCURED", err)

					continue
				}

				fmt.Printf("THIS IS THE REMOTE SUBSCRIBE %+v", s)

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
