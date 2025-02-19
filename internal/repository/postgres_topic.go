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
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type topicRepository struct {
	db     database.Connection
	cache  cache.Cache[domain.ExternalResource]
	tracer trace.Tracer
}

var _ domain.TopicRepository = (*topicRepository)(nil)

func NewPostgresTopicRepository(db database.Connection, cacheConn cache.Connection) domain.TopicRepository {
	tracer := otel.Tracer("db:postgres:topics")
	return &topicRepository{
		db:     db,
		cache:  cache.NewRedisCache[domain.ExternalResource](cacheConn),
		tracer: tracer,
	}
}

func (r *topicRepository) GetBySlug(ctx context.Context, slug string) (domain.Topic, error) {
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
	externalResources := make([]domain.ExternalResource, 0)

	cacheKey := fmt.Sprintf("topics:%d:external_resources", topic.ID)

	if resources, ok := r.cache.List(ctx, cacheKey); ok {
		externalResources = resources
	} else {
		externalResources, err = r.fetchExternalResourcesByTopicID(ctx, topic.ID)
		if err != nil && !errors.Is(err, domain.ErrExternalResourcesNotFound) {
			return domain.Topic{}, err
		}

		r.cache.Push(ctx, cacheKey, externalResources...)
	}

	for _, resource := range externalResources {
		topic.AddResource(resource)
	}

	return topic, nil
}

func (r *topicRepository) fetch(ctx context.Context, query string, args ...any) ([]domain.Topic, error) {
	ctx, span := spanWithSelectQuery(ctx, r.tracer, "(*topicRepository.fetch)", query)
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
		err = rows.Scan(
			&topic.ID,
			&topic.RoadmapID,
			&topicParentID,
			&topic.Title,
			&topic.Slug,
			&topic.Description,
			&topic.Order,
			&topic.Finished,
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

		topics = append(topics, topic)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return topics, nil
}

func (r *topicRepository) fetchExternalResourcesByTopicID(ctx context.Context, topicID int) ([]domain.ExternalResource, error) {
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

	ctx, span := spanWithSelectQuery(ctx, r.tracer, "(*topicRepository.fetchExternalResourcesByTopicID)", query)
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

	if len(externalResources) == 0 {
		return nil, domain.ErrExternalResourcesNotFound
	}

	return externalResources, nil
}
