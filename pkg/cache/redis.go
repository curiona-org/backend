package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/vmihailenco/msgpack/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type redisCache[V any] struct {
	conn   *redis.Client
	tracer trace.Tracer
}

func NewRedisCache[V any](conn Connection) Cache[V] {
	tracer := otel.Tracer("cache:redis")
	if _, ok := conn.(*redis.Client); !ok {
		return NewNoopCache[V]() // temporary
	}

	return &redisCache[V]{
		conn:   conn.(*redis.Client),
		tracer: tracer,
	}
}

func (c *redisCache[V]) Get(ctx context.Context, key string) (V, bool) {
	ctx, span := c.tracer.Start(ctx, "(*redisCache[V]).Get")
	defer span.End()

	var value V
	data, err := c.conn.Get(ctx, key).Result()
	if err != nil {
		return value, false
	}

	if err := msgpack.Unmarshal([]byte(data), &value); err != nil {
		return value, false
	}

	return value, true
}
func (c *redisCache[V]) Set(ctx context.Context, key string, value V) {
	ctx, span := c.tracer.Start(ctx, "(*redisCache[V]).Set")
	defer span.End()

	data, err := msgpack.Marshal(value)
	if err != nil {
		return
	}

	c.conn.Set(ctx, key, data, 0)
}

func (c *redisCache[V]) Delete(ctx context.Context, key ...string) error {
	ctx, span := c.tracer.Start(ctx, "(*redisCache[V]).Delete")
	defer span.End()

	if len(key) == 0 {
		return nil
	}
	pipe := c.conn.Pipeline()
	pipe.Select(ctx, 0)
	pipe.Del(ctx, key...)
	_, err := pipe.Exec(ctx)
	return err
}
func (c *redisCache[V]) Truncate(ctx context.Context) error {
	ctx, span := c.tracer.Start(ctx, "(*redisCache[V]).Truncate")
	defer span.End()

	pipe := c.conn.Pipeline()
	pipe.Select(ctx, 0)
	pipe.FlushDB(ctx)
	_, err := pipe.Exec(ctx)
	return err
}
