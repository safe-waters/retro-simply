package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/safe-waters/retro-simply/backend/pkg/auth"
	"github.com/safe-waters/retro-simply/backend/pkg/broker"
	"github.com/safe-waters/retro-simply/backend/pkg/client"
	"github.com/safe-waters/retro-simply/backend/pkg/handlers"
	"github.com/safe-waters/retro-simply/backend/pkg/middleware"
	"github.com/safe-waters/retro-simply/backend/pkg/store"
	"github.com/safe-waters/retro-simply/backend/pkg/tracer_provider"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

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

func applyMiddleware(h http.Handler, mwfs ...func(next http.Handler) http.Handler) http.Handler {
	for i := len(mwfs) - 1; i >= 0; i-- {
		h = mwfs[i](h)
	}

	return h
}

func main() {
	addr := "otel-agent:4317"
	shutdown := tracer_provider.Initialize(addr, "api")
	defer shutdown()

	var (
		dURL    = mustGetEnvStr("DATA_STORE_URL")
		bURL    = mustGetEnvStr("BROKER_URL")
		qURL    = mustGetEnvStr("QUEUE_URL")
		port    = mustGetEnvStr("PORT")
		version = mustGetEnvStr("VERSION")
		secret  = mustGetEnvStr("SECRET")
		dPool   = mustGetEnvInt("DATA_STORE_POOL_SIZE")
		bPool   = mustGetEnvInt("BROKER_POOL_SIZE")
		qPool   = mustGetEnvInt("QUEUE_POOL_SIZE")
		qKey    = mustGetEnvStr("QUEUE_KEY")
	)

	s := store.New(mustNewRedisClient(dURL, dPool))
	b := broker.New(mustNewRedisClient(bURL, bPool))
	q := broker.New(mustNewRedisClient(qURL, qPool))

	j := auth.NewJWT([]byte(secret))
	pm := auth.NewPasswordManager()

	apiRoute := fmt.Sprintf("/api/%s", version)
	regRoute := fmt.Sprintf("%s/registration/", apiRoute)
	retRoute := fmt.Sprintf("%s/retrospectives/", apiRoute)

	reg := applyMiddleware(
		handlers.NewRegistration(
			regRoute,
			s,
			j,
			pm,
		),
		middleware.MethodTypeFunc(http.MethodPost),
		middleware.CorrelationIDFunc,
		middleware.JSONContentTypeFunc,
	)

	ret := applyMiddleware(
		handlers.NewRetrospective(
			s,
			b,
			q,
			qKey,
		),
		middleware.MethodTypeFunc(http.MethodGet),
		middleware.CorrelationIDFunc,
		middleware.AuthFunc(j, retRoute),
	)

	http.Handle(regRoute, otelhttp.NewHandler(reg, regRoute))
	http.Handle(retRoute, otelhttp.NewHandler(ret, retRoute))

	http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}
