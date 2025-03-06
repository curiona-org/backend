package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/curiona-org/backend/internal/domain"
	"github.com/curiona-org/backend/pkg/cache"
	"github.com/curiona-org/backend/pkg/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/stephenafamo/bob/dialect/psql/um"
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

func (r *TopicRepository) topicColumns() []any {
	return []any{
		psql.Quote(domain.TopicTable, "id"),
		psql.Quote(domain.TopicTable, "account_id"),
		psql.Quote(domain.TopicTable, "roadmap_id"),
		psql.Quote(domain.TopicTable, "parent_id"),
		psql.Quote(domain.TopicTable, "title"),
		psql.Quote(domain.TopicTable, "slug"),
		psql.Quote(domain.TopicTable, "description"),
		psql.Quote(domain.TopicTable, "order"),
		psql.Quote(domain.TopicTable, "is_finished"),
		psql.Quote(domain.TopicTable, "external_search_query"),
		psql.Quote(domain.TopicTable, "created_at"),
		psql.Quote(domain.TopicTable, "updated_at"),
	}
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
			&topic.AccountID,
			&topic.RoadmapID,
			&topicParentID,
			&topic.Title,
			&topic.Slug,
			&topic.Description,
			&topic.Order,
			&topic.IsFinished,
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
	}

	if err = rows.Err(); err != nil {
		span.SetStatus(codes.Error, "failed to fetch topics")
		span.RecordError(err)
		return nil, err
	}

	return topics, nil
}

func (r *TopicRepository) GetBySlug(ctx context.Context, slug string) (domain.Topic, error) {
	traceCtx, span := r.tracer.Start(ctx, "(*TopicRepository.GetBySlug)")
	defer span.End()

	query, args := psql.Select(
		sm.Columns(r.topicColumns()...),
		sm.From(domain.TopicTable),
		sm.Where(psql.Quote(domain.TopicTable, "slug").EQ(psql.Arg(slug))),
	).MustBuild(ctx)

	topics, err := r.fetch(traceCtx, query, args...)
	if err != nil {
		return domain.Topic{}, err
	}

	if len(topics) == 0 {
		return domain.Topic{}, domain.ErrTopicNotFound
	}

	topic := topics[0]

	// Fetch the associated external resources separately, either from cache or postgres.
	var externalResources []domain.ExternalResource
	externalResourceCacher := cache.New[domain.ExternalResource](r.cache)
	externalResourceCacheKey := &cache.Key{
		Key: fmt.Sprintf("%s:%d:external_resources", domain.TopicTable, topic.ID),
	}
	if externalResourceCacher.Exists(traceCtx, externalResourceCacheKey) {
		span.AddEvent("external resource cache hit")

		externalResources, _ = externalResourceCacher.List(traceCtx, externalResourceCacheKey)
	} else {
		span.AddEvent("external resource cache miss")

		var err error
		externalResources, err = r.fetchExternalResourcesByTopicID(traceCtx, topic.ID)
		if err != nil && !errors.Is(err, domain.ErrExternalResourcesNotFound) {
			return domain.Topic{}, err
		}
	}

	span.SetAttributes(attribute.Int("external_resources_count", len(externalResources)))

	for _, resource := range externalResources {
		topic.AddResource(resource)
	}

	return topic, nil
}

func (r *TopicRepository) UpdateTopicStatus(ctx context.Context, slug string, updateFn func(roadmap *domain.Roadmap, topic *domain.Topic) (bool, error)) error {
	traceCtx, span := r.tracer.Start(ctx, "(*TopicRepository.UpdateTopicStatus)")
	defer span.End()

	err := r.db.InTx(ctx, func(tx pgx.Tx) error {
		fetchTopicQuery, fetchTopicArgs := psql.Select(
			sm.Columns(r.topicColumns()...),
			sm.From(domain.TopicTable),
			sm.Where(psql.Quote("slug").EQ(psql.Arg(slug))),
		).MustBuild(ctx)

		topics, err := r.fetch(ctx, fetchTopicQuery, fetchTopicArgs...)
		if err != nil {
			return err
		}

		if len(topics) == 0 {
			return domain.ErrTopicNotFound
		}

		topic := topics[0]

		roadmaps, err := r.fetchRoadmapByID(ctx, topic.RoadmapID)
		if err != nil {
			return err
		}

		if len(roadmaps) == 0 {
			return domain.ErrRoadmapNotFound
		}

		roadmap := roadmaps[0]
		updated, err := updateFn(&roadmap, &topic)
		if err != nil {
			return err
		}

		if !updated {
			return nil
		}

		updateTopicQuery, updateTopicArgs := psql.Update(
			um.Table(domain.TopicTable),
			um.SetCol("is_finished").ToArg(topic.IsFinished),
			um.Where(psql.Quote(domain.TopicTable, "slug").EQ(psql.Arg(slug))),
		).MustBuild(ctx)
		_, updateSpan := spanWithUpdateQuery(traceCtx, r.tracer, "(*TopicRepository.Update)", updateTopicQuery)
		defer updateSpan.End()

		if _, err = tx.Exec(ctx, updateTopicQuery, updateTopicArgs...); err != nil {
			updateSpan.SetStatus(codes.Error, "failed to update topic")
			updateSpan.RecordError(err)
			return err
		}

		updateRoadmapQuery, updateRoadmapArgs := psql.Update(
			um.Table(domain.RoadmapTable),
			um.SetCol("total_finished_topics").ToArg(roadmap.TotalFinishedTopics),
			um.Where(psql.Quote(domain.RoadmapTable, "id").EQ(psql.Arg(roadmap.ID))),
		).MustBuild(ctx)
		_, updateRoadmapSpan := spanWithUpdateQuery(traceCtx, r.tracer, "(*TopicRepository.Update)", updateRoadmapQuery)
		defer updateRoadmapSpan.End()

		if _, err = tx.Exec(ctx, updateRoadmapQuery, updateRoadmapArgs...); err != nil {
			updateRoadmapSpan.SetStatus(codes.Error, "failed to update roadmap")
			updateRoadmapSpan.RecordError(err)
			return err
		}

		return nil
	})
	return err
}

func (r *TopicRepository) fetchExternalResourcesByTopicID(ctx context.Context, topicID int) ([]domain.ExternalResource, error) {
	cacher := cache.New[domain.ExternalResource](r.cache)
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

		cacher.Write(traceCtx, &cache.Key{
			Namespace: fmt.Sprintf("%s:%d:external_resources", domain.TopicTable, topicID),
			Key:       strconv.Itoa(externalResource.ID),
		}, externalResource, 0)
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

func (r *TopicRepository) fetchRoadmapByID(ctx context.Context, id int) ([]domain.Roadmap, error) {
	query, args := psql.Select(
		sm.Columns(
			psql.Quote(domain.RoadmapTable, "id"),
			psql.Quote(domain.RoadmapTable, "account_id"),
			psql.Quote(domain.RoadmapTable, "title"),
			psql.Quote(domain.RoadmapTable, "slug"),
			psql.Quote(domain.RoadmapTable, "description"),
			psql.Quote(domain.RoadmapTable, "total_topics"),
			psql.Quote(domain.RoadmapTable, "total_finished_topics"),
			psql.Quote(domain.RoadmapTable, "created_at"),
			psql.Quote(domain.RoadmapTable, "updated_at"),
		),
		sm.From(domain.RoadmapTable),
		sm.Where(psql.And(
			psql.Quote(domain.RoadmapTable, "id").EQ(psql.Arg(id)),
			psql.Quote("deleted_at").IsNull())),
	).MustBuild(ctx)
	ctx, span := spanWithSelectQuery(ctx, r.tracer, "(*TopicRepository.fetchRoadmapByID)", query)
	defer span.End()

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		span.SetStatus(codes.Error, "failed to fetch roadmaps")
		span.RecordError(err)
		return nil, err
	}
	defer rows.Close()

	var roadmaps []domain.Roadmap
	for rows.Next() {
		var roadmap domain.Roadmap
		err = rows.Scan(
			&roadmap.ID,
			&roadmap.AccountID,
			&roadmap.Title,
			&roadmap.Slug,
			&roadmap.Description,
			&roadmap.TotalTopics,
			&roadmap.TotalFinishedTopics,
			&roadmap.CreatedAt,
			&roadmap.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		roadmaps = append(roadmaps, roadmap)
	}

	if err = rows.Err(); err != nil {
		span.SetStatus(codes.Error, "failed to fetch roadmaps")
		span.RecordError(err)
		return nil, err
	}

	return roadmaps, nil
}
