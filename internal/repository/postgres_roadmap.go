package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/roadmap-thesis/backend/internal/domain"
	"github.com/roadmap-thesis/backend/pkg/cache"
	"github.com/roadmap-thesis/backend/pkg/database"
	"github.com/rs/zerolog/log"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/dialect"
	"github.com/stephenafamo/bob/dialect/psql/dm"
	"github.com/stephenafamo/bob/dialect/psql/im"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type RoadmapRepository struct {
	db     database.Connection
	cache  *cache.Connection
	tracer trace.Tracer
}

func NewPostgresRoadmapRepository(db database.Connection, cache *cache.Connection) *RoadmapRepository {
	tracer := otel.Tracer("db:postgres:roadmaps")
	return &RoadmapRepository{
		db:     db,
		cache:  cache,
		tracer: tracer,
	}
}

func (r *RoadmapRepository) GetBySlug(ctx context.Context, slug string) (domain.Roadmap, error) {
	cacher := cache.New[domain.Roadmap](r.cache)
	if roadmap, ok := cacher.Get(ctx, "roadmap:"+slug); ok {
		return roadmap, nil
	}

	query, args := psql.Select(
		sm.Columns(
			psql.Quote(domain.RoadmapTable, "id"),
			psql.Quote(domain.RoadmapTable, "account_id"),
			psql.Quote(domain.RoadmapTable, "title"),
			psql.Quote(domain.RoadmapTable, "slug"),
			psql.Quote(domain.RoadmapTable, "description"),
			psql.Quote(domain.RoadmapTable, "created_at"),
			psql.Quote(domain.RoadmapTable, "updated_at"),
			psql.Quote(domain.PersonalizationOptionsTable, "id"),
			psql.Quote(domain.PersonalizationOptionsTable, "account_id"),
			psql.Quote(domain.PersonalizationOptionsTable, "roadmap_id"),
			psql.Quote(domain.PersonalizationOptionsTable, "daily_time_availability"),
			psql.Quote(domain.PersonalizationOptionsTable, "total_duration"),
			psql.Quote(domain.PersonalizationOptionsTable, "skill_level"),
			psql.Quote(domain.PersonalizationOptionsTable, "additional_info"),
			psql.Quote(domain.PersonalizationOptionsTable, "created_at"),
			psql.Quote(domain.PersonalizationOptionsTable, "updated_at"),
		),
		sm.From(domain.RoadmapTable),
		sm.LeftJoin(domain.PersonalizationOptionsTable).
			OnEQ(psql.Quote(domain.PersonalizationOptionsTable, "roadmap_id"), psql.Quote(domain.RoadmapTable, "id")),
		sm.Where(psql.Quote(domain.RoadmapTable, "slug").EQ(psql.Arg(slug))),
	).MustBuild(ctx)

	roadmaps, err := r.fetch(ctx, query, args...)
	if err != nil {
		return domain.Roadmap{}, err
	}

	if len(roadmaps) == 0 {
		return domain.Roadmap{}, domain.ErrRoadmapNotFound
	}

	roadmap := roadmaps[0]
	topics, err := r.fetchTopicsByRoadmapID(ctx, roadmap.ID)
	if err != nil {
		return domain.Roadmap{}, err
	}

	roadmap.SetTopics(topics)

	cacher.Set(ctx, "roadmap:"+slug, roadmap)

	return roadmap, nil
}

func (r *RoadmapRepository) ListByAccountID(ctx context.Context, accountID int) ([]domain.Roadmap, error) {
	query, args := psql.Select(
		sm.Columns(
			psql.Quote(domain.RoadmapTable, "id"),
			psql.Quote(domain.RoadmapTable, "account_id"),
			psql.Quote(domain.RoadmapTable, "title"),
			psql.Quote(domain.RoadmapTable, "slug"),
			psql.Quote(domain.RoadmapTable, "description"),
			psql.Quote(domain.RoadmapTable, "created_at"),
			psql.Quote(domain.RoadmapTable, "updated_at"),
			psql.Quote(domain.PersonalizationOptionsTable, "id"),
			psql.Quote(domain.PersonalizationOptionsTable, "account_id"),
			psql.Quote(domain.PersonalizationOptionsTable, "roadmap_id"),
			psql.Quote(domain.PersonalizationOptionsTable, "daily_time_availability"),
			psql.Quote(domain.PersonalizationOptionsTable, "total_duration"),
			psql.Quote(domain.PersonalizationOptionsTable, "skill_level"),
			psql.Quote(domain.PersonalizationOptionsTable, "additional_info"),
			psql.Quote(domain.PersonalizationOptionsTable, "created_at"),
			psql.Quote(domain.PersonalizationOptionsTable, "updated_at"),
		),
		sm.From(domain.RoadmapTable),
		sm.LeftJoin(domain.PersonalizationOptionsTable).
			OnEQ(psql.Quote(domain.PersonalizationOptionsTable, "roadmap_id"), psql.Quote(domain.RoadmapTable, "id")),
		sm.Where(psql.Quote(domain.RoadmapTable, "account_id").EQ(psql.Arg(accountID))),
	).MustBuild(ctx)

	roadmaps, err := r.fetch(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	if len(roadmaps) == 0 {
		return nil, domain.ErrRoadmapNotFound
	}

	return roadmaps, nil
}

func (r *RoadmapRepository) fetch(ctx context.Context, query string, args ...any) ([]domain.Roadmap, error) {
	ctx, span := spanWithSelectQuery(ctx, r.tracer, "(*RoadmapRepository.fetch)", query)
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
		var personalizationOptions domain.PersonalizationOptions
		err = rows.Scan(
			&roadmap.ID,
			&roadmap.AccountID,
			&roadmap.Title,
			&roadmap.Slug,
			&roadmap.Description,
			&roadmap.CreatedAt,
			&roadmap.UpdatedAt,
			&personalizationOptions.ID,
			&personalizationOptions.AccountID,
			&personalizationOptions.RoadmapID,
			&personalizationOptions.DailyTimeAvailability,
			&personalizationOptions.TotalDuration,
			&personalizationOptions.SkillLevel,
			&personalizationOptions.AdditionalInfo,
			&personalizationOptions.CreatedAt,
			&personalizationOptions.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		roadmap.SetPersonalizationOptions(&personalizationOptions)
		roadmaps = append(roadmaps, roadmap)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return roadmaps, nil
}

func (r *RoadmapRepository) fetchTopicsByRoadmapID(ctx context.Context, roadmapID int) ([]*domain.Topic, error) {
	query, args := psql.Select(
		sm.Columns("id", "roadmap_id", "parent_id", "title", "slug", "description", psql.Quote("order"), "finished", "external_search_query", "created_at", "updated_at"),
		sm.From(domain.TopicTable),
		sm.Where(psql.Quote("roadmap_id").EQ(psql.Arg(roadmapID))),
		sm.OrderBy(psql.Quote("order")),
	).MustBuild(ctx)

	ctx, span := spanWithSelectQuery(ctx, r.tracer, "(*RoadmapRepository.fetchTopicsByRoadmapID)", query)
	defer span.End()

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		span.SetStatus(codes.Error, "failed to fetch roadmaps")
		span.RecordError(err)
		return nil, err
	}
	defer rows.Close()

	var topics []*domain.Topic
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

		topics = append(topics, &topic)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return topics, nil
}

func (r *RoadmapRepository) Save(ctx context.Context, input *domain.Roadmap) (domain.Roadmap, error) {
	query, args := psql.Insert(
		im.Into(domain.RoadmapTable, "account_id", "title", "slug", "description", "created_at", "updated_at"),
		im.Values(psql.Arg(input.AccountID, input.Title, input.Slug, input.Description, input.CreatedAt, input.UpdatedAt)),
		im.Returning("id", "slug"),
	).MustBuild(ctx)

	traceCtx, span := spanWithInsertQuery(ctx, r.tracer, "(*RoadmapRepository.Save)", query)
	defer span.End()

	var roadmap domain.Roadmap
	err := r.db.InTx(ctx, func(tx pgx.Tx) error {
		err := tx.QueryRow(ctx, query, args...).Scan(
			&roadmap.ID,
			&roadmap.Slug,
		)
		if err != nil {
			return err
		}

		if err = r.saveTopicsAndSubtopics(traceCtx, tx, roadmap.ID, input.Topics); err != nil {
			return err
		}

		if err = r.savePersonalizationOptions(traceCtx, tx, roadmap.ID, input.PersonalizationOptions); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return domain.Roadmap{}, err
	}

	return roadmap, nil
}

func (r *RoadmapRepository) saveTopicsAndSubtopics(ctx context.Context, tx pgx.Tx, roadmapID int, topics []*domain.Topic) error {
	// subTopicMap with topic's slug as the key to its subtopics
	subTopicMap := make(map[string][]*domain.Topic)

	// Insert the topics
	insertTopicMods := []bob.Mod[*dialect.InsertQuery]{
		im.Into(domain.TopicTable, "roadmap_id", "title", "slug", "description", "order", "finished", "external_search_query", "created_at", "updated_at"),
	}
	for _, topic := range topics {
		subTopicMap[topic.Slug] = topic.Subtopics
		arg := psql.Arg(roadmapID, topic.Title, topic.Slug, topic.Description, topic.Order, topic.Finished, topic.ExternalSearchQuery, topic.CreatedAt, topic.UpdatedAt)
		insertTopicMods = append(insertTopicMods, im.Values(arg))
	}
	insertTopicMods = append(insertTopicMods, im.Returning("id", "slug"))

	query, args := psql.Insert(
		insertTopicMods...,
	).MustBuild(ctx)

	ctx, span := spanWithInsertQuery(ctx, r.tracer, "(*RoadmapRepository.saveTopicsAndSubtopics)", query)
	defer span.End()

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	var mergedTopicAndSubtopic []*domain.Topic
	for rows.Next() {
		var savedTopic domain.Topic
		err = rows.Scan(
			&savedTopic.ID,
			&savedTopic.Slug,
		)
		if err != nil {
			return err
		}

		mergedTopicAndSubtopic = append(mergedTopicAndSubtopic, &savedTopic)
		mergedTopicAndSubtopic = append(mergedTopicAndSubtopic, subTopicMap[savedTopic.Slug]...)
	}

	if err = rows.Err(); err != nil {
		return err
	}

	log.Debug().Any("mergedTopicAndSubtopic", mergedTopicAndSubtopic).Send()

	var linkedSubtopics [][]any

	// Link the subtopics to their parent topic and set the roadmap ID
	parentID := 0
	for _, item := range mergedTopicAndSubtopic {
		// check if the current item is a parent topic, since we've
		// stored the parent topic first
		if item.ID != 0 {
			parentID = item.ID
			continue
		}

		// Link the subtopic
		linkedSubtopics = append(linkedSubtopics, []any{
			roadmapID, parentID,
			item.Title, item.Slug, item.Description, item.Order, item.Finished, item.ExternalSearchQuery,
			item.CreatedAt, item.UpdatedAt,
		})
	}
	log.Debug().Any("linkedSubtopics", linkedSubtopics).Send()

	// Store the subtopics
	_, err = tx.CopyFrom(ctx,
		pgx.Identifier{domain.TopicTable},
		[]string{"roadmap_id", "parent_id",
			"title", "slug", "description", "order", "finished", "external_search_query",
			"created_at", "updated_at"},
		pgx.CopyFromRows(linkedSubtopics),
	)

	return err
}

func (r *RoadmapRepository) savePersonalizationOptions(ctx context.Context, tx pgx.Tx, roadmapID int, input *domain.PersonalizationOptions) error {
	query, args := psql.Insert(
		im.Into(domain.PersonalizationOptionsTable,
			"account_id",
			"roadmap_id",
			"daily_time_availability",
			"total_duration",
			"skill_level",
			"additional_info",
			"created_at",
			"updated_at",
		),
		im.Values(psql.Arg(
			input.AccountID,
			roadmapID,
			input.DailyTimeAvailability,
			input.TotalDuration,
			input.SkillLevel,
			input.AdditionalInfo,
			input.CreatedAt,
			input.UpdatedAt,
		)),
	).MustBuild(ctx)

	ctx, span := spanWithInsertQuery(ctx, r.tracer, "(*RoadmapRepository.savePersonalizationOptions)", query)
	defer span.End()

	_, err := tx.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (r *RoadmapRepository) Delete(ctx context.Context, id int) (domain.Roadmap, error) {
	query, args := psql.Delete(
		dm.From(domain.RoadmapTable),
		dm.Where(psql.Quote("id").EQ(psql.Arg(id))),
	).MustBuild(ctx)

	ctx, span := spanWithDeleteQuery(ctx, r.tracer, "(*RoadmapRepository.Delete)", query)
	defer span.End()

	var roadmap domain.Roadmap
	err := r.db.InTx(ctx, func(tx pgx.Tx) error {
		err := tx.QueryRow(ctx, query, args...).Scan(
			&roadmap.ID,
			&roadmap.AccountID,
			&roadmap.Title,
			&roadmap.Slug,
			&roadmap.Description,
			&roadmap.CreatedAt,
			&roadmap.UpdatedAt,
		)
		if err != nil {
			span.SetStatus(codes.Error, "failed to delete roadmap")
			span.RecordError(err)
			return err
		}

		return nil
	})
	if err != nil {
		return domain.Roadmap{}, err
	}

	return roadmap, nil
}
