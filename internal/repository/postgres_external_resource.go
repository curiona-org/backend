package repository

import (
	"context"
	"fmt"
	"strconv"

	"github.com/curiona-org/backend/internal/domain"
	"github.com/curiona-org/backend/pkg/cache"
	"github.com/curiona-org/backend/pkg/database"
	"github.com/jackc/pgx/v5"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type ExternalResourceRepository struct {
	db     database.Connection
	cache  *cache.Connection
	tracer trace.Tracer
}

func NewPostgresExternalResourceRepository(db database.Connection, cache *cache.Connection) *ExternalResourceRepository {
	tracer := otel.Tracer("db:postgres:external_resources")
	return &ExternalResourceRepository{
		db:     db,
		cache:  cache,
		tracer: tracer,
	}
}

func (r *ExternalResourceRepository) GetByTopicID(ctx context.Context, topicID int) ([]domain.ExternalResource, error) {
	if topicID == 0 {
		return nil, nil
	}

	cacher := cache.New[domain.ExternalResource](r.cache)
	if resources, ok := cacher.List(ctx, &cache.Key{
		Key: fmt.Sprintf("%s:%d:external_resources", domain.TopicTable, topicID),
	}); ok {
		return resources, nil
	}

	query, args := psql.Select(
		sm.Columns(
			psql.Quote(domain.ExternalResourceTable, "id"),
			psql.Quote(domain.ExternalResourceTable, "topic_id"),
			psql.Quote(domain.ExternalResourceTable, "title"),
			psql.Quote(domain.ExternalResourceTable, "author"),
			psql.Quote(domain.ExternalResourceTable, "url"),
			psql.Quote(domain.ExternalResourceTable, "cover_url"),
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

	cacher := cache.New[domain.ExternalResource](r.cache)
	var externalResources []domain.ExternalResource
	for rows.Next() {
		var externalResource domain.ExternalResource
		err = rows.Scan(
			&externalResource.ID,
			&externalResource.TopicID,
			&externalResource.Title,
			&externalResource.Author,
			&externalResource.URL,
			&externalResource.CoverURL,
			&externalResource.Type,
			&externalResource.CreatedAt,
			&externalResource.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// TODO: writing to caching inside the loop is not efficient
		// consider writing to cache in bulk or use pipelining
		cacher.Write(ctx, &cache.Key{
			Namespace: fmt.Sprintf("%s:%d:external_resources", domain.TopicTable, externalResource.TopicID),
			Key:       strconv.Itoa(externalResource.ID),
		}, externalResource, cache.DefaultTTL)

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
			resources = append(resources, []any{
				topicID, res.Title, res.Author, res.URL, res.CoverURL, res.Type, res.CreatedAt, res.UpdatedAt,
			})
		}

		// if there are no resources to save, return
		if len(resources) == 0 {
			return nil
		}

		_, err := tx.CopyFrom(ctx,
			pgx.Identifier{domain.ExternalResourceTable},
			[]string{"topic_id", "title", "author", "url", "cover_url", "type", "created_at", "updated_at"},
			pgx.CopyFromRows(resources),
		)
		return err
	})
	if err != nil {
		return err
	}

	return nil
}
