package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/roadmap-thesis/backend/internal/domain"
	"github.com/roadmap-thesis/backend/pkg/cache"
	"github.com/roadmap-thesis/backend/pkg/database"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type ExternalResourceRepository struct {
	db     database.Connection
	cache  cache.Cache[domain.ExternalResource]
	tracer trace.Tracer
}

func NewPostgresExternalResourceRepository(db database.Connection, cacheConn cache.Connection) *ExternalResourceRepository {
	tracer := otel.Tracer("db:postgres:external_resources")
	return &ExternalResourceRepository{
		db:     db,
		cache:  cache.NewRedisCache[domain.ExternalResource](cacheConn),
		tracer: tracer,
	}
}

func (r *ExternalResourceRepository) GetByTopicID(ctx context.Context, topicID int) ([]domain.ExternalResource, error) {
	if topicID == 0 {
		return nil, nil
	}

	if resources, ok := r.cache.List(ctx, fmt.Sprintf("topics:%d:external_resources", topicID)); ok {
		return resources, nil
	}

	query, args := psql.Select(
		sm.Columns(
			psql.Quote(domain.ExternalResourceTable, "id"),
			psql.Quote(domain.ExternalResourceTable, "topic_id"),
			psql.Quote(domain.ExternalResourceTable, "title"),
			psql.Quote(domain.ExternalResourceTable, "url"),
			psql.Quote(domain.ExternalResourceTable, "type"),
			psql.Quote(domain.ExternalResourceTable, "created_at"),
			psql.Quote(domain.ExternalResourceTable, "updated_at"),
		),
		sm.From(domain.ExternalResourceTable),
		sm.Where(psql.Quote(domain.ExternalResourceTable, "topic_id").EQ(psql.Arg(topicID))),
	).MustBuild(ctx)

	externalResources, err := r.fetch(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	if len(externalResources) == 0 {
		return nil, domain.ErrExternalResourcesNotFound
	}

	r.cache.Set(ctx, fmt.Sprintf("topics:%d:external_resources", topicID), externalResources...)

	return externalResources, nil
}

func (r *ExternalResourceRepository) fetch(ctx context.Context, query string, args ...any) ([]domain.ExternalResource, error) {
	ctx, span := spanWithSelectQuery(ctx, r.tracer, "(*ExternalResourceRepository.fetch)", query)
	defer span.End()

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		span.SetStatus(codes.Error, "failed to fetch external resources")
		span.RecordError(err)
		return nil, err
	}
	defer rows.Close()

	var externalResources []domain.ExternalResource
	for rows.Next() {
		var externalResource domain.ExternalResource
		err = rows.Scan(
			&externalResource.ID,
			&externalResource.TopicID,
			&externalResource.Title,
			&externalResource.URL,
			&externalResource.Type,
			&externalResource.CreatedAt,
			&externalResource.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		externalResources = append(externalResources, externalResource)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return externalResources, nil
}

func (r *ExternalResourceRepository) BulkSave(ctx context.Context, topicID int, resource []*domain.ExternalResource) error {
	ctx, span := r.tracer.Start(ctx, "(*ExternalResourceRepository.BulkSave)")
	defer span.End()

	err := r.db.InTx(ctx, func(tx pgx.Tx) error {
		var resources [][]any

		for _, res := range resource {
			r.cache.Push(ctx, fmt.Sprintf("topics:%d:external_resources", topicID), *res)

			resources = append(resources, []any{
				topicID, res.Title, res.URL, res.Type, res.CreatedAt, res.UpdatedAt,
			})
		}

		// if there are no resources to save, return
		if len(resources) == 0 {
			return nil
		}

		_, err := tx.CopyFrom(ctx,
			pgx.Identifier{domain.ExternalResourceTable},
			[]string{"topic_id", "title", "url", "type", "created_at", "updated_at"},
			pgx.CopyFromRows(resources),
		)
		return err
	})
	if err != nil {
		return err
	}

	return nil
}
