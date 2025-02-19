package cache

import (
	"context"

	"github.com/roadmap-thesis/backend/pkg/redis"
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

type Connection struct {
	Config *Config
	Redis  *redis.Client
}

func New[V any](conn *Connection) Cache[V] {
	if conn != nil && conn.Config.Type == TypeRedis && conn.Redis != nil {
		return NewRedisCache[V](conn.Redis)
	}

	return NewNoopCache[V]()
}

func NewConnection(ctx context.Context, cfg *Config) (*Connection, error) {
	conn := &Connection{Config: cfg}
	switch cfg.Type {
	case TypeRedis:
		rdb, err := redis.New(ctx, cfg.RedisConfig)
		if err != nil {
			return nil, err
		}
		conn.Redis = rdb
	}
	return conn, nil
}

func (c *Connection) Close() error {
	if c != nil && c.Config.Type == TypeRedis && c.Redis != nil {
		return c.Redis.Close()
	}

	return nil
}
