package repository

import (
	"context"

	"github.com/curiona-org/backend/internal/domain"
	"github.com/curiona-org/backend/internal/filter"
	"github.com/curiona-org/backend/internal/logger"
	"github.com/curiona-org/backend/pkg/cache"
	"github.com/curiona-org/backend/pkg/database"
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

type roadmapColumnsOptions struct {
	includeBookmark              bool
	includeProgression           bool
	includePersonalizationOption bool
	includeAccount               bool
}

func (r *RoadmapRepository) roadmapColumns(opt roadmapColumnsOptions) []any {
	columns := []any{
		psql.Quote(domain.RoadmapTable, "id"),
		psql.Quote(domain.RoadmapTable, "account_id"),
		psql.Quote(domain.RoadmapTable, "title"),
		psql.Quote(domain.RoadmapTable, "slug"),
		psql.Quote(domain.RoadmapTable, "description"),
		psql.Quote(domain.RoadmapTable, "total_topics"),
		psql.Quote(domain.RoadmapTable, "created_at"),
		psql.Quote(domain.RoadmapTable, "updated_at"),
		psql.Quote(domain.RoadmapTable, "deleted_at"),
	}

	if opt.includeBookmark {
		columns = append(columns,
			psql.Case().
				When(psql.Quote(domain.BookmarkTable, "account_id").IsNotNull(), psql.S("true")).
				Else(psql.S("false")).
				As("is_bookmarked"))
	}

	if opt.includeProgression {
		columns = append(columns,
			psql.Quote(domain.RoadmapProgressionTable, "id"),
			psql.Quote(domain.RoadmapProgressionTable, "account_id"),
			psql.Quote(domain.RoadmapProgressionTable, "roadmap_id"),
			psql.Quote(domain.RoadmapProgressionTable, "total_topics"),
			psql.F("COALESCE", psql.Quote(domain.RoadmapProgressionTable, "total_finished_topics"), 0),
			psql.Quote(domain.RoadmapProgressionTable, "is_finished"),
			psql.Quote(domain.RoadmapProgressionTable, "finished_at"),
			psql.Quote(domain.RoadmapProgressionTable, "created_at"),
			psql.Quote(domain.RoadmapProgressionTable, "updated_at"))
	}

	if opt.includePersonalizationOption {
		columns = append(columns,
			psql.Quote(domain.PersonalizationOptionsTable, "id"),
			psql.Quote(domain.PersonalizationOptionsTable, "account_id"),
			psql.Quote(domain.PersonalizationOptionsTable, "roadmap_id"),
			psql.Quote(domain.PersonalizationOptionsTable, "daily_time_availability"),
			psql.Quote(domain.PersonalizationOptionsTable, "total_duration"),
			psql.Quote(domain.PersonalizationOptionsTable, "skill_level"),
			psql.Quote(domain.PersonalizationOptionsTable, "additional_info"),
			psql.Quote(domain.PersonalizationOptionsTable, "created_at"),
			psql.Quote(domain.PersonalizationOptionsTable, "updated_at"),
		)
	}

	if opt.includeAccount {
		columns = append(columns,
			psql.Quote(domain.AccountTable, "id"),
			psql.Quote(domain.AccountTable, "method"),
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
		)
	}

	return columns
}

type roadmapFetchConfig struct {
	query                        string
	args                         []any
	includeBookmark              bool
	includeProgression           bool
	includePersonalizationOption bool
	includeAccount               bool
	options                      map[string]any
}

func (r *RoadmapRepository) fetch(ctx context.Context, cfg roadmapFetchConfig) ([]domain.Roadmap, error) {
	ctx, span := spanWithSelectQuery(ctx, r.tracer, "(*RoadmapRepository.fetch)", cfg.query)
	defer span.End()

	rows, err := r.db.Query(ctx, cfg.query, cfg.args...)
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
		dest := []any{
			&roadmap.ID,
			&roadmap.AccountID,
			&roadmap.Title,
			&roadmap.Slug,
			&roadmap.Description,
			&roadmap.TotalTopics,
			&roadmap.CreatedAt,
			&roadmap.UpdatedAt,
			&roadmapDeletedAt,
		}

		var roadmapIsBookmarked pgtype.Bool
		if cfg.includeBookmark {
			dest = append(dest, &roadmapIsBookmarked)
		}

		var roadmapProgressionID, roadmapProgressionAccountID, roadmapProgressionRoadmapID, roadmapProgressionTotalTopics, roadmapProgressionTotalFinishedTopics pgtype.Int4
		var roadmapProgressionIsFinished pgtype.Bool
		var roadmapProgressionCreatedAt, roadmapProgressionUpdatedAt pgtype.Timestamp
		var roadmapProgressionFinishedAt pgtype.Timestamp
		if cfg.includeProgression {
			dest = append(dest,
				&roadmapProgressionID,
				&roadmapProgressionAccountID,
				&roadmapProgressionRoadmapID,
				&roadmapProgressionTotalTopics,
				&roadmapProgressionTotalFinishedTopics,
				&roadmapProgressionIsFinished,
				&roadmapProgressionFinishedAt,
				&roadmapProgressionCreatedAt,
				&roadmapProgressionUpdatedAt,
			)
		}

		var personalizationOptions domain.PersonalizationOptions
		if cfg.includePersonalizationOption {
			dest = append(dest,
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
		}

		var account domain.Account
		var profile domain.Profile
		if cfg.includeAccount {
			dest = append(dest,
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
		}

		if cfg.options == nil {
			cfg.options = make(map[string]any)
		}

		var totalBookmarks pgtype.Int8
		adminWithTotalBookmarks, adminWithTotalBookmarksOk := cfg.options["admin.with_total_bookmarks"].(bool)
		if adminWithTotalBookmarksOk && adminWithTotalBookmarks {
			dest = append(dest, &totalBookmarks)
		}

		err = rows.Scan(dest...)
		if err != nil {
			return nil, err
		}

		if roadmapDeletedAt.Valid {
			roadmap.DeletedAt = roadmapDeletedAt.Time
		}

		if roadmapIsBookmarked.Valid {
			roadmap.IsBookmarked = roadmapIsBookmarked.Bool
		}

		if adminWithTotalBookmarksOk && adminWithTotalBookmarks && totalBookmarks.Valid {
			roadmap.TotalBookmarks = int(totalBookmarks.Int64)
		} else {
			roadmap.TotalBookmarks = 0
		}

		if cfg.includeProgression && roadmapProgressionID.Valid {
			roadmapProgression := new(domain.RoadmapProgression)
			if roadmapProgressionFinishedAt.Valid {
				roadmapProgression.FinishedAt = roadmapProgressionFinishedAt.Time
			}

			roadmapProgression.ID = int(roadmapProgressionID.Int32)
			roadmapProgression.AccountID = int(roadmapProgressionAccountID.Int32)
			roadmapProgression.RoadmapID = int(roadmapProgressionRoadmapID.Int32)
			roadmapProgression.TotalTopics = int(roadmapProgressionTotalTopics.Int32)
			roadmapProgression.TotalFinishedTopics = int(roadmapProgressionTotalFinishedTopics.Int32)
			roadmapProgression.IsFinished = roadmapProgressionIsFinished.Bool
			roadmapProgression.CreatedAt = roadmapProgressionCreatedAt.Time
			roadmapProgression.UpdatedAt = roadmapProgressionUpdatedAt.Time
			roadmap.SetProgression(roadmapProgression)
		}

		if cfg.includePersonalizationOption {
			roadmap.SetPersonalizationOptions(&personalizationOptions)
		}

		if cfg.includeAccount {
			account.SetProfile(&profile)
			roadmap.SetCreator(&account)
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

func (r *RoadmapRepository) GetBySlug(ctx context.Context, filter filter.Filters) (domain.Roadmap, error) {
	var roadmap domain.Roadmap

	query, args := psql.Select(
		sm.Columns(r.roadmapColumns(roadmapColumnsOptions{
			includeBookmark:              false,
			includeProgression:           false,
			includePersonalizationOption: true,
			includeAccount:               true,
		})...),
		sm.From(domain.RoadmapTable),
		sm.LeftJoin(domain.PersonalizationOptionsTable).OnEQ(
			psql.Quote(domain.PersonalizationOptionsTable, "roadmap_id"),
			psql.Quote(domain.RoadmapTable, "id")),
		sm.LeftJoin(domain.AccountTable).OnEQ(
			psql.Quote(domain.AccountTable, "id"),
			psql.Quote(domain.RoadmapTable, "account_id")),
		sm.LeftJoin(domain.ProfileTable).OnEQ(
			psql.Quote(domain.ProfileTable, "id"),
			psql.Quote(domain.RoadmapTable, "account_id")),
		sm.Where(psql.And(
			psql.Quote(domain.RoadmapTable, "slug").EQ(psql.Arg(filter.Slug)),
			psql.Quote(domain.RoadmapTable, "deleted_at").IsNull())),
	).MustBuild(ctx)

	roadmaps, err := r.fetch(ctx, roadmapFetchConfig{
		query:                        query,
		args:                         args,
		includeBookmark:              false,
		includeProgression:           false,
		includePersonalizationOption: true,
		includeAccount:               true,
	})
	if err != nil {
		return domain.Roadmap{}, err
	}

	if len(roadmaps) == 0 {
		return domain.Roadmap{}, domain.ErrRoadmapNotFound
	}

	roadmap = roadmaps[0]
	topics, err := r.fetchTopicsByRoadmapID(ctx, roadmap.ID, filter)
	if err != nil {
		return domain.Roadmap{}, err
	}

	roadmap.SetTopics(topics)

	return roadmap, nil
}

func (r *RoadmapRepository) GetByID(ctx context.Context, id int) (domain.Roadmap, error) {
	var roadmap domain.Roadmap

	query, args := psql.Select(
		sm.Columns(r.roadmapColumns(roadmapColumnsOptions{
			includeProgression:           true,
			includePersonalizationOption: true,
			includeAccount:               true,
		})...),
		sm.From(domain.RoadmapTable),
		sm.LeftJoin(domain.BookmarkTable).OnEQ(
			psql.Quote(domain.BookmarkTable, "roadmap_id"),
			psql.Quote(domain.RoadmapTable, "id")),
		sm.LeftJoin(domain.RoadmapProgressionTable).OnEQ(
			psql.Quote(domain.RoadmapProgressionTable, "roadmap_id"),
			psql.Quote(domain.RoadmapTable, "id")),
		sm.LeftJoin(domain.PersonalizationOptionsTable).OnEQ(
			psql.Quote(domain.PersonalizationOptionsTable, "roadmap_id"),
			psql.Quote(domain.RoadmapTable, "id")),
		sm.LeftJoin(domain.AccountTable).OnEQ(
			psql.Quote(domain.AccountTable, "id"),
			psql.Quote(domain.RoadmapTable, "account_id")),
		sm.LeftJoin(domain.ProfileTable).OnEQ(
			psql.Quote(domain.ProfileTable, "id"),
			psql.Quote(domain.RoadmapTable, "account_id")),
		sm.Where(psql.And(
			psql.Quote(domain.RoadmapTable, "id").EQ(psql.Arg(id)),
			psql.Quote(domain.RoadmapTable, "deleted_at").IsNull())),
	).MustBuild(ctx)

	roadmaps, err := r.fetch(ctx, roadmapFetchConfig{
		query:                        query,
		args:                         args,
		includeProgression:           true,
		includePersonalizationOption: true,
		includeAccount:               true,
	})
	if err != nil {
		return domain.Roadmap{}, err
	}

	if len(roadmaps) == 0 {
		return domain.Roadmap{}, domain.ErrRoadmapNotFound
	}

	roadmap = roadmaps[0]
	topics, err := r.fetchTopicsByRoadmapID(ctx, roadmap.ID, filter.Filters{}) // TODO: fix filters
	if err != nil {
		return domain.Roadmap{}, err
	}

	roadmap.SetTopics(topics)

	return roadmap, nil
}

func (r *RoadmapRepository) ListAll(ctx context.Context, filters filter.Filters) ([]domain.Roadmap, error) {
	colOpt := roadmapColumnsOptions{
		includePersonalizationOption: true,
		includeAccount:               true,
	}

	cols := r.roadmapColumns(colOpt)
	adminWithTotalBookmarks, adminWithTotalBookmarksOk := filters.Options["admin.with_total_bookmarks"].(bool)
	if adminWithTotalBookmarksOk && adminWithTotalBookmarks {
		cols = append(cols, psql.F("COUNT", psql.Quote(domain.BookmarkTable, "roadmap_id")))
	}

	if filters.AccountID > 0 {
		colOpt.includeBookmark = true
		colOpt.includeProgression = true
	}

	selectQuery := psql.Select(
		sm.Columns(cols...),
		sm.From(domain.RoadmapTable),
		sm.LeftJoin(domain.PersonalizationOptionsTable).OnEQ(
			psql.Quote(domain.PersonalizationOptionsTable, "roadmap_id"),
			psql.Quote(domain.RoadmapTable, "id")),
		sm.LeftJoin(domain.AccountTable).OnEQ(
			psql.Quote(domain.AccountTable, "id"),
			psql.Quote(domain.RoadmapTable, "account_id")),
		sm.LeftJoin(domain.ProfileTable).OnEQ(
			psql.Quote(domain.ProfileTable, "id"),
			psql.Quote(domain.RoadmapTable, "account_id")),
	)

	if filters.AccountID > 0 {
		selectQuery.Apply(
			sm.LeftJoin(domain.BookmarkTable).On(
				psql.And(
					psql.Quote(domain.BookmarkTable, "roadmap_id").EQ(psql.Quote(domain.RoadmapTable, "id")),
					psql.Quote(domain.BookmarkTable, "account_id").EQ(psql.Arg(filters.AccountID)),
				),
			),
			sm.LeftJoin(domain.RoadmapProgressionTable).On(
				psql.And(
					psql.Quote(domain.RoadmapProgressionTable, "roadmap_id").EQ(psql.Quote(domain.RoadmapTable, "id")),
					psql.Quote(domain.RoadmapProgressionTable, "account_id").EQ(psql.Arg(filters.AccountID)),
				),
			))
	} else {
		selectQuery.Apply(sm.LeftJoin(domain.RoadmapProgressionTable).OnEQ(
			psql.Quote(domain.RoadmapProgressionTable, "roadmap_id"),
			psql.Quote(domain.RoadmapTable, "id")),
		)
	}

	if adminWithTotalBookmarksOk && adminWithTotalBookmarks {
		selectQuery.Apply(
			sm.LeftJoin(domain.BookmarkTable).OnEQ(
				psql.Quote(domain.BookmarkTable, "roadmap_id"),
				psql.Quote(domain.RoadmapTable, "id")),
		)
	}

	if filters.Search != "" {
		selectQuery.Apply(
			sm.Where(psql.And(
				psql.Or(
					psql.Quote(domain.RoadmapTable, "title").ILike(psql.Arg("%"+filters.Search+"%")),
					psql.Quote(domain.RoadmapTable, "description").ILike(psql.Arg("%"+filters.Search+"%")),
					psql.Quote(domain.ProfileTable, "name").ILike(psql.Arg("%"+filters.Search+"%")),
				),
				psql.Quote(domain.RoadmapTable, "deleted_at").IsNull()),
			))
	} else {
		selectQuery.Apply(sm.Where(psql.Quote(domain.RoadmapTable, "deleted_at").IsNull()))
	}

	if adminWithTotalBookmarksOk && adminWithTotalBookmarks {
		selectQuery.Apply(
			sm.GroupBy(psql.Quote(domain.RoadmapTable, "id")),
			sm.GroupBy(psql.Quote(domain.PersonalizationOptionsTable, "id")),
			sm.GroupBy(psql.Quote(domain.AccountTable, "id")),
			sm.GroupBy(psql.Quote(domain.ProfileTable, "id")),
		)
	}

	if filters.OrderBy == filter.OrderByOldest {
		selectQuery.Apply(sm.OrderBy(psql.Quote(domain.RoadmapTable, "created_at")).Asc())
	} else {
		selectQuery.Apply(sm.OrderBy(psql.Quote(domain.RoadmapTable, "created_at")).Desc())
	}

	selectQuery.Apply(
		sm.Offset(psql.Arg(filters.Paginator.Skip)),
		sm.Limit(psql.Arg(filters.Paginator.Limit)),
	)

	query, args := selectQuery.MustBuild(ctx)

	fetchOpt := roadmapFetchConfig{
		query:                        query,
		args:                         args,
		includeBookmark:              colOpt.includeBookmark,
		includeProgression:           colOpt.includeProgression,
		includePersonalizationOption: colOpt.includePersonalizationOption,
		includeAccount:               colOpt.includeAccount,
		options:                      filters.Options,
	}

	return r.fetch(ctx, fetchOpt)
}

func (r *RoadmapRepository) ListByAccountID(ctx context.Context, filters filter.Filters) ([]domain.Roadmap, error) {
	selectQuery := psql.Select(
		sm.Columns(r.roadmapColumns(roadmapColumnsOptions{
			includeBookmark:              true,
			includeProgression:           true,
			includePersonalizationOption: true,
			includeAccount:               true,
		})...),
		sm.From(domain.RoadmapTable),
		sm.LeftJoin(domain.BookmarkTable).On(
			psql.And(
				psql.Quote(domain.BookmarkTable, "roadmap_id").EQ(psql.Quote(domain.RoadmapTable, "id")),
				psql.Quote(domain.BookmarkTable, "account_id").EQ(psql.Arg(filters.AccountID)),
			)),
		sm.LeftJoin(domain.RoadmapProgressionTable).On(
			psql.And(
				psql.Quote(domain.RoadmapProgressionTable, "roadmap_id").EQ(psql.Quote(domain.RoadmapTable, "id")),
				psql.Quote(domain.RoadmapProgressionTable, "account_id").EQ(psql.Arg(filters.AccountID)),
			)),
		sm.LeftJoin(domain.PersonalizationOptionsTable).OnEQ(
			psql.Quote(domain.PersonalizationOptionsTable, "roadmap_id"),
			psql.Quote(domain.RoadmapTable, "id")),
		sm.LeftJoin(domain.AccountTable).OnEQ(
			psql.Quote(domain.AccountTable, "id"),
			psql.Quote(domain.RoadmapTable, "account_id")),
		sm.LeftJoin(domain.ProfileTable).OnEQ(
			psql.Quote(domain.ProfileTable, "id"),
			psql.Quote(domain.RoadmapTable, "account_id")),
	)

	if filters.Search != "" {
		selectQuery.Apply(
			sm.Where(psql.And(
				psql.Or(
					psql.Quote(domain.RoadmapTable, "title").ILike(psql.Arg("%"+filters.Search+"%")),
					psql.Quote(domain.RoadmapTable, "description").ILike(psql.Arg("%"+filters.Search+"%")),
					psql.Quote(domain.ProfileTable, "name").ILike(psql.Arg("%"+filters.Search+"%")),
				),
				psql.Quote(domain.RoadmapTable, "account_id").EQ(psql.Arg(filters.AccountID)),
				psql.Quote(domain.RoadmapTable, "deleted_at").IsNull()),
			))
	} else {
		selectQuery.Apply(sm.Where(psql.And(
			psql.Quote(domain.RoadmapTable, "account_id").EQ(psql.Arg(filters.AccountID)),
			psql.Quote(domain.RoadmapTable, "deleted_at").IsNull())),
		)
	}

	if filters.OrderBy == filter.OrderByOldest {
		selectQuery.Apply(sm.OrderBy(psql.Quote(domain.RoadmapTable, "created_at")).Asc())
	} else {
		selectQuery.Apply(sm.OrderBy(psql.Quote(domain.RoadmapTable, "created_at")).Desc())
	}

	selectQuery.Apply(
		sm.Offset(psql.Arg(filters.Paginator.Skip)),
		sm.Limit(psql.Arg(filters.Paginator.Limit)),
	)

	query, args := selectQuery.MustBuild(ctx)

	return r.fetch(ctx, roadmapFetchConfig{
		query:                        query,
		args:                         args,
		includeBookmark:              true,
		includeProgression:           true,
		includePersonalizationOption: true,
		includeAccount:               true,
	})
}

func (r *RoadmapRepository) ListAccountOnProgressRoadmaps(ctx context.Context, accountID int, filters filter.Filters) ([]domain.Roadmap, error) {
	selectQuery := psql.Select(
		sm.Columns(r.roadmapColumns(roadmapColumnsOptions{
			includeBookmark:              true,
			includeProgression:           true,
			includePersonalizationOption: true,
			includeAccount:               true,
		})...),
		sm.From(domain.RoadmapTable),
		sm.LeftJoin(domain.BookmarkTable).On(
			psql.And(
				psql.Quote(domain.BookmarkTable, "roadmap_id").EQ(psql.Quote(domain.RoadmapTable, "id")),
				psql.Quote(domain.BookmarkTable, "account_id").EQ(psql.Arg(accountID)),
			)),
		sm.LeftJoin(domain.RoadmapProgressionTable).On(
			psql.And(
				psql.Quote(domain.RoadmapProgressionTable, "roadmap_id").EQ(psql.Quote(domain.RoadmapTable, "id")),
				psql.Quote(domain.RoadmapProgressionTable, "account_id").EQ(psql.Arg(accountID)),
			)),
		sm.LeftJoin(domain.PersonalizationOptionsTable).OnEQ(
			psql.Quote(domain.PersonalizationOptionsTable, "roadmap_id"),
			psql.Quote(domain.RoadmapTable, "id")),
		sm.LeftJoin(domain.AccountTable).OnEQ(
			psql.Quote(domain.AccountTable, "id"),
			psql.Quote(domain.RoadmapTable, "account_id")),
		sm.LeftJoin(domain.ProfileTable).OnEQ(
			psql.Quote(domain.ProfileTable, "id"),
			psql.Quote(domain.RoadmapTable, "account_id")),
	)

	if filters.Search != "" {
		selectQuery.Apply(
			sm.Where(psql.And(
				psql.Or(
					psql.Quote(domain.RoadmapTable, "title").ILike(psql.Arg("%"+filters.Search+"%")),
					psql.Quote(domain.RoadmapTable, "description").ILike(psql.Arg("%"+filters.Search+"%")),
					psql.Quote(domain.ProfileTable, "name").ILike(psql.Arg("%"+filters.Search+"%")),
				),
				psql.Quote(domain.RoadmapProgressionTable, "account_id").EQ(psql.Arg(accountID)),
				psql.Quote(domain.RoadmapProgressionTable, "is_finished").EQ(psql.S("false")),
				psql.Quote(domain.RoadmapProgressionTable, "total_finished_topics").GT(psql.Arg(0)),
				psql.Quote(domain.RoadmapTable, "deleted_at").IsNull()),
			))
	} else {
		selectQuery.Apply(sm.Where(psql.And(
			psql.Quote(domain.RoadmapProgressionTable, "account_id").EQ(psql.Arg(accountID)),
			psql.Quote(domain.RoadmapProgressionTable, "is_finished").EQ(psql.S("false")),
			psql.Quote(domain.RoadmapProgressionTable, "total_finished_topics").GT(psql.Arg(0)),
			psql.Quote(domain.RoadmapTable, "deleted_at").IsNull(),
		)))
	}

	selectQuery.Apply(sm.OrderBy(psql.Quote(domain.RoadmapProgressionTable, "total_finished_topics")).Desc())

	selectQuery.Apply(
		sm.Offset(psql.Arg(filters.Paginator.Skip)),
		sm.Limit(psql.Arg(filters.Paginator.Limit)),
	)

	query, args := selectQuery.MustBuild(ctx)
	ctx, span := spanWithSelectQuery(ctx, r.tracer, "(*RoadmapRepository.ListAccountOnProgressRoadmaps)", query)
	defer span.End()

	return r.fetch(ctx, roadmapFetchConfig{
		query:                        query,
		args:                         args,
		includeBookmark:              true,
		includeProgression:           true,
		includePersonalizationOption: true,
		includeAccount:               true,
	})
}

func (r *RoadmapRepository) ListAccountFinishedRoadmaps(ctx context.Context, accountID int, filters filter.Filters) ([]domain.Roadmap, error) {
	selectQuery := psql.Select(
		sm.Columns(r.roadmapColumns(roadmapColumnsOptions{
			includeBookmark:              true,
			includeProgression:           true,
			includePersonalizationOption: true,
			includeAccount:               true,
		})...),
		sm.From(domain.RoadmapTable),
		sm.LeftJoin(domain.BookmarkTable).On(
			psql.And(
				psql.Quote(domain.BookmarkTable, "roadmap_id").EQ(psql.Quote(domain.RoadmapTable, "id")),
				psql.Quote(domain.BookmarkTable, "account_id").EQ(psql.Arg(accountID)),
			)),
		sm.LeftJoin(domain.RoadmapProgressionTable).On(
			psql.And(
				psql.Quote(domain.RoadmapProgressionTable, "roadmap_id").EQ(psql.Quote(domain.RoadmapTable, "id")),
				psql.Quote(domain.RoadmapProgressionTable, "account_id").EQ(psql.Arg(accountID)),
			)),
		sm.LeftJoin(domain.PersonalizationOptionsTable).OnEQ(
			psql.Quote(domain.PersonalizationOptionsTable, "roadmap_id"),
			psql.Quote(domain.RoadmapTable, "id")),
		sm.LeftJoin(domain.AccountTable).OnEQ(
			psql.Quote(domain.AccountTable, "id"),
			psql.Quote(domain.RoadmapTable, "account_id")),
		sm.LeftJoin(domain.ProfileTable).OnEQ(
			psql.Quote(domain.ProfileTable, "id"),
			psql.Quote(domain.RoadmapTable, "account_id")),
	)

	if filters.Search != "" {
		selectQuery.Apply(
			sm.Where(psql.And(
				psql.Or(
					psql.Quote(domain.RoadmapTable, "title").ILike(psql.Arg("%"+filters.Search+"%")),
					psql.Quote(domain.RoadmapTable, "description").ILike(psql.Arg("%"+filters.Search+"%")),
					psql.Quote(domain.ProfileTable, "name").ILike(psql.Arg("%"+filters.Search+"%")),
				),
				psql.Quote(domain.RoadmapProgressionTable, "account_id").EQ(psql.Arg(accountID)),
				psql.Quote(domain.RoadmapProgressionTable, "is_finished").EQ(psql.S("true")),
				psql.Quote(domain.RoadmapProgressionTable, "total_finished_topics").EQ(psql.Quote(domain.RoadmapProgressionTable, "total_topics")),
				psql.Quote(domain.RoadmapTable, "deleted_at").IsNull()),
			))
	} else {
		selectQuery.Apply(sm.Where(psql.And(
			psql.Quote(domain.RoadmapProgressionTable, "account_id").EQ(psql.Arg(accountID)),
			psql.Quote(domain.RoadmapProgressionTable, "is_finished").EQ(psql.S("true")),
			psql.Quote(domain.RoadmapProgressionTable, "total_finished_topics").EQ(psql.Quote(domain.RoadmapProgressionTable, "total_topics")),
			psql.Quote(domain.RoadmapTable, "deleted_at").IsNull(),
		)))
	}

	selectQuery.Apply(sm.OrderBy(psql.Quote(domain.RoadmapProgressionTable, "finished_at")).Desc())

	selectQuery.Apply(
		sm.Offset(psql.Arg(filters.Paginator.Skip)),
		sm.Limit(psql.Arg(filters.Paginator.Limit)),
	)

	query, args := selectQuery.MustBuild(ctx)

	ctx, span := spanWithSelectQuery(ctx, r.tracer, "(*RoadmapRepository.ListAccountFinishedRoadmaps)", query)
	defer span.End()

	return r.fetch(ctx, roadmapFetchConfig{
		query:                        query,
		args:                         args,
		includeBookmark:              true,
		includeProgression:           true,
		includePersonalizationOption: true,
		includeAccount:               true,
	})
}

func (r *RoadmapRepository) GetRoadmapProgression(ctx context.Context, accountID, roadmapID int) (domain.RoadmapProgression, error) {
	query, args := psql.Select(
		sm.Columns("id", "account_id", "roadmap_id", "total_topics", "total_finished_topics", "is_finished", "finished_at", "created_at", "updated_at"),
		sm.From(domain.RoadmapProgressionTable),
		sm.Where(psql.And(
			psql.Quote(domain.RoadmapProgressionTable, "account_id").EQ(psql.Arg(accountID)),
			psql.Quote(domain.RoadmapProgressionTable, "roadmap_id").EQ(psql.Arg(roadmapID)),
		)),
	).MustBuild(ctx)

	ctx, span := spanWithSelectQuery(ctx, r.tracer, "(*RoadmapRepository.GetRoadmapProgression)", query)
	defer span.End()

	var roadmapProgression domain.RoadmapProgression
	var roadmapProgressionFinishedAt pgtype.Timestamp
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&roadmapProgression.ID,
		&roadmapProgression.AccountID,
		&roadmapProgression.RoadmapID,
		&roadmapProgression.TotalTopics,
		&roadmapProgression.TotalFinishedTopics,
		&roadmapProgression.IsFinished,
		&roadmapProgressionFinishedAt,
		&roadmapProgression.CreatedAt,
		&roadmapProgression.UpdatedAt,
	)
	if err != nil {
		span.SetStatus(codes.Error, "failed to fetch roadmap progression")
		span.RecordError(err)
		return domain.RoadmapProgression{}, err
	}

	if roadmapProgressionFinishedAt.Valid {
		roadmapProgression.FinishedAt = roadmapProgressionFinishedAt.Time
	}

	// Fetch the topics progression
	topicsProgressionQuery, topicsProgressionArgs := psql.Select(
		sm.Columns("topic_id", "is_finished", "finished_at"),
		sm.From(domain.RoadmapTopicProgressionTable),
		sm.Where(psql.Quote("progression_id").EQ(psql.Arg(roadmapProgression.ID))),
	).MustBuild(ctx)

	topicsProgressionRows, err := r.db.Query(ctx, topicsProgressionQuery, topicsProgressionArgs...)
	if err != nil {
		span.SetStatus(codes.Error, "failed to fetch topics progression")
		span.RecordError(err)
		return domain.RoadmapProgression{}, err
	}
	defer topicsProgressionRows.Close()

	roadmapProgression.TopicProgressionMap = make(map[int]*domain.RoadmapTopicProgression)

	for topicsProgressionRows.Next() {
		var topicProgression domain.RoadmapTopicProgression
		err = topicsProgressionRows.Scan(
			&topicProgression.TopicID,
			&topicProgression.IsFinished,
			&topicProgression.FinishedAt,
		)
		if err != nil {
			span.SetStatus(codes.Error, "failed to scan topics progression")
			span.RecordError(err)
			return domain.RoadmapProgression{}, err
		}

		roadmapProgression.TopicProgressionMap[topicProgression.TopicID] = &topicProgression
	}

	return roadmapProgression, nil
}

func (r *RoadmapRepository) Count(ctx context.Context) (uint64, error) {
	query, args := psql.Select(
		sm.Columns(psql.F("COUNT", "*")),
		sm.From(domain.RoadmapTable),
		sm.Where(psql.Quote(domain.RoadmapTable, "deleted_at").IsNull()),
	).MustBuild(ctx)

	ctx, span := spanWithSelectQuery(ctx, r.tracer, "(*RoadmapRepository.Count)", query)
	defer span.End()

	var count uint64
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		span.SetStatus(codes.Error, "failed to count roadmaps")
		span.RecordError(err)
		return 0, err
	}

	return count, nil
}

func (r *RoadmapRepository) CountOnProgress(ctx context.Context) (uint64, error) {
	query, args := psql.Select(
		sm.Columns(psql.F("COUNT", sm.Distinct(psql.Quote(domain.RoadmapProgressionTable, "roadmap_id")))),
		sm.From(domain.RoadmapTable),
		sm.LeftJoin(domain.RoadmapProgressionTable).On(
			psql.Quote(domain.RoadmapProgressionTable, "roadmap_id").EQ(psql.Quote(domain.RoadmapTable, "id")),
		),
		sm.Where(psql.And(
			psql.Quote(domain.RoadmapProgressionTable, "is_finished").EQ(psql.S("false")),
			psql.Quote(domain.RoadmapProgressionTable, "total_finished_topics").GT(psql.Arg(0)),
			psql.Quote(domain.RoadmapTable, "deleted_at").IsNull(),
		)),
	).MustBuild(ctx)

	ctx, span := spanWithSelectQuery(ctx, r.tracer, "(*RoadmapRepository.Count)", query)
	defer span.End()

	var count uint64
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		span.SetStatus(codes.Error, "failed to count roadmaps")
		span.RecordError(err)
		return 0, err
	}

	return count, nil
}

func (r *RoadmapRepository) CountFinished(ctx context.Context) (uint64, error) {
	query, args := psql.Select(
		sm.Columns(psql.F("COUNT", sm.Distinct(psql.Quote(domain.RoadmapProgressionTable, "roadmap_id")))),
		sm.From(domain.RoadmapTable),
		sm.LeftJoin(domain.RoadmapProgressionTable).On(
			psql.Quote(domain.RoadmapProgressionTable, "roadmap_id").EQ(psql.Quote(domain.RoadmapTable, "id")),
		),
		sm.Where(psql.And(
			psql.Quote(domain.RoadmapProgressionTable, "is_finished").EQ(psql.S("true")),
			psql.Quote(domain.RoadmapTable, "deleted_at").IsNull(),
		)),
	).MustBuild(ctx)

	ctx, span := spanWithSelectQuery(ctx, r.tracer, "(*RoadmapRepository.Count)", query)
	defer span.End()

	var count uint64
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		span.SetStatus(codes.Error, "failed to count roadmaps")
		span.RecordError(err)
		return 0, err
	}

	return count, nil
}

func (r *RoadmapRepository) CountBySearching(ctx context.Context, search string) (uint64, error) {
	query, args := psql.Select(
		sm.Columns(psql.F("COUNT", "*")),
		sm.From(domain.RoadmapTable),
		sm.Where(psql.And(
			psql.Or(
				psql.Quote(domain.RoadmapTable, "title").ILike(psql.Arg("%"+search+"%")),
				psql.Quote(domain.RoadmapTable, "description").ILike(psql.Arg("%"+search+"%")),
			),
			psql.Quote(domain.RoadmapTable, "deleted_at").IsNull(),
		)),
	).MustBuild(ctx)

	ctx, span := spanWithSelectQuery(ctx, r.tracer, "(*RoadmapRepository.CountBySearching)", query)
	defer span.End()

	var count uint64
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		span.SetStatus(codes.Error, "failed to count roadmaps by searching")
		span.RecordError(err)
		return 0, err
	}

	return count, nil
}

func (r *RoadmapRepository) CountByAccountID(ctx context.Context, accountID int) (uint64, error) {
	query, args := psql.Select(
		sm.Columns(psql.F("COUNT", "*")),
		sm.From(domain.RoadmapTable),
		sm.Where(psql.And(
			psql.Quote(domain.RoadmapTable, "account_id").EQ(psql.Arg(accountID)),
			psql.Quote(domain.RoadmapTable, "deleted_at").IsNull(),
		)),
	).MustBuild(ctx)

	ctx, span := spanWithSelectQuery(ctx, r.tracer, "(*RoadmapRepository.CountByAccountID)", query)
	defer span.End()

	var count uint64
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		span.SetStatus(codes.Error, "failed to count user roadmaps")
		span.RecordError(err)
		return 0, err
	}

	return count, nil
}

func (r *RoadmapRepository) CountByAccountIdAndSearch(ctx context.Context, accountID int, search string) (uint64, error) {
	query, args := psql.Select(
		sm.Columns(psql.F("COUNT", "*")),
		sm.From(domain.RoadmapTable),
		sm.Where(psql.And(
			psql.Quote(domain.RoadmapTable, "account_id").EQ(psql.Arg(accountID)),
			psql.Or(
				psql.Quote(domain.RoadmapTable, "title").ILike(psql.Arg("%"+search+"%")),
				psql.Quote(domain.RoadmapTable, "description").ILike(psql.Arg("%"+search+"%")),
			),
			psql.Quote(domain.RoadmapTable, "deleted_at").IsNull(),
		)),
	).MustBuild(ctx)

	ctx, span := spanWithSelectQuery(ctx, r.tracer, "(*RoadmapRepository.CountAccountAndSearch)", query)
	defer span.End()

	var count uint64
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		span.SetStatus(codes.Error, "failed to count user roadmaps by searching")
		span.RecordError(err)
		return 0, err
	}

	return count, nil
}

func (r *RoadmapRepository) CountAccountFinishedRoadmaps(ctx context.Context, accountID int) (uint64, error) {
	query, args := psql.Select(
		sm.Columns(psql.F("COUNT", "*")),
		sm.From(domain.RoadmapProgressionTable),
		sm.Where(psql.And(
			psql.Quote(domain.RoadmapProgressionTable, "account_id").EQ(psql.Arg(accountID)),
			psql.Quote(domain.RoadmapProgressionTable, "is_finished").EQ(psql.S("true")),
		)),
	).MustBuild(ctx)

	ctx, span := spanWithSelectQuery(ctx, r.tracer, "(*RoadmapRepository.CountAccountFinishedRoadmaps)", query)
	defer span.End()

	var count uint64
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		span.SetStatus(codes.Error, "failed to count finished roadmaps by account ID")
		span.RecordError(err)
		return 0, err
	}

	return count, nil
}

func (r *RoadmapRepository) CountAccountOnProgressRoadmaps(ctx context.Context, accountID int) (uint64, error) {
	query, args := psql.Select(
		sm.Columns(psql.F("COUNT", "*")),
		sm.From(domain.RoadmapProgressionTable),
		sm.Where(psql.And(
			psql.Quote(domain.RoadmapProgressionTable, "account_id").EQ(psql.Arg(accountID)),
			psql.Quote(domain.RoadmapProgressionTable, "is_finished").EQ(psql.S("false")),
			psql.Quote(domain.RoadmapProgressionTable, "total_finished_topics").GT(psql.Arg(0)),
			psql.Quote(domain.RoadmapTable, "deleted_at").IsNull(),
		)),
	).MustBuild(ctx)

	ctx, span := spanWithSelectQuery(ctx, r.tracer, "(*RoadmapRepository.CountAccountOnProgressRoadmaps)", query)
	defer span.End()

	var count uint64
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		span.SetStatus(codes.Error, "failed to count on progress roadmaps by account ID")
		span.RecordError(err)
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

		input.ID = roadmap.ID

		if err = r.saveInitialProgression(traceCtx, tx, input); err != nil {
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

func (r *RoadmapRepository) saveInitialProgression(ctx context.Context, tx pgx.Tx, input *domain.Roadmap) error {
	query, args := psql.Insert(
		im.Into(domain.RoadmapProgressionTable, "account_id", "roadmap_id", "total_topics", "total_finished_topics", "created_at", "updated_at"),
		im.Values(psql.Arg(input.AccountID, input.ID, input.TotalTopics, 0, input.CreatedAt, input.UpdatedAt)),
	).MustBuild(ctx)

	ctx, span := spanWithInsertQuery(ctx, r.tracer, "(*RoadmapRepository.saveInitialProgression)", query)
	defer span.End()

	_, err := tx.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (r *RoadmapRepository) saveTopicsAndSubtopics(ctx context.Context, tx pgx.Tx, roadmapID int, topics []*domain.Topic) error {
	// subTopicMap with topic's slug as the key to its subtopics
	subTopicMap := make(map[string][]*domain.Topic)

	// Insert the topics
	insertTopicMods := []bob.Mod[*dialect.InsertQuery]{
		im.Into(domain.TopicTable, "account_id", "roadmap_id", "title", "slug", "description", "pro_tips", "order", "external_search_query", "created_at", "updated_at"),
	}
	for _, topic := range topics {
		subTopicMap[topic.Slug] = topic.Subtopics
		arg := psql.Arg(topic.AccountID, roadmapID, topic.Title, topic.Slug, topic.Description, topic.ProTips, topic.Order, topic.ExternalSearchQuery, topic.CreatedAt, topic.UpdatedAt)
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

	linkedSubtopics := make([][]any, 0, len(mergedTopicAndSubtopic))

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
			item.Title, item.Slug, item.Description, item.ProTips, item.Order,
			item.ExternalSearchQuery, item.CreatedAt, item.UpdatedAt,
		})
	}
	log.Debug().Any("linkedSubtopics", linkedSubtopics).Send()

	// Store the subtopics
	_, err = tx.CopyFrom(ctx,
		pgx.Identifier{domain.TopicTable},
		[]string{"account_id", "roadmap_id", "parent_id",
			"title", "slug", "description", "pro_tips", "order",
			"external_search_query", "created_at", "updated_at"},
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
			sm.Columns(r.roadmapColumns(roadmapColumnsOptions{
				includeProgression:           false,
				includePersonalizationOption: false,
				includeAccount:               false,
			})...),
			sm.From(domain.RoadmapTable),
			sm.Where(psql.And(
				psql.Quote(domain.RoadmapTable, "slug").EQ(psql.Arg(slug)),
				psql.Quote(domain.RoadmapTable, "deleted_at").IsNull()),
			),
		).MustBuild(ctx)

		roadmaps, err := r.fetch(traceCtx, roadmapFetchConfig{
			query: fetchRoadmapQuery,
			args:  fetchRoadmapArgs,
		})
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

		updateRoadmapQueryBuilder := psql.Update(
			um.Table(domain.RoadmapTable),
		)
		if roadmap.IsDeleted() {
			cacher := cache.New[domain.Roadmap](r.cache)
			err := cacher.Delete(traceCtx, &cache.Key{Namespace: domain.RoadmapTable, Key: slug})
			if err != nil {
				// TODO: should retry or log this error
				log.Error().Err(err).Msg("failed to delete roadmap from cache")
			}
			updateRoadmapQueryBuilder.Apply(um.SetCol("deleted_at").ToArg(roadmap.DeletedAt))
		}
		updateRoadmapQueryBuilder.Apply(um.Where(psql.Quote(domain.RoadmapTable, "slug").EQ(psql.Arg(slug))))

		updateRoadmapQuery, updateRoadmapArgs := updateRoadmapQueryBuilder.MustBuild(ctx)
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

func (r *RoadmapRepository) UpdateByID(ctx context.Context, id int, updateFn func(roadmap *domain.Roadmap) (bool, error)) error {
	traceCtx, span := r.tracer.Start(ctx, "(*RoadmapRepository.Update)")
	defer span.End()

	err := r.db.InTx(ctx, func(tx pgx.Tx) error {
		fetchRoadmapQuery, fetchRoadmapArgs := psql.Select(
			sm.Columns(r.roadmapColumns(roadmapColumnsOptions{
				includeProgression:           false,
				includePersonalizationOption: false,
				includeAccount:               false,
			})...),
			sm.From(domain.RoadmapTable),
			sm.Where(psql.And(
				psql.Quote(domain.RoadmapTable, "id").EQ(psql.Arg(id)),
				psql.Quote(domain.RoadmapTable, "deleted_at").IsNull()),
			),
		).MustBuild(ctx)

		roadmaps, err := r.fetch(traceCtx, roadmapFetchConfig{
			query: fetchRoadmapQuery,
			args:  fetchRoadmapArgs,
		})
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

		updateRoadmapQueryBuilder := psql.Update(
			um.Table(domain.RoadmapTable),
		)
		if roadmap.IsDeleted() {
			cacher := cache.New[domain.Roadmap](r.cache)
			err := cacher.Delete(traceCtx, &cache.Key{Namespace: domain.RoadmapTable, Key: roadmap.Slug})
			if err != nil {
				// TODO: should retry or log this error
				log.Error().Err(err).Msg("failed to delete roadmap from cache")
			}
			updateRoadmapQueryBuilder.Apply(um.SetCol("deleted_at").ToArg(roadmap.DeletedAt))
		}
		updateRoadmapQueryBuilder.Apply(um.Where(psql.Quote(domain.RoadmapTable, "id").EQ(psql.Arg(id))))

		updateRoadmapQuery, updateRoadmapArgs := updateRoadmapQueryBuilder.MustBuild(ctx)
		_, updateSpan := spanWithUpdateQuery(traceCtx, r.tracer, "(*RoadmapRepository.Update)", updateRoadmapQuery)
		defer updateSpan.End()

		if _, err = tx.Exec(ctx, updateRoadmapQuery, updateRoadmapArgs...); err != nil {
			updateSpan.SetStatus(codes.Error, "failed to update roadmap")
			updateSpan.RecordError(err)
			return err
		}

		return nil
	})
	return err
}

func (r *RoadmapRepository) fetchTopicsByRoadmapID(ctx context.Context, roadmapID int, filter filter.Filters) ([]*domain.Topic, error) {
	query, args := psql.Select(
		sm.Columns(
			psql.Quote(domain.TopicTable, "id"),
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
			psql.Quote(domain.RoadmapTopicProgressionTable, "is_finished"),
		),
		sm.From(domain.TopicTable),
		sm.LeftJoin(domain.RoadmapTopicProgressionTable).On(
			psql.And(
				psql.Quote(domain.RoadmapTopicProgressionTable, "topic_id").EQ(psql.Quote(domain.TopicTable, "id")),
				psql.Quote(domain.RoadmapTopicProgressionTable, "account_id").EQ(psql.Arg(filter.AccountID)),
			)),
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

	allTopics := make(map[int]*domain.Topic)
	var parentTopics []*domain.Topic

	for rows.Next() {
		var topic domain.Topic
		var topicParentID pgtype.Int4
		var externalSearchQuery pgtype.Text
		var topicIsFinished pgtype.Bool
		err = rows.Scan(
			&topic.ID,
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
			&topicIsFinished,
		)
		if err != nil {
			return nil, err
		}

		if topicParentID.Valid {
			topic.ParentID = int(topicParentID.Int32)
		} else {
			topic.ParentID = 0
		}

		if topicIsFinished.Valid {
			topic.IsFinished = topicIsFinished.Bool
		} else {
			topic.IsFinished = false
		}

		if externalSearchQuery.Valid {
			topic.ExternalSearchQuery = externalSearchQuery.String
		} else {
			topic.ExternalSearchQuery = ""
		}

		topic.Subtopics = make([]*domain.Topic, 0)

		allTopics[topic.ID] = &topic
		if topic.ParentID == 0 {
			parentTopics = append(parentTopics, &topic)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Now link subtopics to their parent topics
	for _, topic := range allTopics {
		if topic.ParentID != 0 {
			if parent, exists := allTopics[topic.ParentID]; exists {
				parent.Subtopics = append(parent.Subtopics, topic)
			}
		}
	}

	return parentTopics, nil
}
