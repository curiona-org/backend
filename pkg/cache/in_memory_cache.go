package cache

import (
	"context"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/curiona-org/backend/internal/logger"
	"github.com/vmihailenco/msgpack/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type inMemoryCache[V any] struct {
	mtx    sync.RWMutex
	cache  map[string][]byte
	index  map[string][]string
	nSize  atomic.Int64
	tracer trace.Tracer
}

var _ Cache[any] = (*inMemoryCache[any])(nil)

func NewInMemoryCache[V any]() Cache[V] {
	tracer := otel.Tracer("cache:in_memory")
	cacher := &inMemoryCache[V]{
		cache:  make(map[string][]byte),
		index:  make(map[string][]string),
		tracer: tracer,
	}

	go cacher.runJanitor()
	return cacher
}

func (c *inMemoryCache[V]) Read(ctx context.Context, k *Key, out *V) bool {
	key := k.String()

	ctx, span := spanWithKey(ctx, c.tracer, "(*inMemoryCache[V]).Read", key)
	defer span.End()

	c.mtx.RLock()
	data, ok := c.cache[key]
	c.mtx.RUnlock()

	if !ok {
		return false
	}

	var entry Entry[V]
	if err := msgpack.Unmarshal([]byte(data), &entry); err != nil {
		span.SetStatus(codes.Error, "failed to unmarshal data for key: "+key)
		span.RecordError(err)
		return false
	}

	if entry.ttl > 0 && time.Since(entry.createdAt) > entry.ttl {
		c.mtx.Lock()
		delete(c.cache, key)
		c.nSize.Add(-1)
		c.mtx.Unlock()
		return false
	}

	out = &entry.value
	return ok
}

func (c *inMemoryCache[V]) List(ctx context.Context, k *Key) ([]V, bool) {
	key := k.String()

	ctx, span := spanWithKey(ctx, c.tracer, "(*inMemoryCache[V]).List", key)
	defer span.End()

	c.mtx.RLock()
	keys, ok := c.index[key]
	c.mtx.RUnlock()

	if !ok {
		return nil, false
	}

	values := make([]V, 0, len(keys))
	c.mtx.Lock()
	for _, key := range keys {
		data, ok := c.cache[key]
		if !ok {
			continue
		}

		var entry Entry[V]
		if err := msgpack.Unmarshal([]byte(data), &entry); err != nil {
			span.SetStatus(codes.Error, "failed to unmarshal data for key: "+key)
			span.RecordError(err)
			continue
		}

		if entry.ttl > 0 && time.Since(entry.createdAt) > entry.ttl {
			c.mtx.Lock()
			delete(c.cache, key)
			c.nSize.Add(-1)
			c.mtx.Unlock()
			continue
		}

		values = append(values, entry.value)
	}
	c.mtx.Unlock()

	return values, true
}

func (c *inMemoryCache[V]) Write(ctx context.Context, k *Key, value V, ttl time.Duration) {
	key := k.String()

	ctx, span := spanWithKey(ctx, c.tracer, "(*inMemoryCache[V]).Write", key)
	defer span.End()

	entry := Entry[V]{
		value:     value,
		ttl:       ttl,
		createdAt: time.Now(),
	}

	data, err := msgpack.Marshal(entry)
	if err != nil {
		return
	}

	c.mtx.Lock()
	c.cache[key] = data
	c.nSize.Add(1)
	c.mtx.Unlock()
}

func (c *inMemoryCache[V]) Exists(ctx context.Context, k *Key) bool {
	key := k.String()
	ctx, span := spanWithKey(ctx, c.tracer, "(*inMemoryCache[V]).Exists", key)
	defer span.End()

	c.mtx.RLock()
	defer c.mtx.RUnlock()

	_, ok := c.cache[key]
	return ok
}

func (c *inMemoryCache[V]) Delete(ctx context.Context, ks ...*Key) error {
	if len(ks) == 0 {
		return nil
	}
	var keys []string
	for _, k := range ks {
		keys = append(keys, k.String())
	}

	ctx, span := spanWithKey(ctx, c.tracer, "(*inMemoryCache[V]).Delete", strings.Join(keys, ", "))
	defer span.End()

	c.mtx.Lock()
	defer c.mtx.Unlock()

	for _, key := range keys {
		if key == "*" {
			clear(c.cache)
			return nil
		}

		_, ok := c.cache[key]
		if !ok {
			// try deletion if the provided key uses a pattern

			// check if it's a valid pattern first
			matched, err := filepath.Match(key, "")
			if !matched || err != nil {
				continue
			}

			delete(c.cache, key)
			c.nSize.Add(-1)
		}
	}
	return nil
}

func (c *inMemoryCache[V]) Truncate(ctx context.Context) error {
	ctx, span := spanWithKey(ctx, c.tracer, "(*inMemoryCache[V]).Truncate", "")
	defer span.End()

	c.mtx.Lock()
	defer c.mtx.Unlock()
	clear(c.cache)
	return nil
}

func (c *inMemoryCache[V]) runJanitor() {
	log := logger.Get()
	ticker := time.NewTicker(time.Minute)
	for {
		select {
		case <-ticker.C:
			c.mtx.Lock()
			for key, data := range c.cache {
				var entry Entry[V]
				if err := msgpack.Unmarshal([]byte(data), &entry); err != nil {
					log.Error().Err(err).Msg("failed to unmarshal data for key: " + key)
					continue
				}

				if entry.ttl > 0 && time.Since(entry.createdAt) > entry.ttl {
					delete(c.cache, key)
					c.nSize.Add(-1)
				}
			}
			c.mtx.Unlock()
		}
	}
}
