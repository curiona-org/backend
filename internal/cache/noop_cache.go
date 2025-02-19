package cache

import (
	"context"
)

type noopCache[V any] struct {
}

func NewNoopCache[V any]() Cache[V] {
	return &noopCache[V]{}
}

func (c *noopCache[V]) Get(ctx context.Context, key string) (V, bool) {
	_ = key
	return *new(V), false
}

func (c *noopCache[V]) GetArray(ctx context.Context, key string) ([]V, bool) {
	_ = key
	return nil, false
}

func (c *noopCache[V]) List(ctx context.Context, key string) ([]V, bool) {
	_ = key
	return nil, false
}

func (c *noopCache[V]) Push(ctx context.Context, key string, value V) {
	_ = key
	_ = value
}

func (c *noopCache[V]) Exists(ctx context.Context, key string) bool {
	_ = key
	return false
}

func (c *noopCache[V]) Set(ctx context.Context, key string, value ...V) {
	_ = key
	_ = value
}

func (c *noopCache[V]) Delete(ctx context.Context, key ...string) error {
	_ = key
	return nil
}
func (c *noopCache[V]) Truncate(ctx context.Context) error {
	return nil
}
