package cache

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type noopCache[V any] struct {
	tracer trace.Tracer
}

func NewNoopCache[V any]() Cache[V] {
	tracer := otel.Tracer("cache:noop")
	return &noopCache[V]{tracer: tracer}
}

func (c *noopCache[V]) Get(ctx context.Context, key string) (V, bool) {
	_, span := c.tracer.Start(ctx, "(*noopCache[V]).Get")
	defer span.End()
	_ = key
	return *new(V), false
}

func (c *noopCache[V]) GetArray(ctx context.Context, key string) ([]V, bool) {
	_, span := c.tracer.Start(ctx, "(*noopCache[V]).GetArray")
	defer span.End()
	_ = key
	return nil, false
}

func (c *noopCache[V]) List(ctx context.Context, key string) ([]V, bool) {
	_, span := c.tracer.Start(ctx, "(*noopCache[V]).List")
	defer span.End()
	_ = key
	return nil, false
}

func (c *noopCache[V]) Push(ctx context.Context, key string, value ...V) {
	_, span := c.tracer.Start(ctx, "(*noopCache[V]).Push")
	defer span.End()
	_ = key
	_ = value
}

func (c *noopCache[V]) Exists(ctx context.Context, key string) bool {
	_, span := c.tracer.Start(ctx, "(*noopCache[V]).Exists")
	defer span.End()
	_ = key
	return false
}

func (c *noopCache[V]) Set(ctx context.Context, key string, value ...V) {
	_, span := c.tracer.Start(ctx, "(*noopCache[V]).Set")
	defer span.End()
	_ = key
	_ = value
}

func (c *noopCache[V]) Delete(ctx context.Context, key ...string) error {
	_, span := c.tracer.Start(ctx, "(*noopCache[V]).Delete")
	defer span.End()
	_ = key
	return nil
}
func (c *noopCache[V]) Truncate(ctx context.Context) error {
	_, span := c.tracer.Start(ctx, "(*noopCache[V]).Truncate")
	defer span.End()
	return nil
}
