package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/curiona-org/backend/internal/domain"
	"github.com/curiona-org/backend/internal/logger"
	"github.com/curiona-org/backend/pkg/cache"
	"github.com/curiona-org/backend/pkg/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/im"
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
		psql.Quote(domain.TopicTable, "pro_tips"),
		psql.Quote(domain.TopicTable, "order"),
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
			&topic.ProTips,
			&topic.Order,
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
		span.SetStatus(codes.Error, "failed to fetch topics")
		span.RecordError(err)
		return nil, err
	}

	return topics, nil
}

func (r *TopicRepository) GetBySlug(ctx context.Context, slug string) (domain.Topic, error) {
	traceCtx, span := r.tracer.Start(ctx, "(*TopicRepository.GetBySlug)")
	defer span.End()

	log := logger.FromContext(ctx)

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

	// Fetch topic progression
	topicProgress, err := r.fetchTopicProgressionByID(traceCtx, topic.AccountID, topic.ID)
	if err != nil {
		log.Err(err).Msg("failed to fetch topic progression")
	}

	if !topicProgress.IsZero() {
		topic.IsFinished = topicProgress.IsFinished
		topic.FinishedAt = topicProgress.FinishedAt
	}

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

func (r *TopicRepository) UpdateTopicStatus(ctx context.Context, accountID int, slug string, updateFn func(roadmap *domain.Roadmap, topic *domain.Topic) (bool, error)) error {
	traceCtx, span := r.tracer.Start(ctx, "(*TopicRepository.UpdateTopicStatus)")
	defer span.End()

	log := logger.FromContext(ctx)
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

		log.Debug().Msgf("Topic found: %v", topic)
		topicProgress, err := r.fetchTopicProgressionByID(ctx, accountID, topic.ID)
		if err != nil {
			log.Err(err).Msg("failed to fetch topic progression")
		}

		if !topicProgress.IsZero() {
			topic.IsFinished = topicProgress.IsFinished
			topic.FinishedAt = topicProgress.FinishedAt
			log.Debug().Msgf("Topic progression found: %v", topicProgress)
		}

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

		progression, _ := r.fetchRoadmapProgressionByID(ctx, accountID, roadmap.ID)
		if progression.IsZero() {
			// Create initial progression
			saveInitialProgressionQuery, saveInitialProgressionArgs := psql.Insert(
				im.Into(domain.RoadmapProgressionTable, "account_id", "roadmap_id", "total_topics", "total_finished_topics"),
				im.Values(psql.Arg(accountID, roadmap.ID, roadmap.TotalTopics, 0)),
				im.Returning("id"),
			).MustBuild(ctx)

			ctx, span := spanWithInsertQuery(traceCtx, r.tracer, "(*TopicRepository.Insert)", saveInitialProgressionQuery)
			defer span.End()

			err = tx.QueryRow(ctx, saveInitialProgressionQuery, saveInitialProgressionArgs...).Scan(&progression.ID)
			if err != nil {
				span.SetStatus(codes.Error, "failed to save initial roadmap progression")
				span.RecordError(err)
				return err
			}
		}

		upsertTopicProgressionQuery, upsertTopicProgressionArgs := psql.Insert(
			im.Into(domain.RoadmapTopicProgressionTable, "progression_id", "account_id", "topic_id", "is_finished", "finished_at", "created_at", "updated_at"),
			im.Values(psql.Arg(progression.ID, accountID, topic.ID, topic.IsFinished, topic.FinishedAt, time.Now(), time.Now())),
			im.OnConflict("progression_id", "topic_id").DoUpdate(
				im.SetCol("is_finished").ToArg(topic.IsFinished),
				im.SetCol("finished_at").ToArg(topic.FinishedAt),
				im.SetCol("updated_at").ToArg(time.Now()),
				im.Where(psql.And(
					psql.Quote(domain.RoadmapTopicProgressionTable, "progression_id").EQ(psql.Arg(progression.ID)),
					psql.Quote(domain.RoadmapTopicProgressionTable, "topic_id").EQ(psql.Arg(topic.ID)))),
			),
		).MustBuild(ctx)
		_, upsertSpan := spanWithInsertQuery(traceCtx, r.tracer, "(*TopicRepository.Insert)", upsertTopicProgressionQuery)
		defer upsertSpan.End()

		if _, err = tx.Exec(ctx, upsertTopicProgressionQuery, upsertTopicProgressionArgs...); err != nil {
			upsertSpan.SetStatus(codes.Error, "failed to upsert topic progression")
			upsertSpan.RecordError(err)
			return err
		}

		// Update progression total finished
		newTotal := progression.TotalFinishedTopics + 1
		if !topic.IsFinished {
			newTotal = progression.TotalFinishedTopics - 1
		}

		if newTotal < 0 {
			newTotal = 0
		}

		var roadmapIsFinished bool
		if newTotal >= roadmap.TotalTopics {
			newTotal = roadmap.TotalTopics
			roadmapIsFinished = true
		} else {
			roadmapIsFinished = false
		}

		updateProgressionQB := psql.Update(
			um.Table(domain.RoadmapProgressionTable),
			um.SetCol("total_finished_topics").ToArg(newTotal),
			um.SetCol("is_finished").ToArg(roadmapIsFinished),
			um.SetCol("updated_at").ToArg(time.Now()),
		)

		if roadmapIsFinished {
			updateProgressionQB.Apply(um.SetCol("finished_at").ToArg(topic.FinishedAt))
		}

		updateProgressionQB.Apply(
			um.Where(psql.And(
				psql.Quote(domain.RoadmapProgressionTable, "id").EQ(psql.Arg(progression.ID)),
			)))

		updateProgressionQuery, updateProgressionArgs := updateProgressionQB.MustBuild(ctx)

		_, updateProgressionSpan := spanWithUpdateQuery(traceCtx, r.tracer, "(*TopicRepository.Update)", updateProgressionQuery)
		defer updateProgressionSpan.End()

		if _, err = tx.Exec(ctx, updateProgressionQuery, updateProgressionArgs...); err != nil {
			updateProgressionSpan.SetStatus(codes.Error, "failed to update roadmap progression")
			updateProgressionSpan.RecordError(err)
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

func (r *TopicRepository) fetchRoadmapProgressionByID(ctx context.Context, accountID, roadmapID int) (domain.RoadmapProgression, error) {
	query, args := psql.Select(
		sm.Columns("id", "account_id", "roadmap_id", "total_finished_topics"),
		sm.From(domain.RoadmapProgressionTable),
		sm.Where(psql.And(
			psql.Quote(domain.RoadmapProgressionTable, "account_id").EQ(psql.Arg(accountID)),
			psql.Quote(domain.RoadmapProgressionTable, "roadmap_id").EQ(psql.Arg(roadmapID)),
		)),
	).MustBuild(ctx)

	ctx, span := spanWithSelectQuery(ctx, r.tracer, "(*RoadmapRepository.GetRoadmapProgression)", query)
	defer span.End()

	var roadmapProgression domain.RoadmapProgression
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&roadmapProgression.ID,
		&roadmapProgression.AccountID,
		&roadmapProgression.RoadmapID,
		&roadmapProgression.TotalFinishedTopics,
	)
	if err != nil {
		span.SetStatus(codes.Error, "failed to fetch roadmap progression")
		span.RecordError(err)
		return domain.RoadmapProgression{}, err
	}

	return roadmapProgression, nil
}

func (r *TopicRepository) fetchTopicProgressionByID(ctx context.Context, accountID, topicID int) (domain.RoadmapTopicProgression, error) {
	query, args := psql.Select(
		sm.Columns("topic_id", "is_finished", "finished_at"),
		sm.From(domain.RoadmapTopicProgressionTable),
		sm.Where(psql.And(
			psql.Quote(domain.RoadmapTopicProgressionTable, "account_id").EQ(psql.Arg(accountID)),
			psql.Quote(domain.RoadmapTopicProgressionTable, "topic_id").EQ(psql.Arg(topicID)),
		)),
	).MustBuild(ctx)

	ctx, span := spanWithSelectQuery(ctx, r.tracer, "(*RoadmapRepository.GetTopicProgression)", query)
	defer span.End()

	var topicProgression domain.RoadmapTopicProgression
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&topicProgression.TopicID,
		&topicProgression.IsFinished,
		&topicProgression.FinishedAt,
	)
	if err != nil {
		span.SetStatus(codes.Error, "failed to fetch topic progression")
		span.RecordError(err)
		return domain.RoadmapTopicProgression{}, err
	}

	return topicProgression, nil
}
