package option

import (
	"context"

	"github.com/curiona-org/backend/internal/config"
	"github.com/curiona-org/backend/internal/logger"
	"github.com/curiona-org/backend/internal/provider"
	"github.com/curiona-org/backend/pkg/cache"
	"github.com/curiona-org/backend/pkg/database"
	"github.com/curiona-org/backend/pkg/googleapi/book"
	"github.com/curiona-org/backend/pkg/googleapi/youtube"
	"github.com/curiona-org/backend/pkg/llm"
	"github.com/curiona-org/backend/pkg/redis"
	"github.com/curiona-org/backend/pkg/tracing"
	"github.com/hibiken/asynq"
	"github.com/pkg/errors"
)

// WithLLM provides a LLM client.
func WithLLM() provider.Option {
	return &withLLMProviderOption{}
}

type withLLMProviderOption struct{}

func (o *withLLMProviderOption) Apply(p *provider.Provider) {
	p.Group.Go(func() error {
		var err error
		p.LLM, err = llm.NewClient(
			config.LLMProvider(),
			config.LLMAPIKey(),
			config.LLMModel())
		if err != nil {
			return errors.Wrap(err, "initializing llm client")
		}
		log := logger.Get()
		log.Info().Msg("initialized llm client")
		return nil
	})
}

// WithPostgresDB provides a postgresql database connection.
func WithPostgresDB(ctx context.Context) provider.Option {
	return &withPostgresDBProviderOption{ctx: ctx}
}

type withPostgresDBProviderOption struct{ ctx context.Context }

func (o *withPostgresDBProviderOption) Apply(p *provider.Provider) {
	p.Group.Go(func() error {
		var err error
		p.DB, err = database.New(o.ctx, &database.Config{
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
		log := logger.Get()
		log.Info().Msg("initialized postgresql")
		return nil
	})
}

// WithCache provides a redis connection for caching.
func WithCache(ctx context.Context, cacheType cache.Type) provider.Option {
	return &withCacheProviderOption{ctx: ctx, cacheType: cacheType}
}

type withCacheProviderOption struct {
	ctx       context.Context
	cacheType cache.Type
}

func (o *withCacheProviderOption) Apply(p *provider.Provider) {
	p.Group.Go(func() error {
		var err error
		switch o.cacheType {
		case cache.TypeRedis:
			p.Cache, err = cache.NewConnection(o.ctx, &cache.Config{
				Type: cache.TypeRedis,
				RedisConfig: &redis.Config{
					DB:       config.RedisDB(),
					Network:  config.RedisNetwork(),
					Addr:     config.RedisAddr(),
					Username: config.RedisUsername(),
					Password: config.RedisPassword(),
				},
			})
		case cache.TypeNoop:
			p.Cache, err = cache.NewConnection(o.ctx, &cache.Config{
				Type: cache.TypeNoop,
			})
		case cache.TypeInMemory:
			p.Cache, err = cache.NewConnection(o.ctx, &cache.Config{
				Type: cache.TypeInMemory,
			})
		}
		if err != nil {
			return errors.Wrap(err, "initializing cache")
		}
		log := logger.Get()
		log.Info().Msgf("initialized cache: %s", o.cacheType)
		return nil
	})
}

// WithQueue provides an Asynq Queue. It must have a cache provider initialized
// with redis in order for the queue to work.
//
//	provider.New(
//		option.WithCache(ctx, cache.TypeRedis),
//		option.WithQueue()
//	)
func WithQueue() provider.Option {
	return &withQueueProviderOption{}
}

type withQueueProviderOption struct{}

func (*withQueueProviderOption) Apply(p *provider.Provider) {
	p.Group.Go(func() error {
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
		log := logger.Get()
		log.Info().Msg("initialized queue")
		return nil
	})
}

// WithYoutubeClient provides a Youtube Data API.
func WithYoutubeClient() provider.Option {
	return &withYoutubeClientProviderOption{}
}

type withYoutubeClientProviderOption struct{}

func (*withYoutubeClientProviderOption) Apply(p *provider.Provider) {
	p.Group.Go(func() error {
		p.Youtube = youtube.New(config.YoutubeAPIKey())
		log := logger.Get()
		log.Info().Msg("initialized youtube client")
		return nil
	})
}

// WithGoogleBooksClient provides a Google Books API client.
func WithGoogleBooksClient() provider.Option {
	return &withGoogleBooksClientProviderOption{}
}

type withGoogleBooksClientProviderOption struct{}

func (*withGoogleBooksClientProviderOption) Apply(p *provider.Provider) {
	p.Group.Go(func() error {
		p.GoogleBooks = book.New(config.GoogleBooksAPIKey())
		log := logger.Get()
		log.Info().Msg("initialized google books client")
		return nil
	})
}

// WithTracing provides a opentelemetry tracing provider.
func WithTracing(ctx context.Context) provider.Option {
	return &withTracingProviderOption{ctx: ctx}
}

type withTracingProviderOption struct{ ctx context.Context }

func (o *withTracingProviderOption) Apply(p *provider.Provider) {
	p.Group.Go(func() error {
		var err error
		p.Tracing, err = tracing.NewProvider(o.ctx, tracing.ProviderConfig{
			OTLPExporterEndpoint: config.OTLPExporterEndpoint(),
			AppName:              config.AppName(),
			AppEnv:               config.AppEnv(),
		})
		if err != nil {
			return errors.Wrap(err, "initializing tracer provider")
		}
		log := logger.Get()
		log.Info().Msg("initialized otel tracing provider")
		return nil
	})
}
