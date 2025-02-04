package main

import (
	"context"

	_ "github.com/joho/godotenv/autoload"
	"github.com/roadmap-thesis/backend/internal/api"
	"github.com/roadmap-thesis/backend/internal/backend"
	"github.com/roadmap-thesis/backend/internal/clients"
	"github.com/roadmap-thesis/backend/internal/repository"
	"github.com/roadmap-thesis/backend/pkg/config"
	"github.com/roadmap-thesis/backend/pkg/logger"
	"github.com/roadmap-thesis/backend/pkg/tracing"
	"github.com/rs/zerolog/log"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config.Init()
	logger.Init()

	clients, err := clients.New(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize clients")
	}
	defer clients.Close()

	trace, err := tracing.NewProvider(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize tracer provider")
	}
	defer func() {
		if err := trace.Shutdown(ctx); err != nil {
			log.Fatal().Err(err).Msg("Failed shutting down tracer provider")
		}
	}()

	log.Info().Msg("Bootstrapping application...")
	repository := repository.New(clients.DB)

	backend := backend.New(
		repository,
		clients.LLM,
		clients.Google,
	)

	api := api.New(config.Port(), backend)

	log.Info().Msg("Starting Application Server...")
	api.Start(ctx)
}
