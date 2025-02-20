package main

import (
	"context"
	"flag"

	"github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
	"github.com/pressly/goose/v3"
	"github.com/roadmap-thesis/backend/internal/config"
	"github.com/roadmap-thesis/backend/internal/database"
	"github.com/roadmap-thesis/backend/internal/logger"
	"github.com/rs/zerolog/log"
)

func main() {
	command := flag.String("command", "up", "migration command (up/down/reset)")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config.Init()
	logger.Init(config.IsDevelopment())

	postgresConn, err := database.New(ctx, &database.Config{
		Name:                  config.DBName(),
		Host:                  config.DBHost(),
		Port:                  config.DBPort(),
		User:                  config.DBUser(),
		Password:              config.DBPassword(),
		ConnectionTimeout:     config.DBConnectionTimeout(),
		PoolMaxConnections:    config.DBPoolMaxConnections(),
		PoolMinConnections:    config.DBPoolMinConnections(),
		PoolMaxConnLifetime:   config.DBPoolMaxConnLifetime(),
		PoolMaxConnIdleTime:   config.DBPoolMaxConnIdleTime(),
		PoolHealthCheckPeriod: config.DBPoolHealthCheckPeriod(),
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize provider") //nolint:gocritic
	}
	defer postgresConn.Close()

	if err = goose.SetDialect("pgx"); err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize goose")
	}

	goose.SetTableName("schema_migrations")

	db := stdlib.OpenDBFromPool(postgresConn.Pool())
	defer db.Close()

	dir := "./migrations"
	switch *command {
	case "up":
		err = goose.Up(db, dir)
	case "down":
		err = goose.Down(db, dir)
	case "reset":
		err = goose.Reset(db, dir)
	default:
		log.Fatal().Err(err).Msg("Unknown migration command")
	}

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to run migrations")
	}

	log.Info().Msg("Migrations applied successfully!")
}
