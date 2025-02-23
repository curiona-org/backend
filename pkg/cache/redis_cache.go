package cache

import (
	"context"
	"strings"
	"time"

	"github.com/curiona-org/backend/pkg/redis"
	baseredis "github.com/redis/go-redis/v9"
	"github.com/vmihailenco/msgpack/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type redisCache[V any] struct {
	conn   *redis.Client
	tracer trace.Tracer
}

var _ Cache[any] = (*redisCache[any])(nil)

func NewRedisCache[V any](conn *redis.Client) Cache[V] {
	tracer := otel.Tracer("cache:redis")
	return &redisCache[V]{
		conn:   conn,
		tracer: tracer,
	}
}

func (c *redisCache[V]) Read(ctx context.Context, k *Key, out *V) bool {
	key := k.String()

	ctx, span := spanWithOperationKey(ctx, c.tracer, "(*redisCache[V]).Read", "HGET", key)
	defer span.End()

	data, err := c.conn.HGet(ctx, key, "data").Result()
	if err != nil {
		span.RecordError(err, trace.WithAttributes(attribute.String("key", key)))
		return false
	}

	if data == "" {
		_, span = spanWithOperationKey(ctx, c.tracer, "(*redisCache[V]).Read", "DEL", key)
		defer span.End()
		c.conn.Del(ctx, key)
		return false
	}

	if err = msgpack.Unmarshal([]byte(data), &out); err != nil {
		span.SetStatus(codes.Error, "failed to unmarshal data for key: "+key)
		span.RecordError(err)
		return false
	}

	return true
}

func (c *redisCache[V]) List(ctx context.Context, k *Key) ([]V, bool) {
	ctx, span := spanWithOperationKey(ctx, c.tracer, "(*redisCache[V]).List", "SMEMBERS", k.Key)
	defer span.End()

	keys, err := c.conn.SMembers(ctx, k.String()).Result()
	if err != nil {
		span.SetStatus(codes.Error, "failed to read key: "+k.Key)
		span.RecordError(err)
		return nil, false
	}

	var values []V
	results, err := c.conn.Pipelined(ctx, func(pipe baseredis.Pipeliner) error {
		for _, key := range keys {
			pipe.HGet(ctx, key, "data")
		}
		return nil
	})
	if err != nil {
		span.SetStatus(codes.Error, "failed to read key: "+k.Key)
		span.RecordError(err)
		return nil, false
	}

	for _, result := range results {
		hget, ok := result.(*baseredis.StringCmd)
		if !ok {
			span.SetStatus(codes.Error, "failed to read key: "+k.Key)
			span.RecordError(err)
			continue
		}

		data, err := hget.Result()
		if err != nil {
			span.RecordError(err)
			continue
		}

		if data == "" {
			_, span = spanWithOperationKey(ctx, c.tracer, "(*redisCache[V]).List", "DEL", k.Key)
			defer span.End()
			c.conn.Del(ctx, k.Key)
			continue
		}

		var value V
		if err = msgpack.Unmarshal([]byte(data), &value); err != nil {
			span.SetStatus(codes.Error, "failed to unmarshal data for key: "+k.Key)
			span.RecordError(err)
			continue
		}

		values = append(values, value)
	}

	if len(values) == 0 {
		return nil, false
	}

	return values, true
}

func (c *redisCache[V]) Write(ctx context.Context, k *Key, value V, ttl time.Duration) {
	key := k.String()
	namespace := k.Namespace

	_, span := c.tracer.Start(ctx, "(*redisCache[V]).Write", trace.WithAttributes(attribute.String("key", key)))
	defer span.End()

	data, err := msgpack.Marshal(value)
	if err != nil {
		return
	}

	_, err = c.conn.Pipelined(ctx, func(pipe baseredis.Pipeliner) error {
		pipe.HSet(ctx, key, "data", data)
		if ttl > 0 {
			pipe.Expire(ctx, key, ttl)
		}

		pipe.SAdd(ctx, namespace, key)
		return nil
	})
	if err != nil {
		span.SetStatus(codes.Error, "failed to write key: "+key)
		span.RecordError(err)
	}
}

func (c *redisCache[V]) Exists(ctx context.Context, k *Key) bool {
	key := k.String()
	if key == "" {
		key = k.Namespace // check if the namespace exists
	}
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

func (c *redisCache[V]) Delete(ctx context.Context, ks ...*Key) error {
	var keys []string
	for _, k := range ks {
		keys = append(keys, k.String())
	}

	ctx, span := spanWithOperationKey(ctx, c.tracer, "(*redisCache[V]).Delete", "DEL", strings.Join(keys, ", "))
	defer span.End()

	if len(keys) == 0 {
		return nil
	}
	pipe := c.conn.Pipeline()
	pipe.Select(ctx, 0)
	pipe.Del(ctx, keys...)
	_, err := pipe.Exec(ctx)
	if err != nil {
		span.SetStatus(codes.Error, "failed to delete keys: "+strings.Join(keys, ", "))
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
