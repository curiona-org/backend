package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/roadmap-thesis/backend/internal/domain"
	"github.com/roadmap-thesis/backend/pkg/cache"
	"github.com/roadmap-thesis/backend/pkg/database"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type TopicRepository struct {
	db     database.Connection
	cache  *cache.Connection
	tracer trace.Tracer
}

func NewPostgresTopicRepository(db database.Connection, cache *cache.Connection) *TopicRepository {
	tracer := otel.Tracer("db:postgres:topics")
	return &TopicRepository{
		db:     db,
		cache:  cache,
		tracer: tracer,
	}
}

func (r *TopicRepository) GetBySlug(ctx context.Context, slug string) (domain.Topic, error) {
	query, args := psql.Select(
		sm.Columns(
			psql.Quote(domain.TopicTable, "id"),
			psql.Quote(domain.TopicTable, "roadmap_id"),
			psql.Quote(domain.TopicTable, "parent_id"),
			psql.Quote(domain.TopicTable, "title"),
			psql.Quote(domain.TopicTable, "slug"),
			psql.Quote(domain.TopicTable, "description"),
			psql.Quote(domain.TopicTable, "order"),
			psql.Quote(domain.TopicTable, "finished"),
			psql.Quote(domain.TopicTable, "external_search_query"),
			psql.Quote(domain.TopicTable, "created_at"),
			psql.Quote(domain.TopicTable, "updated_at"),
		),
		sm.From(domain.TopicTable),
		sm.Where(psql.Quote(domain.TopicTable, "slug").EQ(psql.Arg(slug))),
	).MustBuild(ctx)

	topics, err := r.fetch(ctx, query, args...)
	if err != nil {
		return domain.Topic{}, err
	}

	if len(topics) == 0 {
		return domain.Topic{}, domain.ErrTopicNotFound
	}

	topic := topics[0]
	var externalResources []domain.ExternalResource

	cacheKey := fmt.Sprintf("topics:%d:external_resources", topic.ID)

	traceCtx, span := r.tracer.Start(ctx, "(*TopicRepository.GetBySlug)")
	defer span.End()

	cacher := cache.New[domain.ExternalResource](r.cache)
	if cacher.Exists(ctx, cacheKey) {
		span.AddEvent("cache hit", trace.WithAttributes(attribute.String("cache_key", cacheKey)))
		resources, _ := cacher.List(traceCtx, cacheKey)
		externalResources = resources
		span.SetAttributes(
			attribute.Bool("cache_hit", true),
			attribute.Int("external_resources_count", len(externalResources)))
	} else {
		span.AddEvent("cache miss", trace.WithAttributes(attribute.String("cache_key", cacheKey)))
		externalResources, err = r.fetchExternalResourcesByTopicID(ctx, topic.ID)
		if err != nil && !errors.Is(err, domain.ErrExternalResourcesNotFound) {
			return domain.Topic{}, err
		}

		span.SetAttributes(
			attribute.Bool("cache_hit", false),
			attribute.Int("external_resources_count", len(externalResources)))
	}

	for _, resource := range externalResources {
		topic.AddResource(resource)
	}

	return topic, nil
}

func (r *TopicRepository) fetch(ctx context.Context, query string, args ...any) ([]domain.Topic, error) {
	ctx, span := spanWithSelectQuery(ctx, r.tracer, "(*TopicRepository.fetch)", query)
	defer span.End()

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		span.SetStatus(codes.Error, "failed to fetch topics")
		span.RecordError(err)
		return nil, err
	}
	defer rows.Close()

	var topics []domain.Topic
	for rows.Next() {
		var topic domain.Topic
		var topicParentID pgtype.Int4
		var externalSearchQuery pgtype.Text
		err = rows.Scan(
			&topic.ID,
			&topic.RoadmapID,
			&topicParentID,
			&topic.Title,
			&topic.Slug,
			&topic.Description,
			&topic.Order,
			&topic.Finished,
			&externalSearchQuery,
			&topic.CreatedAt,
			&topic.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if topicParentID.Valid {
			topic.ParentID = int(topicParentID.Int32)
		} else {
			topic.ParentID = 0
		}

		if externalSearchQuery.Valid {
			topic.ExternalSearchQuery = externalSearchQuery.String
		} else {
			topic.ExternalSearchQuery = ""
		}

		topics = append(topics, topic)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return topics, nil
}

func (r *TopicRepository) fetchExternalResourcesByTopicID(ctx context.Context, topicID int) ([]domain.ExternalResource, error) {
	cacher := cache.New[domain.ExternalResource](r.cache)
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

	traceCtx, span := spanWithSelectQuery(ctx, r.tracer, "(*TopicRepository.fetchExternalResourcesByTopicID)", query)
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

		cacheKey := fmt.Sprintf("topics:%d:external_resources", externalResource.TopicID)
		cacher.Push(traceCtx, cacheKey, externalResource)
		externalResources = append(externalResources, externalResource)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	if len(externalResources) == 0 {
		return nil, domain.ErrExternalResourcesNotFound
	}

	return externalResources, nil
}
