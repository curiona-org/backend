package provider

import (
	"context"

	"github.com/curiona-org/backend/internal/logger"
	"github.com/curiona-org/backend/pkg/cache"
	"github.com/curiona-org/backend/pkg/database"
	"github.com/curiona-org/backend/pkg/googleapi/book"
	"github.com/curiona-org/backend/pkg/googleapi/youtube"
	"github.com/curiona-org/backend/pkg/llm"
	"github.com/hibiken/asynq"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/sdk/trace"
	"golang.org/x/sync/errgroup"
)

// Provider is a collection of clients to be used in the application.
type Provider struct {
	LLM         llm.Client
	DB          database.Connection
	Cache       *cache.Connection
	Queue       *asynq.Client
	QueueServer *asynq.Server
	Tracing     *trace.TracerProvider
	GoogleBooks book.Client
	Youtube     youtube.Client

	Group errgroup.Group
}

type Option interface {
	// Apply attaches a client to the provider.
	Apply(*Provider)
}

func New(opts ...Option) (*Provider, error) {
	p := &Provider{}
	for _, opt := range opts {
		opt.Apply(p)
	}

	return p.init()
}

// init initializes all the clients concurrently.
// If any of the clients fail to initialize, the function returns an error.
func (p *Provider) init() (*Provider, error) {
	if err := p.Group.Wait(); err != nil {
		return nil, errors.Wrap(err, "initializing clients")
	}

	return p, nil
}

// Close closes all the clients.
func (p *Provider) Close(ctx context.Context) {
	log := logger.FromContext(ctx)
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

	if p.QueueServer != nil {
		p.QueueServer.Stop()
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
