package provider

import (
	"context"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/roadmap-thesis/backend/internal/config"
	"github.com/roadmap-thesis/backend/pkg/book"
	"github.com/roadmap-thesis/backend/pkg/cache"
	"github.com/roadmap-thesis/backend/pkg/database"
	"github.com/roadmap-thesis/backend/pkg/llm"
	"github.com/roadmap-thesis/backend/pkg/tracing"
	"github.com/roadmap-thesis/backend/pkg/youtube"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/sdk/trace"
	"golang.org/x/sync/errgroup"
)

type Provider struct {
	LLM         llm.Client
	DB          database.Connection
	Redis       *redis.Client
	Tracing     *trace.TracerProvider
	GoogleBooks book.Client
	Youtube     youtube.Client
}

func New(ctx context.Context) (*Provider, error) {
	p := &Provider{}

	var group errgroup.Group

	group.Go(func() error {
		log.Info().Msg("initializing llm client")
		var err error
		p.LLM, err = llm.NewClient(
			config.LLMProvider(),
			config.LLMAPIKey(),
			config.LLMModel())
		if err != nil {
			return errors.Wrap(err, "initializing llm client")
		}
		return nil
	})

	group.Go(func() error {
		log.Info().Msg("initializing postgresql")
		var err error
		p.DB, err = database.New(ctx, &database.Config{
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
			return errors.Wrap(err, "initializing postgresql")
		}
		return nil
	})

	group.Go(func() error {
		log.Info().Msg("initializing redis client")
		var err error
		p.Redis, err = cache.NewRedisConnection(ctx, &cache.RedisConfig{
			DB:       config.RedisDB(),
			Network:  config.RedisNetwork(),
			Addr:     config.RedisAddr(),
			Username: config.RedisUsername(),
			Password: config.RedisPassword(),
		})
		if err != nil {
			return errors.Wrap(err, "initializing redis client")
		}
		return nil
	})

	group.Go(func() error {
		log.Info().Msg("initializing otel tracing provider")
		var err error
		p.Tracing, err = tracing.NewProvider(ctx, tracing.ProviderConfig{
			OTLPExporterEndpoint: config.OTLPExporterEndpoint(),
			AppName:              config.AppName(),
			AppEnv:               config.AppEnv(),
		})
		if err != nil {
			return errors.Wrap(err, "initializing tracer provider")
		}
		return nil
	})

	group.Go(func() error {
		log.Info().Msg("initializing google books client")
		var err error
		p.GoogleBooks, err = book.NewAPI(book.GoogleBooks)
		if err != nil {
			return errors.Wrap(err, "initializing google books client")
		}
		return nil
	})

	group.Go(func() error {
		log.Info().Msg("initializing youtube client")
		var err error
		p.Youtube, err = youtube.New(ctx, config.YoutubeAPIKey())
		if err != nil {
			return errors.Wrap(err, "initializing youtube client")
		}
		return nil
	})

	if err := group.Wait(); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *Provider) Close(ctx context.Context) {
	p.DB.Close()
	p.Redis.Close()
	if err := p.Tracing.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed shutting down tracer provider")
	}
	log.Info().Msg("clients shutdown complete")
}
