package repository

import (
	"context"

	"github.com/roadmap-thesis/backend/internal/domain"
	"github.com/roadmap-thesis/backend/pkg/cache"
	"github.com/roadmap-thesis/backend/pkg/database"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

type Repository interface {
	Account() domain.AccountRepository
	Roadmap() domain.RoadmapRepository
	Topic() domain.TopicRepository
	PersonalizationOptions() domain.PersonalizationOptionsRepository
	Session() domain.SessionRepository
}

type repository struct {
	account                domain.AccountRepository
	roadmap                domain.RoadmapRepository
	topic                  domain.TopicRepository
	personalizationOptions domain.PersonalizationOptionsRepository
	session                domain.SessionRepository
}

var _ Repository = (*repository)(nil)

func NewPostgresRepository(db database.Connection, cacheConn cache.Connection) Repository {
	return &repository{
		account:                NewPostgresAccountRepository(db),
		roadmap:                NewPostgresRoadmapRepository(db, cacheConn),
		topic:                  NewPostgresTopicRepository(db),
		personalizationOptions: NewPostgresPersonalizationOptionsRepository(db),
		session:                NewPostgresSessionRepository(db),
	}
}

func (r *repository) Account() domain.AccountRepository {
	return r.account
}

func (r *repository) Roadmap() domain.RoadmapRepository {
	return r.roadmap
}

func (r *repository) Topic() domain.TopicRepository {
	return r.topic
}

func (r *repository) PersonalizationOptions() domain.PersonalizationOptionsRepository {
	return r.personalizationOptions
}

func (r *repository) Session() domain.SessionRepository {
	return r.session
}

//nolint:spancheck
func spanWithQuery(ctx context.Context, tracer trace.Tracer, method, query string) (context.Context, trace.Span) {
	ctx, span := tracer.Start(ctx, method)
	span.SetAttributes(semconv.DBStatementKey.String(query))
	return ctx, span
}
