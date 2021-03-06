package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/safe-waters/retro-simply/backend/pkg/broker"
	"github.com/safe-waters/retro-simply/backend/pkg/client"
	"github.com/safe-waters/retro-simply/backend/pkg/data"
	"github.com/safe-waters/retro-simply/backend/pkg/store"
	"github.com/safe-waters/retro-simply/backend/pkg/tracer_provider"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

var tr = otel.Tracer("cmd/worker")

func mustGetEnvStr(k string) string {
	v := os.Getenv(k)
	if v == "" {
		panic(fmt.Sprintf("'%s' environment variable missing", k))
	}

	return v
}

func mustGetEnvInt(k string) int {
	vs := mustGetEnvStr(k)

	v, err := strconv.Atoi(vs)
	if err != nil {
		panic(
			fmt.Sprintf(
				"'%s' environment variable cannot be parsed to an integer",
				k,
			),
		)
	}

	return v
}

func mustNewRedisClient(url string, poolSize int) *client.C {
	c, err := client.New(url, poolSize)
	if err != nil {
		panic(err)
	}

	a := time.After(60 * time.Second)
	t := time.NewTicker(3 * time.Second)
loop:
	for {
		select {
		case <-t.C:
			if err := c.Ping(context.Background()).Err(); err == nil {
				break loop
			}
		case <-a:
			panic("timeout connecting to redis")
		}
	}

	return c
}

func storeState(ctx context.Context, st *data.State, s *store.S) {
	ctx, span := tr.Start(ctx, "worker store state")
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)

	defer func() {
		cancel()
		span.End()
	}()

	_, err := s.StoreState(ctx, st)
	if err != nil {
		span.RecordError(err)
	}
}

func main() {
	var (
		otelURL = mustGetEnvStr("OTEL_AGENT_URL")
		dURL    = mustGetEnvStr("DATA_STORE_URL")
		dPool   = mustGetEnvInt("DATA_STORE_POOL_SIZE")
		qURL    = mustGetEnvStr("QUEUE_URL")
		qPool   = mustGetEnvInt("QUEUE_POOL_SIZE")
		qKey    = mustGetEnvStr("QUEUE_KEY")
	)

	shutdown := tracer_provider.Initialize(otelURL, "worker")
	defer shutdown()

	q := broker.New(mustNewRedisClient(qURL, qPool))
	s := store.New(mustNewRedisClient(dURL, dPool))

	msgs, err := q.Subscribe(context.Background(), qKey)
	if err != nil {
		panic(err)
	}

	for m := range msgs {
		var pr propagation.TraceContext
		ctx := pr.Extract(
			context.Background(),
			propagation.HeaderCarrier(m.Header),
		)

		go storeState(ctx, m.State, s)
	}
}
