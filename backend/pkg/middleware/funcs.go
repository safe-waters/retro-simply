package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/safe-waters/retro-simply/backend/pkg/auth"
	"github.com/safe-waters/retro-simply/backend/pkg/data"
	"github.com/safe-waters/retro-simply/backend/pkg/user"
	"go.opentelemetry.io/otel"
)

var tr = otel.Tracer("pkg/middleware")

type TokenValidator interface {
	ValidateToken(
		ctx context.Context,
		r *http.Request,
		cc *auth.ComparisonClaims,
	) error
}

func AuthFunc(t TokenValidator, route string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, span := tr.Start(r.Context(), "auth middleware")
			defer span.End()

			rId := strings.TrimPrefix(r.URL.Path, route)
			if !data.RoomIDRegex.MatchString(rId) {
				err := fmt.Errorf("invalid room id '%s'", rId)
				span.RecordError(err)

				http.Error(
					w,
					http.StatusText(http.StatusBadRequest),
					http.StatusBadRequest,
				)

				return
			}

			c := auth.NewComparisonClaims(rId)
			if err := t.ValidateToken(r.Context(), r, c); err != nil {
				span.RecordError(err)

				http.Error(
					w,
					http.StatusText(http.StatusBadRequest),
					http.StatusBadRequest,
				)

				return
			}

			u, _ := user.FromContext(r.Context())
			u.RoomId = rId

			ctx := user.WithContext(r.Context(), u)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func JSONContentTypeFunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, span := tr.Start(r.Context(), "JSON content type middleware")
		defer span.End()

		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func MethodTypeFunc(t string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, span := tr.Start(r.Context(), "method type middleware")
			defer span.End()

			if r.Method != t {
				span.RecordError(fmt.Errorf("'%s' not allowed", r.Method))
				http.Error(
					w,
					http.StatusText(http.StatusMethodNotAllowed),
					http.StatusMethodNotAllowed,
				)

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
