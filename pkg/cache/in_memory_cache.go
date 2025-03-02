package cache

import (
	"context"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"slices"

	"github.com/curiona-org/backend/internal/logger"
	"github.com/vmihailenco/msgpack/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var (
	inMemoryStore      = make(map[string][]byte)
	inMemoryIndexStore = make(map[string][]string)
)

type inMemoryEntry[V any] struct {
	Value     V
	TTL       time.Duration
	CreatedAt time.Time
}

type inMemoryCache[V any] struct {
	mtx    sync.RWMutex
	nSize  atomic.Int64
	tracer trace.Tracer
}

var _ Cache[any] = (*inMemoryCache[any])(nil)

func NewInMemoryCache[V any]() Cache[V] {
	tracer := otel.Tracer("cache:in_memory")
	cacher := &inMemoryCache[V]{
		tracer: tracer,
	}

	go cacher.runJanitor()
	return cacher
}

func (c *inMemoryCache[V]) Read(ctx context.Context, k *Key, out *V) bool {
	key := k.String()

	_, span := spanWithKey(ctx, c.tracer, "(*inMemoryCache[V]).Read", key)
	defer span.End()

	entry, ok := c.lookup(key)
	if !ok {
		return false
	}

	*out = entry.Value
	return true
}

func (c *inMemoryCache[V]) List(ctx context.Context, k *Key) ([]V, bool) {
	key := k.String()

	_, span := spanWithKey(ctx, c.tracer, "(*inMemoryCache[V]).List", key)
	defer span.End()

	entries, ok := c.lookupIndex(key)
	if !ok {
		return nil, false
	}

	values := make([]V, 0, len(entries))
	for _, entry := range entries {
		values = append(values, entry.Value)
	}

	return values, true
}

func (c *inMemoryCache[V]) Write(ctx context.Context, k *Key, value V, TTL time.Duration) {
	key := k.String()
	namespace := k.Namespace

	_, span := spanWithKey(ctx, c.tracer, "(*inMemoryCache[V]).Write", key)
	defer span.End()

	entry := inMemoryEntry[V]{
		Value:     value,
		TTL:       TTL,
		CreatedAt: time.Now(),
	}

	data, err := msgpack.Marshal(entry)
	if err != nil {
		return
	}

	c.mtx.Lock()
	inMemoryStore[key] = data
	inMemoryIndexStore[namespace] = append(inMemoryIndexStore[namespace], key)
	c.nSize.Add(1)
	c.mtx.Unlock()
}

func (c *inMemoryCache[V]) Exists(ctx context.Context, k *Key) bool {
	key := k.String()
	if key == "" {
		key = k.Namespace // check if the namespace exists
	}
	_, span := spanWithKey(ctx, c.tracer, "(*inMemoryCache[V]).Exists", key)
	defer span.End()

	c.mtx.RLock()
	defer c.mtx.RUnlock()

	_, ok := inMemoryStore[key]
	if ok {
		return true
	}

	_, ok = inMemoryIndexStore[key]
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

	_, span := spanWithKey(ctx, c.tracer, "(*inMemoryCache[V]).Delete", strings.Join(keys, ", "))
	defer span.End()

	c.mtx.Lock()
	defer c.mtx.Unlock()

	for _, key := range keys {
		if key == "*" {
			clear(inMemoryStore)
			clear(inMemoryIndexStore)
			return nil
		}

		_, ok := inMemoryStore[key]
		if !ok {
			// try deletion if the provided key uses a pattern
			for k := range inMemoryStore {
				if matched, _ := filepath.Match(key, k); matched {
					delete(inMemoryStore, k)
					c.nSize.Add(-1)
				}
			}

			for k := range inMemoryIndexStore {
				if matched, _ := filepath.Match(key, k); matched {
					delete(inMemoryIndexStore, k)
				}
			}
			continue
		}

		delete(inMemoryStore, key)
	}
	return nil
}

func (c *inMemoryCache[V]) Truncate(ctx context.Context) error {
	_, span := spanWithKey(ctx, c.tracer, "(*inMemoryCache[V]).Truncate", "")
	defer span.End()

	c.mtx.Lock()
	defer c.mtx.Unlock()
	clear(inMemoryStore)
	return nil
}

func (c *inMemoryCache[V]) lookup(key string) (inMemoryEntry[V], bool) {
	c.mtx.RLock()
	data, ok := inMemoryStore[key]
	if !ok {
		return inMemoryEntry[V]{}, false
	}
	c.mtx.RUnlock()

	var entry inMemoryEntry[V]
	if err := msgpack.Unmarshal(data, &entry); err != nil {
		return inMemoryEntry[V]{}, false
	}

	if entry.TTL > 0 && time.Since(entry.CreatedAt) > entry.TTL {
		c.mtx.Lock()
		delete(inMemoryStore, key)
		c.nSize.Add(-1)
		c.mtx.Unlock()
		return inMemoryEntry[V]{}, false
	}

	return entry, true
}

func (c *inMemoryCache[V]) lookupIndex(key string) ([]inMemoryEntry[V], bool) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	keys, ok := inMemoryIndexStore[key]
	if !ok {
		return nil, false
	}

	entries := make([]inMemoryEntry[V], 0, len(keys))
	for idx, key := range keys {
		entry, entryExists := c.lookup(key)
		if !entryExists {
			// key still exists in secondary index but not in the primary store
			// remove it from the secondary index array
			keys = slices.Delete(keys, idx, idx+1)
			inMemoryIndexStore[key] = keys
			continue
		}

		entries = append(entries, entry)
	}

	return entries, true
}

func (c *inMemoryCache[V]) runJanitor() {
	log := logger.Get()
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		c.mtx.Lock()
		for key, data := range inMemoryStore {
			var entry inMemoryEntry[V]
			if err := msgpack.Unmarshal(data, &entry); err != nil {
				log.Error().Err(err).Msg("failed to unmarshal data for key: " + key)
				continue
			}

			if entry.TTL > 0 && time.Since(entry.CreatedAt) > entry.TTL {
				delete(inMemoryStore, key)
				c.nSize.Add(-1)
			}
		}
		c.mtx.Unlock()
	}
}
