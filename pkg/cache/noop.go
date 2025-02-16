package cache

import "context"

type noopCache[V any] struct{}

func NewNoopCache[V any]() Cache[V] {
	return &noopCache[V]{}
}

func (*noopCache[V]) Get(ctx context.Context, key string) (value V, ok bool) {
	return
}

func (*noopCache[V]) Set(ctx context.Context, key string, value V) {}

func (*noopCache[V]) Delete(ctx context.Context, key ...string) error {
	return nil
}
func (*noopCache[V]) Truncate(ctx context.Context) error {
	return nil
}
