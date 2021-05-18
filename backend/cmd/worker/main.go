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
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var tr = otel.Tracer("cmd/worker/main")

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
	ctx, span := tr.Start(ctx, "store state")
	//ctx, cancel := context.WithTimeout(ctx, 30*time.Second)

	// defer func() {
	// 	//cancel()
	span.End()
	// }()

	_, err := s.StoreState(ctx, st)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}

func main() {
	addr := "otel-agent:4317"
	shutdown := tracer_provider.Initialize(addr, "api")
	defer shutdown()

	var (
		dURL  = mustGetEnvStr("DATA_STORE_URL")
		dPool = mustGetEnvInt("DATA_STORE_POOL_SIZE")
		qURL  = mustGetEnvStr("QUEUE_URL")
		qPool = mustGetEnvInt("QUEUE_POOL_SIZE")
		qKey  = mustGetEnvStr("QUEUE_KEY")
	)

	q := broker.New(mustNewRedisClient(qURL, qPool))
	s := store.New(mustNewRedisClient(dURL, dPool))

	br, err := q.RemoteSubscribe(context.Background(), qKey)
	if err != nil {
		panic(err)
	}

	for rs := range br {
		sctx := trace.NewSpanContext(
			trace.SpanContextConfig{
				TraceID: rs.TraceID,
				SpanID:  rs.SpanID,
				Remote:  rs.Remote,
			},
		)

		ctx := trace.ContextWithRemoteSpanContext(context.Background(), sctx)
		go storeState(ctx, rs.State, s)
	}
}
