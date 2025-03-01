package repository

import (
	"context"

	"github.com/curiona-org/backend/internal/domain"
	"github.com/curiona-org/backend/internal/logger"
	"github.com/curiona-org/backend/pkg/cache"
	"github.com/curiona-org/backend/pkg/database"
	"github.com/curiona-org/backend/pkg/pagination"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/dialect"
	"github.com/stephenafamo/bob/dialect/psql/im"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/stephenafamo/bob/dialect/psql/um"
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
		return nil, err
	}

	return roadmaps, nil
}

func (r *RoadmapRepository) GetBySlug(ctx context.Context, slug string) (domain.Roadmap, error) {
	var roadmap domain.Roadmap

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
			psql.Quote(domain.RoadmapTable, "deleted_at"),
			psql.Quote(domain.PersonalizationOptionsTable, "id"),
			psql.Quote(domain.PersonalizationOptionsTable, "account_id"),
			psql.Quote(domain.PersonalizationOptionsTable, "roadmap_id"),
			psql.Quote(domain.PersonalizationOptionsTable, "daily_time_availability"),
			psql.Quote(domain.PersonalizationOptionsTable, "total_duration"),
			psql.Quote(domain.PersonalizationOptionsTable, "skill_level"),
			psql.Quote(domain.PersonalizationOptionsTable, "additional_info"),
			psql.Quote(domain.PersonalizationOptionsTable, "created_at"),
			psql.Quote(domain.PersonalizationOptionsTable, "updated_at"),
			psql.Quote(domain.AccountTable, "id"),
			psql.Quote(domain.AccountTable, "provider"),
			psql.Quote(domain.AccountTable, "email"),
			psql.Quote(domain.AccountTable, "is_suspended"),
			psql.Quote(domain.AccountTable, "is_admin"),
			psql.Quote(domain.AccountTable, "created_at"),
			psql.Quote(domain.AccountTable, "updated_at"),
			psql.Quote(domain.ProfileTable, "id"),
			psql.Quote(domain.ProfileTable, "name"),
			psql.Quote(domain.ProfileTable, "avatar"),
			psql.Quote(domain.ProfileTable, "created_at"),
			psql.Quote(domain.ProfileTable, "updated_at"),
		),
		sm.From(domain.RoadmapTable),
		sm.LeftJoin(domain.PersonalizationOptionsTable).
			OnEQ(psql.Quote(domain.PersonalizationOptionsTable, "roadmap_id"), psql.Quote(domain.RoadmapTable, "id")),
		sm.LeftJoin(domain.AccountTable).OnEQ(
			psql.Quote(domain.AccountTable, "id"), psql.Quote(domain.RoadmapTable, "account_id")),
		sm.LeftJoin(domain.ProfileTable).OnEQ(
			psql.Quote(domain.ProfileTable, "id"), psql.Quote(domain.RoadmapTable, "account_id")),
		sm.Where(psql.Quote(domain.RoadmapTable, "slug").EQ(psql.Arg(slug)).
			And(psql.Quote("deleted_at").IsNull())),
	).MustBuild(ctx)

	roadmaps, err := r.fetchAll(ctx, query, args...)
	if err != nil {
		return domain.Roadmap{}, err
	}

	if len(roadmaps) == 0 {
		return domain.Roadmap{}, domain.ErrRoadmapNotFound
	}

	roadmap = roadmaps[0]
	topics, err := r.fetchTopicsByRoadmapID(ctx, roadmap.ID)
	if err != nil {
		return domain.Roadmap{}, err
	}

	roadmap.SetTopics(topics)

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
			psql.Quote(domain.RoadmapTable, "total_topics"),
			psql.Quote(domain.RoadmapTable, "total_finished_topics"),
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
		sm.Where(psql.Quote(domain.RoadmapTable, "account_id").EQ(psql.Arg(accountID)).
			And(psql.Quote("deleted_at").IsNull())),
	).MustBuild(ctx)

	roadmaps, err := r.fetchWithPersonalizationOptions(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	if len(roadmaps) == 0 {
		return nil, domain.ErrRoadmapNotFound
	}

	return roadmaps, nil
}

func (r *RoadmapRepository) fetchWithPersonalizationOptions(ctx context.Context, query string, args ...any) ([]domain.Roadmap, error) {
	ctx, span := spanWithSelectQuery(ctx, r.tracer, "(*RoadmapRepository.fetchWithPersonalizationOptions)", query)
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
			&roadmap.TotalTopics,
			&roadmap.TotalFinishedTopics,
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
		sm.Columns("id", "roadmap_id", "parent_id", "title", "slug", "description", psql.Quote("order"), "is_finished", "external_search_query", "created_at", "updated_at"),
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

		topics = append(topics, &topic)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return topics, nil
}

func (r *RoadmapRepository) ListAll(ctx context.Context, pagination pagination.Paginator) ([]domain.Roadmap, error) {
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
			psql.Quote(domain.RoadmapTable, "deleted_at"),
			psql.Quote(domain.PersonalizationOptionsTable, "id"),
			psql.Quote(domain.PersonalizationOptionsTable, "account_id"),
			psql.Quote(domain.PersonalizationOptionsTable, "roadmap_id"),
			psql.Quote(domain.PersonalizationOptionsTable, "daily_time_availability"),
			psql.Quote(domain.PersonalizationOptionsTable, "total_duration"),
			psql.Quote(domain.PersonalizationOptionsTable, "skill_level"),
			psql.Quote(domain.PersonalizationOptionsTable, "additional_info"),
			psql.Quote(domain.PersonalizationOptionsTable, "created_at"),
			psql.Quote(domain.PersonalizationOptionsTable, "updated_at"),
			psql.Quote(domain.AccountTable, "id"),
			psql.Quote(domain.AccountTable, "provider"),
			psql.Quote(domain.AccountTable, "email"),
			psql.Quote(domain.AccountTable, "is_suspended"),
			psql.Quote(domain.AccountTable, "is_admin"),
			psql.Quote(domain.AccountTable, "created_at"),
			psql.Quote(domain.AccountTable, "updated_at"),
			psql.Quote(domain.ProfileTable, "id"),
			psql.Quote(domain.ProfileTable, "name"),
			psql.Quote(domain.ProfileTable, "avatar"),
			psql.Quote(domain.ProfileTable, "created_at"),
			psql.Quote(domain.ProfileTable, "updated_at"),
		),
		sm.From(domain.RoadmapTable),
		sm.LeftJoin(domain.PersonalizationOptionsTable).OnEQ(
			psql.Quote(domain.PersonalizationOptionsTable, "roadmap_id"), psql.Quote(domain.RoadmapTable, "id")),
		sm.LeftJoin(domain.AccountTable).OnEQ(
			psql.Quote(domain.AccountTable, "id"), psql.Quote(domain.RoadmapTable, "account_id")),
		sm.LeftJoin(domain.ProfileTable).OnEQ(
			psql.Quote(domain.ProfileTable, "id"), psql.Quote(domain.RoadmapTable, "account_id")),
		sm.OrderBy(psql.Quote(domain.RoadmapTable, "created_at")).Desc(),
		sm.Offset(psql.Arg(pagination.Skip)),
		sm.Limit(psql.Arg(pagination.Limit)),
	).MustBuild(ctx)

	return r.fetchAll(ctx, query, args...)
}

// fetchAll gets all roadmap related entities (accounts and profiles).
func (r *RoadmapRepository) fetchAll(ctx context.Context, query string, args ...any) ([]domain.Roadmap, error) {
	ctx, span := spanWithSelectQuery(ctx, r.tracer, "(*RoadmapRepository.fetchAll)", query)
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
		var roadmapDeletedAt pgtype.Timestamp
		var personalizationOptions domain.PersonalizationOptions
		var account domain.Account
		var profile domain.Profile
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
			&roadmapDeletedAt,
			&personalizationOptions.ID,
			&personalizationOptions.AccountID,
			&personalizationOptions.RoadmapID,
			&personalizationOptions.DailyTimeAvailability,
			&personalizationOptions.TotalDuration,
			&personalizationOptions.SkillLevel,
			&personalizationOptions.AdditionalInfo,
			&personalizationOptions.CreatedAt,
			&personalizationOptions.UpdatedAt,
			&account.ID,
			&account.Method,
			&account.Email,
			&account.IsSuspended,
			&account.IsAdmin,
			&account.CreatedAt,
			&account.UpdatedAt,
			&profile.ID,
			&profile.Name,
			&profile.Avatar,
			&profile.CreatedAt,
			&profile.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if roadmapDeletedAt.Valid {
			roadmap.DeletedAt = roadmapDeletedAt.Time
		}

		account.SetProfile(&profile)
		roadmap.SetCreator(&account)
		roadmap.SetPersonalizationOptions(&personalizationOptions)
		roadmaps = append(roadmaps, roadmap)
	}

	if err = rows.Err(); err != nil {
		span.SetStatus(codes.Error, "failed to fetch roadmaps")
		span.RecordError(err)
		return nil, err
	}

	return roadmaps, nil
}

func (r *RoadmapRepository) Count(ctx context.Context) (uint64, error) {
	query, args := psql.Select(
		sm.Columns(psql.F("COUNT", "*")),
		sm.From(domain.RoadmapTable),
	).MustBuild(ctx)

	var count uint64
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *RoadmapRepository) Save(ctx context.Context, input *domain.Roadmap) (domain.Roadmap, error) {
	query, args := psql.Insert(
		im.Into(domain.RoadmapTable, "account_id", "title", "slug", "description", "total_topics", "created_at", "updated_at"),
		im.Values(psql.Arg(input.AccountID, input.Title, input.Slug, input.Description, input.TotalTopics, input.CreatedAt, input.UpdatedAt)),
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
		im.Into(domain.TopicTable, "account_id", "roadmap_id", "title", "slug", "description", "order", "is_finished", "external_search_query", "created_at", "updated_at"),
	}
	for _, topic := range topics {
		subTopicMap[topic.Slug] = topic.Subtopics
		arg := psql.Arg(topic.AccountID, roadmapID, topic.Title, topic.Slug, topic.Description, topic.Order, topic.IsFinished, topic.ExternalSearchQuery, topic.CreatedAt, topic.UpdatedAt)
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

	log := logger.FromContext(ctx)

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
			item.AccountID, roadmapID, parentID,
			item.Title, item.Slug, item.Description, item.Order, item.IsFinished, item.ExternalSearchQuery,
			item.CreatedAt, item.UpdatedAt,
		})
	}
	log.Debug().Any("linkedSubtopics", linkedSubtopics).Send()

	// Store the subtopics
	_, err = tx.CopyFrom(ctx,
		pgx.Identifier{domain.TopicTable},
		[]string{"account_id", "roadmap_id", "parent_id",
			"title", "slug", "description", "order", "is_finished", "external_search_query",
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

func (r *RoadmapRepository) Update(ctx context.Context, slug string, updateFn func(roadmap *domain.Roadmap) (bool, error)) error {
	traceCtx, span := r.tracer.Start(ctx, "(*RoadmapRepository.Update)")
	defer span.End()

	err := r.db.InTx(ctx, func(tx pgx.Tx) error {
		fetchRoadmapQuery, fetchRoadmapArgs := psql.Select(
			sm.Columns(
				psql.Quote(domain.RoadmapTable, "id"),
				psql.Quote(domain.RoadmapTable, "account_id"),
				psql.Quote(domain.RoadmapTable, "title"),
				psql.Quote(domain.RoadmapTable, "slug"),
				psql.Quote(domain.RoadmapTable, "description"),
				psql.Quote(domain.RoadmapTable, "total_topics"),
				psql.Quote(domain.RoadmapTable, "completion_percentage"),
				psql.Quote(domain.RoadmapTable, "created_at"),
				psql.Quote(domain.RoadmapTable, "updated_at"),
			),
			sm.From(domain.RoadmapTable),
			sm.Where(
				psql.Quote("slug").EQ(psql.Arg(slug)).
					And(psql.Quote("deleted_at").IsNull()),
			),
		).MustBuild(ctx)

		roadmaps, err := r.fetch(traceCtx, fetchRoadmapQuery, fetchRoadmapArgs...)
		if err != nil {
			return err
		}

		if len(roadmaps) == 0 {
			return domain.ErrRoadmapNotFound
		}

		roadmap := roadmaps[0]
		updated, err := updateFn(&roadmap)
		if err != nil {
			return err
		}

		if !updated {
			return nil
		}

		mods := make([]bob.Mod[*dialect.UpdateQuery], 0)
		mods = append(mods, um.Table(domain.RoadmapTable))
		if roadmap.IsDeleted() {
			cacher := cache.New[domain.Roadmap](r.cache)
			err := cacher.Delete(traceCtx, &cache.Key{Namespace: domain.RoadmapTable, Key: slug})
			if err != nil {
				// TODO: should retry or log this error
				log.Error().Err(err).Msg("failed to delete roadmap from cache")
			}
			mods = append(mods, um.SetCol("deleted_at").ToArg(roadmap.DeletedAt))
		}
		mods = append(mods, um.Where(psql.Quote(domain.RoadmapTable, "slug").EQ(psql.Arg(slug))))

		updateRoadmapQuery, updateRoadmapArgs := psql.Update(
			mods...,
		).MustBuild(ctx)
		_, updateSpan := spanWithUpdateQuery(traceCtx, r.tracer, "(*RoadmapRepository.Update)", updateRoadmapQuery)
		defer updateSpan.End()

		if _, err = tx.Exec(ctx, updateRoadmapQuery, updateRoadmapArgs...); err != nil {
			updateSpan.SetStatus(codes.Error, "failed to update roadmap")
			updateSpan.RecordError(err)
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
