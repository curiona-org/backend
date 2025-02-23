package provider

import (
	"context"

	"github.com/hibiken/asynq"
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
	Queue       *asynq.Client
	QueueServer *asynq.Server
	Tracing     *trace.TracerProvider
	GoogleBooks book.Client
	Youtube     youtube.Client

	group errgroup.Group
}

func New() *Provider {
	return &Provider{}
}

func (p *Provider) WithLLM() *Provider {
	p.group.Go(func() error {
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

	return p
}

func (p *Provider) WithPostgresDB(ctx context.Context) *Provider {
	p.group.Go(func() error {
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
		log.Info().Msg("initialized postgresql")
		return nil
	})

	return p
}

func (p *Provider) WithRedisCache(ctx context.Context) *Provider {
	p.group.Go(func() error {
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
		log.Info().Msg("initialized cache")
		return nil
	})

	return p
}

func (p *Provider) WithNoopCache(ctx context.Context) *Provider {
	p.group.Go(func() error {
		var err error
		p.Cache, err = cache.NewConnection(ctx, &cache.Config{
			Type: cache.TypeNoop,
		})
		if err != nil {
			return errors.Wrap(err, "initializing cache")
		}
		log.Info().Msg("initialized cache")
		return nil
	})

	return p
}

func (p *Provider) WithQueue() *Provider {
	p.group.Go(func() error {
		p.Queue = asynq.NewClient(asynq.RedisClientOpt{
			DB:       config.RedisDB(),
			Network:  config.RedisNetwork(),
			Addr:     config.RedisAddr(),
			Username: config.RedisUsername(),
			Password: config.RedisPassword(),
		})

		p.QueueServer = asynq.NewServer(
			asynq.RedisClientOpt{
				DB:       config.RedisDB(),
				Network:  config.RedisNetwork(),
				Addr:     config.RedisAddr(),
				Username: config.RedisUsername(),
				Password: config.RedisPassword(),
			},
			asynq.Config{
				Concurrency: 10,
			},
		)

		log.Info().Msg("initialized queue")
		return nil
	})

	return p
}

func (p *Provider) WithGoogleBooksClient() *Provider {
	p.group.Go(func() error {
		p.GoogleBooks = book.New(config.GoogleBooksAPIKey())
		log.Info().Msg("initialized google books client")
		return nil
	})

	return p
}

func (p *Provider) WithYoutubeClient() *Provider {
	p.group.Go(func() error {
		p.Youtube = youtube.New(config.YoutubeAPIKey())
		log.Info().Msg("initialized youtube client")
		return nil
	})

	return p
}

func (p *Provider) WithTracing(ctx context.Context) *Provider {
	p.group.Go(func() error {
		var err error
		p.Tracing, err = tracing.NewProvider(ctx, tracing.ProviderConfig{
			OTLPExporterEndpoint: config.OTLPExporterEndpoint(),
			AppName:              config.AppName(),
			AppEnv:               config.AppEnv(),
		})
		if err != nil {
			return errors.Wrap(err, "initializing tracer provider")
		}
		log.Info().Msg("initialized otel tracing provider")
		return nil
	})

	return p
}

// Init initializes all the clients concurrently.
// If any of the clients fail to initialize, the function returns an error.
func (p *Provider) Init() (*Provider, error) {
	if err := p.group.Wait(); err != nil {
		return nil, errors.Wrap(err, "initializing clients")
	}

	return p, nil
}

func (p *Provider) Close(ctx context.Context) {
	if p.DB != nil {
		if err := p.DB.Close(); err != nil {
			log.Warn().Err(err).Msg("failed closing database connection")
		}
	}

	if p.Queue != nil {
		if err := p.Queue.Close(); err != nil {
			log.Warn().Err(err).Msg("failed closing queue connection")
		}
	}

	if p.Cache != nil {
		if err := p.Cache.Close(); err != nil {
			log.Warn().Err(err).Msg("failed closing cache connection")
		}
	}

	if p.Tracing != nil {
		if err := p.Tracing.Shutdown(ctx); err != nil {
			log.Warn().Err(err).Msg("failed shutting down tracer provider")
		}
	}

	log.Info().Msg("clients shutdown complete")
}
