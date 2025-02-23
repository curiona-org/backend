package cache

import (
	"context"
	"time"
)

type noopCache[V any] struct {
}

var _ Cache[any] = (*noopCache[any])(nil)

func NewNoopCache[V any]() Cache[V] {
	return &noopCache[V]{}
}

func (c *noopCache[V]) Read(ctx context.Context, k *Key, out *V) bool {
	_ = k
	_ = out
	return false
}

func (c *noopCache[V]) List(ctx context.Context, k *Key) ([]V, bool) {
	_ = k
	return nil, false
}

func (c *noopCache[V]) Write(ctx context.Context, k *Key, value V, ttl time.Duration) {
	_ = k
	_ = value
	_ = ttl
}

func (c *noopCache[V]) Exists(ctx context.Context, k *Key) bool {
	_ = k
	return false
}

func (c *noopCache[V]) Delete(ctx context.Context, k ...*Key) error {
	_ = k
	return nil
}

func (c *noopCache[V]) Truncate(ctx context.Context) error {
	return nil
}
