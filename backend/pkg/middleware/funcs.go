package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/safe-waters/retro-simply/backend/pkg/auth"
	"github.com/safe-waters/retro-simply/backend/pkg/data"
	"github.com/safe-waters/retro-simply/backend/pkg/logger"
	"github.com/safe-waters/retro-simply/backend/pkg/user"
)

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
			rId := strings.TrimPrefix(r.URL.Path, route)
			if !data.RoomIDRegex.MatchString(rId) {
				logger.Error(
					r.Context(),
					fmt.Errorf("invalid room id '%s'", rId),
				)

				http.Error(
					w,
					http.StatusText(http.StatusBadRequest),
					http.StatusBadRequest,
				)

				return
			}

			c := auth.NewComparisonClaims(rId)
			if err := t.ValidateToken(r.Context(), r, c); err != nil {
				logger.Error(r.Context(), err)

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

func CorrelationIDFunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, _ := user.FromContext(r.Context())
		u.CorrelationId = uuid.New().String()
		ctx := user.WithContext(r.Context(), u)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func JSONContentTypeFunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func MethodTypeFunc(t string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != t {
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
