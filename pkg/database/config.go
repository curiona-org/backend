package database

import (
	"time"
)

type Config struct {
	Name                  string
	Host                  string
	Port                  int
	User                  string
	Password              string
	ConnectionTimeout     int
	PoolMaxConnections    int
	PoolMinConnections    int
	PoolMaxConnLifetime   time.Duration
	PoolMaxConnIdleTime   time.Duration
	PoolHealthCheckPeriod time.Duration
}
