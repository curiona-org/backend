package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/curiona-org/backend/internal/admin"
	"github.com/curiona-org/backend/internal/api"
	"github.com/curiona-org/backend/internal/app"
	"github.com/curiona-org/backend/internal/auth"
	"github.com/curiona-org/backend/internal/auth/oauth"
	"github.com/curiona-org/backend/internal/chat"
	"github.com/curiona-org/backend/internal/config"
	"github.com/curiona-org/backend/internal/logger"
	"github.com/curiona-org/backend/internal/provider"
	"github.com/curiona-org/backend/internal/provider/option"
	"github.com/curiona-org/backend/internal/repository"
	"github.com/curiona-org/backend/internal/worker"
	"github.com/curiona-org/backend/pkg/cache"
	_ "github.com/joho/godotenv/autoload"
	"golang.org/x/sync/errgroup"
)

func run(ctx context.Context) {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	defer cancel()

	config.Init()
	log := logger.Get()

	var cacheType cache.Type
	if config.IsDevelopment() {
		cacheType = cache.TypeInMemory
	} else {
		cacheType = cache.TypeRedis
	}

	withQueueOpt := provider.Option(nil)
	if cacheType == cache.TypeRedis {
		withQueueOpt = option.WithQueue()
	}

	provider, err := provider.New(
		option.WithLLM(),
		option.WithPostgresDB(ctx),
		option.WithCache(ctx, cacheType),
		withQueueOpt,
		option.WithYoutubeClient(),
		option.WithGoogleBooksClient(),
		option.WithTracing(ctx),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize provider") //nolint:gocritic
	}
	defer provider.Close(ctx)

	log.Info().Msg("Bootstrapping application...")
	postgresRepository := repository.NewPostgresRepository(provider.DB, provider.Cache)

	var w worker.Worker
	if config.IsDevelopment() {
		w = worker.NewNoop()
	} else {
		w = worker.NewAsynq(
			provider.Queue,
			provider.QueueServer,
			postgresRepository,
			provider.GoogleBooks,
			provider.Youtube,
		)
	}

	auth := auth.New(
		&auth.Config{
			AccessSecretKey:  config.AccessSecretKey(),
			AccessExpiresIn:  config.AccessExpiresIn(),
			RefreshSecretKey: config.RefreshSecretKey(),
			RefreshExpiresIn: config.RefreshExpiresIn(),
		},
	)

	curionaApp := app.New(
		w,
		postgresRepository,
		provider.LLM,
		auth,
		oauth.NewGoogleProvider(
			config.GoogleClientID(),
			config.GoogleClientSecret(),
		),
		provider.GoogleBooks,
		provider.Youtube,
		provider.Tracing,
	)
	adminApp := admin.New(postgresRepository, auth, provider.Tracing)
	chatApp := chat.New(w, postgresRepository, provider.LLM, auth, provider.Tracing)

	api := api.New(
		ctx,
		config.Port(),
		curionaApp,
		adminApp,
		chatApp,
		provider.Tracing,
	)

	log.Info().Msg("Starting Application Server...")

	ctx = log.WithContext(ctx)
	group, groupCtx := errgroup.WithContext(ctx)

	if cacheType == cache.TypeRedis {
		group.Go(func() error {
			return w.Start(groupCtx)
		})
	}

	group.Go(func() error {
		api.Start(groupCtx)
		return nil
	})

	if err := group.Wait(); err != nil {
		log.Fatal().Err(err).Msg("Encountered an error while running the application")
	}

	log.Info().Msg("Application shutdown")
}

func main() {
	ctx := context.Background()
	run(ctx)
}
