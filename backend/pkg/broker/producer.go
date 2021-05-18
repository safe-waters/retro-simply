package broker

import (
	"go.opentelemetry.io/otel/propagation"
)

var _ propagation.TextMapCarrier = (*ProducerMessageCarrier)(nil)

type ProducerMessageCarrier struct {
	RemoteState *Message
}

func NewProducerMessageCarrier(rs *Message) *ProducerMessageCarrier {
	return &ProducerMessageCarrier{RemoteState: rs}
}

func (p *ProducerMessageCarrier) Get(key string) string {
	if key == "traceparent" {
		// https://www.w3.org/TR/trace-context/#traceparent-header
		// https://github.com/open-telemetry/opentelemetry-go/blob/d616df61f5d163589228c5ff3be4aa5415f5a884/propagation/trace_context_test.go#L38
		// equivalent: "00-" + hex trace id + "-" + hex span id + "-01"
		return p.RemoteState.TraceParent
	}

	return ""
}

func (p *ProducerMessageCarrier) Set(key string, value string) {
	// TODO: trace state
	if key == "traceparent" {
		p.RemoteState.TraceParent = value
	}
}

func (p *ProducerMessageCarrier) Keys() []string {
	return []string{"traceparent"}
}
