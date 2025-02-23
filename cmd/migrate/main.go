package main

import (
	"context"
	"flag"

	"github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
	"github.com/pressly/goose/v3"
	"github.com/roadmap-thesis/backend/internal/config"
	"github.com/roadmap-thesis/backend/internal/logger"
	"github.com/roadmap-thesis/backend/internal/provider"
	"github.com/roadmap-thesis/backend/internal/provider/option"
	"github.com/rs/zerolog/log"
)

func main() {
	command := flag.String("command", "up", "migration command (up/down/reset)")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config.Init()
	logger.Init()

	provider, err := provider.New(option.WithPostgresDB(ctx))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize provider") //nolint:gocritic
	}
	defer provider.Close(ctx)

	if err = goose.SetDialect("pgx"); err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize goose")
	}

	goose.SetTableName("schema_migrations")

	db := stdlib.OpenDBFromPool(provider.DB.Pool())
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
