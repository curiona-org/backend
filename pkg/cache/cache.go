package cache

import (
	"context"
)

type Cache[V any] interface {
	Get(ctx context.Context, key string) (V, bool)
	Set(ctx context.Context, key string, value V)
	Delete(ctx context.Context, key ...string) error
	Truncate(ctx context.Context) error
}

type Connection interface{}
