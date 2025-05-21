package repository

import (
	"context"
	"time"

	"github.com/curiona-org/backend/internal/domain"
	"github.com/curiona-org/backend/internal/filter"
	"github.com/curiona-org/backend/pkg/cache"
	"github.com/curiona-org/backend/pkg/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/dm"
	"github.com/stephenafamo/bob/dialect/psql/im"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type BookmarkRepository struct {
	db     database.Connection
	tracer trace.Tracer
}

func NewPostgresBookmarkRepository(db database.Connection, cache *cache.Connection) *BookmarkRepository {
	tracer := otel.Tracer("db:postgres:bookmarks")
	return &BookmarkRepository{
		db:     db,
		tracer: tracer,
	}
}

func (r *BookmarkRepository) roadmapColumns(opt roadmapColumnsOptions) []any {
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
func (r *BookmarkRepository) fetchRoadmap(ctx context.Context, cfg roadmapFetchConfig) ([]domain.Roadmap, error) {
	ctx, span := spanWithSelectQuery(ctx, r.tracer, "(*BookmarkRepository.fetchRoadmap)", cfg.query)
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

		err = rows.Scan(dest...)
		if err != nil {
			return nil, err
		}

		if roadmapDeletedAt.Valid {
			roadmap.DeletedAt = roadmapDeletedAt.Time
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

		roadmap.IsBookmarked = true
		roadmaps = append(roadmaps, roadmap)
	}

	if err = rows.Err(); err != nil {
		span.SetStatus(codes.Error, "failed to fetch roadmaps")
		span.RecordError(err)
		return nil, err
	}

	return roadmaps, nil
}

func (r *BookmarkRepository) ListBookmarkedRoadmaps(ctx context.Context, filters filter.Filters) ([]domain.Roadmap, error) {
	ctx, span := spanWithSelectQuery(ctx, r.tracer, "(*BookmarkRepository.ListBookmarkedRoadmaps)", "")
	defer span.End()

	selectQuery := psql.Select(
		sm.Columns(r.roadmapColumns(roadmapColumnsOptions{
			includeProgression:           true,
			includePersonalizationOption: true,
			includeAccount:               true,
		})...),
		sm.From(domain.BookmarkTable),
		sm.LeftJoin(domain.RoadmapTable).OnEQ(
			psql.Quote(domain.BookmarkTable, "roadmap_id"),
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
	)

	if filters.AccountID != 0 {
		selectQuery.Apply(sm.LeftJoin(domain.RoadmapProgressionTable).On(
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

	selectQuery.Apply(
		sm.Where(psql.Quote(domain.BookmarkTable, "account_id").EQ(psql.Arg(filters.AccountID))),
		sm.Where(psql.Quote(domain.RoadmapTable, "deleted_at").IsNull()),
	)

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

	return r.fetchRoadmap(ctx, roadmapFetchConfig{
		query:                        query,
		args:                         args,
		includeProgression:           true,
		includePersonalizationOption: true,
		includeAccount:               true,
	})
}

func (r *BookmarkRepository) RoadmapIsBookmarked(ctx context.Context, accountID int, roadmapID int) (bool, error) {
	query, args := psql.Select(
		sm.Columns(psql.F("EXISTS",
			psql.Select(
				sm.Columns(1),
				sm.From(domain.BookmarkTable),
				sm.Where(psql.And(
					psql.Quote(domain.BookmarkTable, "account_id").EQ(psql.Arg(accountID)),
					psql.Quote(domain.BookmarkTable, "roadmap_id").EQ(psql.Arg(roadmapID)),
				)),
			),
		)),
	).MustBuild(ctx)

	ctx, span := spanWithSelectQuery(ctx, r.tracer, "(*BookmarkRepository.RoadmapIsBookmarked)", query)
	defer span.End()

	var exists bool
	err := r.db.QueryRow(ctx, query, args...).Scan(&exists)
	if err != nil {
		span.SetStatus(codes.Error, "failed to check if roadmap is bookmarked")
		span.RecordError(err)
		return false, err
	}

	if exists {
		return true, nil
	}

	span.SetStatus(codes.Error, "roadmap is not bookmarked")
	return false, nil
}

func (r *BookmarkRepository) Count(ctx context.Context, accountID int) (uint64, error) {
	query, args := psql.Select(
		sm.Columns(psql.F("COUNT", "*")),
		sm.From(domain.BookmarkTable),
		sm.Where(psql.Quote(domain.BookmarkTable, "account_id").EQ(psql.Arg(accountID))),
	).MustBuild(ctx)
	ctx, span := spanWithSelectQuery(ctx, r.tracer, "(*BookmarkRepository.Count)", query)
	defer span.End()

	var count uint64
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		span.SetStatus(codes.Error, "failed to count bookmarks")
		span.RecordError(err)
		return 0, err
	}

	return count, nil
}

func (r *BookmarkRepository) CountBySearching(ctx context.Context, accountID int, search string) (uint64, error) {
	query, args := psql.Select(
		sm.Columns(psql.F("COUNT", "*")),
		sm.From(domain.BookmarkTable),
		sm.LeftJoin(domain.RoadmapTable).OnEQ(
			psql.Quote(domain.BookmarkTable, "roadmap_id"),
			psql.Quote(domain.RoadmapTable, "id")),
		sm.LeftJoin(domain.ProfileTable).OnEQ(
			psql.Quote(domain.ProfileTable, "account_id"),
			psql.Quote(domain.RoadmapTable, "account_id")),
		sm.Where(psql.And(
			psql.Quote(domain.BookmarkTable, "account_id").EQ(psql.Arg(accountID)),
			psql.Or(
				psql.Quote(domain.RoadmapTable, "title").ILike(psql.Arg("%"+search+"%")),
				psql.Quote(domain.RoadmapTable, "description").ILike(psql.Arg("%"+search+"%")),
				psql.Quote(domain.ProfileTable, "name").ILike(psql.Arg("%"+search+"%")),
			),
			psql.Quote(domain.RoadmapTable, "deleted_at").IsNull(),
		)),
	).MustBuild(ctx)

	ctx, span := spanWithSelectQuery(ctx, r.tracer, "(*BookmarkRepository.CountBySearching)", query)
	defer span.End()

	var count uint64
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		span.SetStatus(codes.Error, "failed to count bookmarks by searching")
		span.RecordError(err)
		return 0, err
	}

	return count, nil
}

func (r *BookmarkRepository) Save(ctx context.Context, accountID int, slug string) error {
	findRoadmapQuery, args := psql.Select(
		sm.Columns(r.roadmapColumns(roadmapColumnsOptions{
			includeProgression:           false,
			includePersonalizationOption: false,
			includeAccount:               false,
		})...),
		sm.From(domain.RoadmapTable),
		sm.Where(psql.And(
			psql.Quote(domain.RoadmapTable, "slug").EQ(psql.Arg(slug)),
			psql.Quote(domain.RoadmapTable, "deleted_at").IsNull(),
		)),
	).MustBuild(ctx)

	ctx, span := spanWithSelectQuery(ctx, r.tracer, "(*BookmarkRepository.Save)", findRoadmapQuery)
	defer span.End()

	roadmaps, err := r.fetchRoadmap(ctx, roadmapFetchConfig{
		query:                        findRoadmapQuery,
		args:                         args,
		includeProgression:           false,
		includePersonalizationOption: false,
		includeAccount:               false,
	})
	if err != nil {
		span.SetStatus(codes.Error, "failed to fetch roadmap")
		span.RecordError(err)
		return err
	}
	if len(roadmaps) == 0 {
		span.SetStatus(codes.Error, "roadmap not found")
		return domain.ErrRoadmapNotFound
	}

	roadmap := roadmaps[0]
	if roadmap.IsZero() {
		span.SetStatus(codes.Error, "roadmap not found")
		return domain.ErrRoadmapNotFound
	}

	query, args := psql.Insert(
		im.Into(domain.BookmarkTable, "account_id", "roadmap_id", "created_at"),
		im.Values(psql.Arg(accountID, roadmap.ID, time.Now())),
		im.OnConflict().DoNothing(),
	).MustBuild(ctx)

	ctx, span = spanWithInsertQuery(ctx, r.tracer, "(*BookmarkRepository.Save)", query)
	defer span.End()

	err = r.db.InTx(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, query, args...)
		if err != nil {
			span.SetStatus(codes.Error, "failed to insert bookmark")
			span.RecordError(err)
			return err
		}
		return nil
	})
	if err != nil {
		span.SetStatus(codes.Error, "failed to insert bookmark")
		span.RecordError(err)
		return err
	}
	return nil
}

func (r *BookmarkRepository) Delete(ctx context.Context, accountID int, slug string) error {
	findRoadmapQuery, args := psql.Select(
		sm.Columns(r.roadmapColumns(roadmapColumnsOptions{
			includeProgression:           false,
			includePersonalizationOption: false,
			includeAccount:               false,
		})...),
		sm.From(domain.RoadmapTable),
		sm.Where(psql.And(
			psql.Quote(domain.RoadmapTable, "slug").EQ(psql.Arg(slug)),
			psql.Quote(domain.RoadmapTable, "deleted_at").IsNull(),
		)),
	).MustBuild(ctx)

	ctx, span := spanWithSelectQuery(ctx, r.tracer, "(*BookmarkRepository.Delete)", findRoadmapQuery)
	defer span.End()

	roadmaps, err := r.fetchRoadmap(ctx, roadmapFetchConfig{
		query:                        findRoadmapQuery,
		args:                         args,
		includeProgression:           false,
		includePersonalizationOption: false,
		includeAccount:               false,
	})
	if err != nil {
		span.SetStatus(codes.Error, "failed to fetch roadmap")
		span.RecordError(err)
		return err
	}
	if len(roadmaps) == 0 {
		span.SetStatus(codes.Error, "roadmap not found")
		return domain.ErrRoadmapNotFound
	}

	query, args := psql.Delete(
		dm.From(domain.BookmarkTable),
		dm.Where(psql.And(
			psql.Quote(domain.BookmarkTable, "account_id").EQ(psql.Arg(accountID)),
			psql.Quote(domain.BookmarkTable, "roadmap_id").EQ(psql.Arg(roadmaps[0].ID)),
		)),
	).MustBuild(ctx)

	ctx, span = spanWithDeleteQuery(ctx, r.tracer, "(*BookmarkRepository.Delete)", query)
	defer span.End()

	err = r.db.InTx(ctx, func(tx pgx.Tx) error {
		result, err := tx.Exec(ctx, query, args...)
		if err != nil {
			return err
		}
		if result.RowsAffected() == 0 {
			return domain.ErrBookmarkNotFound
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
