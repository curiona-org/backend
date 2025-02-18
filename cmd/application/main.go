package main

import (
	"context"

	_ "github.com/joho/godotenv/autoload"
	"github.com/roadmap-thesis/backend/internal/api"
	"github.com/roadmap-thesis/backend/internal/app"
	"github.com/roadmap-thesis/backend/internal/config"
	"github.com/roadmap-thesis/backend/internal/provider"
	"github.com/roadmap-thesis/backend/internal/repository"
	"github.com/roadmap-thesis/backend/pkg/auth"
	"github.com/roadmap-thesis/backend/pkg/auth/oauth"
	"github.com/roadmap-thesis/backend/pkg/logger"
	"github.com/rs/zerolog/log"
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
	postgresRepository := repository.NewPostgresRepository(provider.DB, provider.Redis)

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
	)

	api := api.New(config.Port(), app)

	log.Info().Msg("Starting Application Server...")
	api.Start(ctx)
}
