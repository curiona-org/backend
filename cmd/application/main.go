package main

import (
	"context"

	_ "github.com/joho/godotenv/autoload"
	"github.com/roadmap-thesis/backend/internal/api"
	"github.com/roadmap-thesis/backend/internal/app"
	"github.com/roadmap-thesis/backend/internal/auth"
	"github.com/roadmap-thesis/backend/internal/auth/oauth"
	"github.com/roadmap-thesis/backend/internal/config"
	"github.com/roadmap-thesis/backend/internal/logger"
	"github.com/roadmap-thesis/backend/internal/provider"
	"github.com/roadmap-thesis/backend/internal/repository"
	"github.com/roadmap-thesis/backend/internal/worker"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config.Init()
	logger.Init(config.IsDevelopment())

	provider, err := provider.New(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize provider") //nolint:gocritic
	}
	defer provider.Close(ctx)

	log.Info().Msg("Bootstrapping application...")
	postgresRepository := repository.NewPostgresRepository(provider.DB, provider.Cache)

	app := app.New(
		postgresRepository,
		provider.LLM,
		auth.New(
			auth.StrategyJWT,
			&auth.Config{
				AccessSecretKey:  config.AccessSecretKey(),
				AccessExpiresIn:  config.AccessExpiresIn(),
				RefreshSecretKey: config.RefreshSecretKey(),
				RefreshExpiresIn: config.RefreshExpiresIn(),
			},
		),
		oauth.NewGoogleProvider(
			config.GoogleClientID(),
			config.GoogleClientSecret(),
		),
		provider.GoogleBooks,
		provider.Youtube,
	)

	api := api.New(config.Port(), app)

	worker := worker.New(
		provider.Queue,
		provider.QueueServer,
		postgresRepository,
		provider.GoogleBooks,
		provider.Youtube,
	)

	log.Info().Msg("Starting Application Server...")

	group, groupCtx := errgroup.WithContext(ctx)
	group.SetLimit(2)

	group.Go(func() error {
		return worker.Start(groupCtx)
	})

	group.Go(func() error {
		api.Start(groupCtx)
		return nil
	})

	if err := group.Wait(); err != nil {
		log.Fatal().Err(err).Msg("Encountered an error while running the application")
	}

	log.Info().Msg("Application shutdown")
}
