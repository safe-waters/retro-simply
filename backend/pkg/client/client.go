package client

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type PubSubChannel interface {
	Receive(ctx context.Context) (interface{}, error)
	Channel() <-chan *redis.Message
	Close() error
}

type Err interface {
	Err() error
}

type StrResult interface {
	Result() (string, error)
	Err
}

type BoolResult interface {
	Result() (bool, error)
	Err
}

type C struct {
	*redis.Client
}

func New(url string, poolSize int) (*C, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}

	opts.PoolSize = poolSize

	return &C{redis.NewClient(opts)}, nil
}

func (c *C) Subscribe(ctx context.Context, channels ...string) PubSubChannel {
	return c.Client.Subscribe(ctx, channels...)
}

func (c *C) Publish(
	ctx context.Context,
	channel string,
	message interface{},
) Err {
	return c.Client.Publish(ctx, channel, message)
}

func (c *C) Get(ctx context.Context, key string) StrResult {
	return c.Client.Get(ctx, key)
}

func (c *C) Watch(
	ctx context.Context,
	fn func(*redis.Tx) error,
	keys ...string,
) error {
	return c.Client.Watch(ctx, fn, keys...)
}

func (c *C) SetNX(
	ctx context.Context,
	key string,
	value interface{},
	expiration time.Duration,
) BoolResult {
	return c.Client.SetNX(ctx, key, value, expiration)
}
