package user

import (
	"context"
)

type key string

const uKey key = "user"

type U struct {
	CorrelationId string
	RoomId        string
}

func FromContext(ctx context.Context) (U, bool) {
	u, ok := ctx.Value(uKey).(U)
	return u, ok
}

func WithContext(ctx context.Context, u U) context.Context {
	return context.WithValue(ctx, uKey, u)
}
