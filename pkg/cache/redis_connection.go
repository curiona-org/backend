package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	DB       int
	Network  string
	Addr     string
	Username string
	Password string
}

func NewRedisConnection(ctx context.Context, cfg *RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		DB:       cfg.DB,
		Network:  cfg.Network,
		Addr:     cfg.Addr,
		Username: cfg.Username,
		Password: cfg.Password,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return rdb, nil
}
