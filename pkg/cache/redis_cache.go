package cache

import (
	"context"
	"strings"

	"github.com/redis/go-redis/v9"
	"github.com/vmihailenco/msgpack/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type redisCache[V any] struct {
	conn   *redis.Client
	tracer trace.Tracer
}

func NewRedisCache[V any](conn Connection) Cache[V] {
	tracer := otel.Tracer("cache:redis")
	cacheConn, ok := conn.(*redis.Client)
	if !ok {
		return NewNoopCache[V]() // temporary
	}

	return &redisCache[V]{
		conn:   cacheConn,
		tracer: tracer,
	}
}

func (c *redisCache[V]) Get(ctx context.Context, key string) (V, bool) {
	ctx, span := spanWithOperationKey(ctx, c.tracer, "(*redisCache[V]).Get", "GET", key)
	defer span.End()

	var value V
	data, err := c.conn.Get(ctx, key).Result()
	if err != nil {
		span.SetStatus(codes.Error, "failed to get key: "+key)
		span.RecordError(err)
		return value, false
	}

	if err = msgpack.Unmarshal([]byte(data), &value); err != nil {
		return value, false
	}

	return value, true
}

// GetArray returns an array stored in a single key.
func (c *redisCache[V]) GetArray(ctx context.Context, key string) ([]V, bool) {
	ctx, span := spanWithOperationKey(ctx, c.tracer, "(*redisCache[V]).GetArray", "GET", key)
	defer span.End()

	var values []V
	data, err := c.conn.Get(ctx, key).Result()
	if err != nil {
		span.SetStatus(codes.Error, "failed to get key: "+key)
		span.RecordError(err)
		return values, false
	}

	if err = msgpack.Unmarshal([]byte(data), &values); err != nil {
		return values, false
	}

	return values, true
}

// List returns a list of values stored using redis list.
func (c *redisCache[V]) List(ctx context.Context, key string) ([]V, bool) {
	ctx, span := spanWithOperationKey(ctx, c.tracer, "(*redisCache[V]).List", "LRANGE", key)
	defer span.End()

	var values []V
	data, err := c.conn.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		span.SetStatus(codes.Error, "failed to get key: "+key)
		span.RecordError(err)
		return values, false
	}

	for _, d := range data {
		var value V
		if err = msgpack.Unmarshal([]byte(d), &value); err != nil {
			span.SetStatus(codes.Error, "failed to unmarshal value")
			span.RecordError(err)
			return values, false
		}
		values = append(values, value)
	}

	return values, true
}

func (c *redisCache[V]) Push(ctx context.Context, key string, value V) {
	ctx, span := spanWithOperationKey(ctx, c.tracer, "(*redisCache[V]).Push", "RPUSH", key)
	defer span.End()

	data, err := msgpack.Marshal(value)
	if err != nil {
		return
	}

	if err = c.conn.RPush(ctx, key, data).Err(); err != nil {
		span.SetStatus(codes.Error, "failed to add key: "+key)
		span.RecordError(err)
	}
}

func (c *redisCache[V]) Exists(ctx context.Context, key string) bool {
	ctx, span := spanWithOperationKey(ctx, c.tracer, "(*redisCache[V]).Exists", "EXISTS", key)
	defer span.End()

	ok, err := c.conn.Exists(ctx, key).Result()
	if err != nil {
		span.SetStatus(codes.Error, "failed to check if key exists: "+key)
		span.RecordError(err)
		return false
	}

	return ok == 1
}

func (c *redisCache[V]) Set(ctx context.Context, key string, value ...V) {
	ctx, span := spanWithOperationKey(ctx, c.tracer, "(*redisCache[V]).Set", "SET", key)
	defer span.End()

	data, err := msgpack.Marshal(value)
	if err != nil {
		return
	}

	if err = c.conn.Set(ctx, key, data, 0).Err(); err != nil {
		span.SetStatus(codes.Error, "failed to set key: "+key)
		span.RecordError(err)
	}
}

func (c *redisCache[V]) Delete(ctx context.Context, key ...string) error {
	ctx, span := spanWithOperationKey(ctx, c.tracer, "(*redisCache[V]).Delete", "DEL", strings.Join(key, ", "))
	defer span.End()

	if len(key) == 0 {
		return nil
	}
	pipe := c.conn.Pipeline()
	pipe.Select(ctx, 0)
	pipe.Del(ctx, key...)
	_, err := pipe.Exec(ctx)
	if err != nil {
		span.SetStatus(codes.Error, "failed to delete keys: "+strings.Join(key, ", "))
		span.RecordError(err)
	}
	return nil
}
func (c *redisCache[V]) Truncate(ctx context.Context) error {
	ctx, span := spanWithOperationKey(ctx, c.tracer, "(*redisCache[V]).Truncate", "FLUSHDB", "")
	defer span.End()

	pipe := c.conn.Pipeline()
	pipe.Select(ctx, 0)
	pipe.FlushDB(ctx)
	_, err := pipe.Exec(ctx)
	if err != nil {
		span.SetStatus(codes.Error, "failed to truncate cache")
		span.RecordError(err)
	}
	return nil
}
