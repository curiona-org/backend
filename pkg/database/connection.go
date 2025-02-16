package database

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	pool *pgxpool.Pool
}

type Connection interface {
	Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, query string, args ...any) pgx.Row
	Query(context.Context, string, ...any) (pgx.Rows, error)

	InTx(ctx context.Context, fn TransactionFunc) error

	// Pool returns the underlying pgxpool.Pool for direct access to the connection pool.
	Pool() *pgxpool.Pool
	Close()
}

// New creates a new PostgreSQL pgxpool client.
func New(ctx context.Context, cfg *Config) (Connection, error) {
	pgxConfig, err := pgxpool.ParseConfig(dbDSN(cfg))
	if err != nil {
		return nil, err
	}

	pgxConfig.BeforeAcquire = func(ctx context.Context, conn *pgx.Conn) bool {
		return conn.Ping(ctx) == nil
	}

	pool, err := pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		return nil, err
	}

	return &DB{pool: pool}, nil
}

// dbDSN builds a connection string suitable for the pgx Postgres driver, using
// the values of vars.
func dbDSN(cfg *Config) string {
	vals := dbValues(cfg)
	p := make([]string, 0, len(vals))
	for k, v := range vals {
		p = append(p, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(p, " ")
}

func setIfNotEmpty(m map[string]string, key string, val string) {
	if val != "" {
		m[key] = val
	}
}

func setIfPositive(m map[string]string, key string, val int) {
	if val > 0 {
		m[key] = strconv.Itoa(val)
	}
}

func setIfPositiveDuration(m map[string]string, key string, d time.Duration) {
	if d > 0 {
		m[key] = d.String()
	}
}

func dbValues(cfg *Config) map[string]string {
	p := map[string]string{}
	setIfNotEmpty(p, "dbname", cfg.Name)
	setIfNotEmpty(p, "user", cfg.User)
	setIfNotEmpty(p, "host", cfg.Host)
	setIfPositive(p, "port", cfg.Port)
	setIfPositive(p, "connect_timeout", cfg.ConnectionTimeout)
	setIfNotEmpty(p, "password", cfg.Password)
	setIfPositive(p, "pool_min_conns", cfg.PoolMinConnections)
	setIfPositive(p, "pool_max_conns", cfg.PoolMaxConnections)
	setIfPositiveDuration(p, "pool_max_conn_lifetime", cfg.PoolMaxConnLifetime)
	setIfPositiveDuration(p, "pool_max_conn_idle_time", cfg.PoolMaxConnIdleTime)
	setIfPositiveDuration(p, "pool_health_check_period", cfg.PoolHealthCheckPeriod)
	return p
}

func (db *DB) Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	return db.pool.Exec(ctx, query, args...)
}

func (db *DB) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	return db.pool.QueryRow(ctx, query, args...)
}

func (db *DB) Query(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	return db.pool.Query(ctx, query, args...)
}

func (db *DB) Pool() *pgxpool.Pool {
	return db.pool
}

func (db *DB) Close() {
	db.pool.Close()
}
