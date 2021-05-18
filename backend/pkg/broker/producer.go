package broker

import (
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

type ProducerMessageCarrier struct{}

func NewProducerMessageCarrier(msg string) {
}

func (p *ProducerMessageCarrier) Get(key string) string { return "" }

func (p *ProducerMessageCarrier) Set(key string, value string) {

}

func (p *ProducerMessageCarrier) Keys() []string {
	var pr propagation.TraceContext
	return pr.Fields()
	// // p.Set(key string, value string)
	// pr.Inject(context.Background(), p)
	// pr.Extract(context.Background(), p)

	//return nil
}
