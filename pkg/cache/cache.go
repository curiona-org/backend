package cache

import (
	"context"
)

type Cache[V any] interface {
	Get(ctx context.Context, key string) (V, bool)
	GetArray(ctx context.Context, key string) ([]V, bool)
	List(ctx context.Context, key string) ([]V, bool)
	Push(ctx context.Context, key string, value V)
	Exists(ctx context.Context, key string) bool
	Set(ctx context.Context, key string, value ...V)
	Delete(ctx context.Context, key ...string) error
	Truncate(ctx context.Context) error
}

type Connection interface{}
