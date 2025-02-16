package main

import (
	"context"

	_ "github.com/joho/godotenv/autoload"
	"github.com/roadmap-thesis/backend/internal/api"
	"github.com/roadmap-thesis/backend/internal/application"
	"github.com/roadmap-thesis/backend/internal/clients"
	"github.com/roadmap-thesis/backend/internal/repository"
	"github.com/roadmap-thesis/backend/pkg/auth"
	"github.com/roadmap-thesis/backend/pkg/auth/oauth"
	"github.com/roadmap-thesis/backend/pkg/config"
	"github.com/roadmap-thesis/backend/pkg/logger"
	"github.com/rs/zerolog/log"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config.Init()
	logger.Init(config.IsDevelopment())

	clients, err := clients.New(ctx)
	if err != nil {
		//nolint:gocritic
		log.Fatal().Err(err).Msg("Failed to initialize clients")
	}
	defer clients.Close(ctx)

	log.Info().Msg("Bootstrapping application...")
	postgresRepository := repository.NewPostgresRepository(clients.DB, clients.Redis)

	application := application.New(
		postgresRepository,
		clients.LLM,
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

	api := api.New(config.Port(), application)

	log.Info().Msg("Starting Application Server...")
	api.Start(ctx)
}
