package cache

import "github.com/curiona-org/backend/pkg/redis"

type Type string

const (
	TypeNoop  Type = "noop"
	TypeRedis Type = "redis"
)

type Config struct {
	Type        Type
	RedisConfig *redis.Config
}
