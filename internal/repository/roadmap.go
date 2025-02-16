package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
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

type roadmapRepository struct {
	db     database.Connection
	cache  cache.Cache[domain.Roadmap]
	tracer trace.Tracer
}

var _ domain.RoadmapRepository = (*roadmapRepository)(nil)

func NewRoadmapRepository(db database.Connection, cacheConn cache.Connection) domain.RoadmapRepository {
	tracer := otel.Tracer("db:postgres:roadmaps")
	return &roadmapRepository{
		db:     db,
		cache:  cache.NewRedisCache[domain.Roadmap](cacheConn),
		tracer: tracer,
	}
}

func (r *roadmapRepository) GetBySlug(ctx context.Context, slug string) (domain.Roadmap, error) {
	if roadmap, ok := r.cache.Get(ctx, "roadmap:"+slug); ok {
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

	r.cache.Set(ctx, "roadmap:"+slug, roadmap)

	return roadmap, nil
}

func (r *roadmapRepository) ListByAccountID(ctx context.Context, accountID int) ([]domain.Roadmap, error) {
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

func (r *roadmapRepository) fetch(ctx context.Context, query string, args ...any) ([]domain.Roadmap, error) {
	ctx, span := spanWithQuery(ctx, r.tracer, "(*roadmapRepository.fetch)", query)
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

func (r *roadmapRepository) fetchTopicsByRoadmapID(ctx context.Context, roadmapID int) ([]*domain.Topic, error) {
	query, args := psql.Select(
		sm.Columns("id", "roadmap_id", psql.F("COALESCE", "parent_id", 0), "title", "slug", "description", psql.Quote("order"), "finished", "created_at", "updated_at"),
		sm.From(domain.TopicTable),
		sm.Where(psql.Quote("roadmap_id").EQ(psql.Arg(roadmapID))),
		sm.OrderBy(psql.Quote("order")),
	).MustBuild(ctx)

	ctx, span := spanWithQuery(ctx, r.tracer, "(*roadmapRepository.fetchTopicsByRoadmapID)", query)
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
		err = rows.Scan(
			&topic.ID,
			&topic.RoadmapID,
			&topic.ParentID,
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

		topics = append(topics, &topic)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return topics, nil
}

func (r *roadmapRepository) Save(ctx context.Context, input *domain.Roadmap) (domain.Roadmap, error) {
	query, args := psql.Insert(
		im.Into(domain.RoadmapTable, "account_id", "title", "slug", "description", "created_at", "updated_at"),
		im.Values(psql.Arg(input.AccountID, input.Title, input.Slug, input.Description, input.CreatedAt, input.UpdatedAt)),
		im.Returning("id", "slug"),
	).MustBuild(ctx)

	traceCtx, span := spanWithQuery(ctx, r.tracer, "(*roadmapRepository.Save)", query)
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

func (r *roadmapRepository) saveTopicsAndSubtopics(ctx context.Context, tx pgx.Tx, roadmapID int, topics []*domain.Topic) error {
	// subTopicMap with topic's slug as the key to its subtopics
	subTopicMap := make(map[string][]*domain.Topic)

	// Insert the topics
	mods := []bob.Mod[*dialect.InsertQuery]{
		im.Into(domain.TopicTable, "roadmap_id", "title", "slug", "description", "order", "finished", "created_at", "updated_at"),
	}
	for _, topic := range topics {
		subTopicMap[topic.Slug] = topic.Subtopics
		arg := psql.Arg(roadmapID, topic.Title, topic.Slug, topic.Description, topic.Order, topic.Finished, topic.CreatedAt, topic.UpdatedAt)
		mods = append(mods, im.Values(arg))
	}
	mods = append(mods, im.Returning("id", "slug"))

	query, args := psql.Insert(
		mods...,
	).MustBuild(ctx)

	ctx, span := spanWithQuery(ctx, r.tracer, "(*roadmapRepository.saveTopicsAndSubtopics)", query)
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
			roadmapID, parentID, item.Title, item.Slug, item.Description, item.Order, item.Finished, item.CreatedAt, item.UpdatedAt,
		})
	}
	log.Debug().Any("linkedSubtopics", linkedSubtopics).Send()

	// Store the subtopics
	_, err = tx.CopyFrom(ctx,
		pgx.Identifier{domain.TopicTable},
		[]string{"roadmap_id", "parent_id", "title", "slug", "description", "order", "finished", "created_at", "updated_at"},
		pgx.CopyFromRows(linkedSubtopics),
	)

	return err
}

func (r *roadmapRepository) savePersonalizationOptions(ctx context.Context, tx pgx.Tx, roadmapID int, input *domain.PersonalizationOptions) error {
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

	ctx, span := spanWithQuery(ctx, r.tracer, "(*roadmapRepository.savePersonalizationOptions)", query)
	defer span.End()

	_, err := tx.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (r *roadmapRepository) Delete(ctx context.Context, id int) (domain.Roadmap, error) {
	query, args := psql.Delete(
		dm.From(domain.RoadmapTable),
		dm.Where(psql.Quote("id").EQ(psql.Arg(id))),
	).MustBuild(ctx)

	ctx, span := spanWithQuery(ctx, r.tracer, "(*roadmapRepository.Delete)", query)
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
