package repository

import (
	"context"

	"github.com/roadmap-thesis/backend/internal/domain"
	"github.com/roadmap-thesis/backend/pkg/database"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

type Repository struct {
	Account                domain.AccountRepository
	Roadmap                domain.RoadmapRepository
	Topic                  domain.TopicRepository
	PersonalizationOptions domain.PersonalizationOptionsRepository
	Session                domain.SessionRepository
}

func New(db database.Connection) *Repository {
	return &Repository{
		Account:                NewAccountRepository(db),
		Roadmap:                NewRoadmapRepository(db),
		Topic:                  NewTopicRepository(db),
		PersonalizationOptions: NewPersonalizationOptionsRepository(db),
		Session:                NewSessionRepository(db),
	}
}

func spanWithQuery(ctx context.Context, tracer trace.Tracer, method, query string) (context.Context, trace.Span) {
	ctx, span := tracer.Start(ctx, method)
	span.SetAttributes(semconv.DBStatementKey.String(query))
	return ctx, span
}
