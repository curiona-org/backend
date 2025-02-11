package clients

import (
	"context"

	"github.com/pkg/errors"
	"github.com/roadmap-thesis/backend/pkg/auth/oauth"
	"github.com/roadmap-thesis/backend/pkg/config"
	"github.com/roadmap-thesis/backend/pkg/database"
	"github.com/roadmap-thesis/backend/pkg/llm"
	"github.com/roadmap-thesis/backend/pkg/tracing"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/sdk/trace"
	"golang.org/x/sync/errgroup"
)

type Clients struct {
	LLM     llm.Client
	DB      database.Connection
	Google  oauth.Client
	Tracing *trace.TracerProvider
}

func New(ctx context.Context) (*Clients, error) {
	c := &Clients{}

	var group errgroup.Group

	group.Go(func() error {
		var err error
		c.LLM, err = llm.NewClient(
			config.LLMProvider(),
			config.LLMAPIKey(),
			config.LLMModel())
		if err != nil {
			return errors.Wrap(err, "initializing llm client")
		}
		return nil
	})

	group.Go(func() error {
		var err error
		c.DB, err = database.New(ctx, &database.Config{
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
		var err error
		c.Tracing, err = tracing.NewProvider(ctx, tracing.ProviderConfig{
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

	return c, nil
}

func (c *Clients) Close(ctx context.Context) {
	c.DB.Close()
	if err := c.Tracing.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed shutting down tracer provider")
	}
	log.Info().Msg("clients shutdown complete")
}
