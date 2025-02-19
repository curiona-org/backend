package provider

import (
	"context"

	"github.com/pkg/errors"
	"github.com/roadmap-thesis/backend/internal/cache"
	"github.com/roadmap-thesis/backend/internal/config"
	"github.com/roadmap-thesis/backend/internal/database"
	"github.com/roadmap-thesis/backend/internal/googleapi/book"
	"github.com/roadmap-thesis/backend/internal/googleapi/youtube"
	"github.com/roadmap-thesis/backend/internal/llm"
	"github.com/roadmap-thesis/backend/internal/redis"
	"github.com/roadmap-thesis/backend/internal/tracing"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/sdk/trace"
	"golang.org/x/sync/errgroup"
)

type Provider struct {
	LLM         llm.Client
	DB          database.Connection
	Cache       *cache.Connection
	Tracing     *trace.TracerProvider
	GoogleBooks book.Client
	Youtube     youtube.Client
}

func New(ctx context.Context) (*Provider, error) {
	p := &Provider{
		GoogleBooks: book.New(config.GoogleBooksAPIKey()),
		Youtube:     youtube.New(config.YoutubeAPIKey()),
	}

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
		log.Info().Msg("initializing cache")
		var err error
		p.Cache, err = cache.NewConnection(ctx, &cache.Config{
			Type: cache.TypeRedis,
			RedisConfig: &redis.Config{
				DB:       config.RedisDB(),
				Network:  config.RedisNetwork(),
				Addr:     config.RedisAddr(),
				Username: config.RedisUsername(),
				Password: config.RedisPassword(),
			},
		})
		if err != nil {
			return errors.Wrap(err, "initializing cache")
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

	if err := group.Wait(); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *Provider) Close(ctx context.Context) {
	if err := p.DB.Close(); err != nil {
		log.Warn().Err(err).Msg("failed closing database connection")
	}

	if err := p.Cache.Close(); err != nil {
		log.Warn().Err(err).Msg("failed closing cache connection")
	}

	if err := p.Tracing.Shutdown(ctx); err != nil {
		log.Warn().Err(err).Msg("failed shutting down tracer provider")
	}

	log.Info().Msg("clients shutdown complete")
}
