package main

import (
	"context"

	_ "github.com/joho/godotenv/autoload"
	"github.com/roadmap-thesis/backend/internal/api"
	"github.com/roadmap-thesis/backend/internal/backend"
	"github.com/roadmap-thesis/backend/internal/clients"
	"github.com/roadmap-thesis/backend/internal/repository"
	"github.com/roadmap-thesis/backend/pkg/auth"
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
		log.Fatal().Err(err).Msg("Failed to initialize clients")
	}
	defer clients.Close(ctx)

	log.Info().Msg("Bootstrapping application...")
	repository := repository.New(clients.DB)

	backend := backend.New(
		repository,
		clients.LLM,
		auth.NewJWT(
			config.JWTSecretKey(),
			config.JWTExpiresIn()),
		clients.Google,
	)

	api := api.New(config.Port(), backend)

	log.Info().Msg("Starting Application Server...")
	api.Start(ctx)
}
