package broker

import (
	"encoding/hex"
	"fmt"

	"go.opentelemetry.io/otel/propagation"
)

type TextMapCarrier interface {
	// Get returns the value associated with the passed key.
	Get(key string) string
	// Set stores the key-value pair.
	Set(key string, value string)
	// Keys lists the keys stored in this carrier.
	Keys() []string
}

var _ propagation.TextMapCarrier = (*ProducerMessageCarrier)(nil)

type ProducerMessageCarrier struct {
	RemoteState *RemoteState
}

// https://www.w3.org/TR/trace-context/#traceparent-header
// https://github.com/open-telemetry/opentelemetry-go/blob/d616df61f5d163589228c5ff3be4aa5415f5a884/propagation/trace_context_test.go#L38
func NewProducerMessageCarrier(rs *RemoteState) *ProducerMessageCarrier {
	return &ProducerMessageCarrier{RemoteState: rs}
}

func (p *ProducerMessageCarrier) Get(key string) string {
	//tp := "00-" + string(p.RemoteState.TraceID[:]) + "-" + string(p.RemoteState.SpanID[:]) + "-08"

	// equivalent header
	// sctx := trace.NewSpanContext(
	// 	trace.SpanContextConfig{
	// 		TraceID: rs.TraceID,
	// 		SpanID:  rs.SpanID,
	// 		Remote:  rs.Remote,
	// 	},
	// )

	tp := "00-" + hex.EncodeToString(p.RemoteState.TraceID[:]) + "-" + hex.EncodeToString(p.RemoteState.SpanID[:]) + "-00"
	fmt.Println("THIS IS THE TRACEPARENT", tp)
	return tp
}

func (p *ProducerMessageCarrier) Set(key string, value string) {
	fmt.Println("SET CALLED")
}

func (p *ProducerMessageCarrier) Keys() []string {
	return []string{"traceparent"}
}
