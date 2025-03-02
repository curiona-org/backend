package cache

import (
	"context"
	"errors"
	"time"

	"github.com/curiona-org/backend/pkg/redis"
)

const (
	DefaultTTL = 5 * time.Minute
)

type Key struct {
	Namespace string
	Key       string
}

func (k *Key) String() string {
	key := k.Key

	if k.Namespace != "" && k.Key != "" {
		key = k.Namespace + ":" + key
	}

	return key
}

type Cache[V any] interface {
	Read(ctx context.Context, k *Key, out *V) bool
	List(ctx context.Context, k *Key) ([]V, bool)
	Write(ctx context.Context, k *Key, value V, ttl time.Duration)
	Exists(ctx context.Context, k *Key) bool
	Delete(ctx context.Context, k ...*Key) error
	Truncate(ctx context.Context) error
}

type Entry[V any] struct {
	value     V
	ttl       time.Duration
	createdAt time.Time
}

type Connection struct {
	Config *Config
	Redis  *redis.Client
}

func New[V any](conn *Connection) Cache[V] {
	if conn != nil && conn.Config.Type == TypeRedis && conn.Redis != nil {
		return NewRedisCache[V](conn.Redis)
	}

	if conn.Config.Type == TypeInMemory {
		return NewInMemoryCache[V]()
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
	case TypeNoop, TypeInMemory:
		// noop
	default:
		return nil, errors.New("invalid cache type")
	}
	return conn, nil
}

func (c *Connection) Close() error {
	if c != nil && c.Config.Type == TypeRedis && c.Redis != nil {
		return c.Redis.Close()
	}

	return nil
}
