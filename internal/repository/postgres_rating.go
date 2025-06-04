package repository

import (
	"context"

	"github.com/curiona-org/backend/internal/domain"
	"github.com/curiona-org/backend/internal/filter"
	"github.com/curiona-org/backend/pkg/database"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/im"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type RatingRepository struct {
	db     database.Connection
	tracer trace.Tracer
}

func NewPostgresRatingRepository(db database.Connection) *RatingRepository {
	tracer := otel.Tracer("db:postgres:roadmaps")
	return &RatingRepository{
		db:     db,
		tracer: tracer,
	}
}

type ratingColumnsOptions struct {
	includeAccount bool
}

func (r *RatingRepository) ratingColumns(opt ratingColumnsOptions) []any {
	cols := []any{
		psql.Quote(domain.RatingTable, "roadmap_id"),
		psql.Quote(domain.RatingTable, "account_id"),
		psql.Quote(domain.RatingTable, "progression_total_topics"),
		psql.Quote(domain.RatingTable, "progression_total_finished_topics"),
		psql.Quote(domain.RatingTable, "rating"),
		psql.Quote(domain.RatingTable, "comment"),
		psql.Quote(domain.RatingTable, "created_at"),
		psql.Quote(domain.RatingTable, "updated_at"),
	}

	if opt.includeAccount {
		cols = append(cols,
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
			psql.Quote(domain.ProfileTable, "updated_at"))
	}

	return cols
}

type ratingFetchConfig struct {
	query          string
	args           []any
	includeAccount bool
}

func (r *RatingRepository) fetch(ctx context.Context, cfg ratingFetchConfig) ([]domain.Rating, error) {
	ctx, span := spanWithSelectQuery(ctx, r.tracer, "(*RatingRepository.fetch)", cfg.query)
	defer span.End()

	rows, err := r.db.Query(ctx, cfg.query, cfg.args...)
	if err != nil {
		span.SetStatus(codes.Error, "failed to fetch ratings")
		span.RecordError(err)
		return nil, err
	}
	defer rows.Close()

	var ratings []domain.Rating
	for rows.Next() {
		var rating domain.Rating
		dest := []any{
			&rating.RoadmapID,
			&rating.AccountID,
			&rating.ProgressionTotalTopics,
			&rating.ProgressionTotalFinishedTopics,
			&rating.Rating,
			&rating.Comment,
			&rating.CreatedAt,
			&rating.UpdatedAt}

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
				&profile.UpdatedAt)
		}

		if err := rows.Scan(dest...); err != nil {
			span.SetStatus(codes.Error, "failed to scan rating row")
			span.RecordError(err)
			return nil, err
		}

		if cfg.includeAccount {
			account.SetProfile(&profile)
			rating.SetAccount(&account)
		}

		ratings = append(ratings, rating)
	}
	if err := rows.Err(); err != nil {
		span.SetStatus(codes.Error, "error encountered while iterating over rows")
		span.RecordError(err)
		return nil, err
	}

	return ratings, nil
}

func (r *RatingRepository) GetRoadmapRatingByAccountID(ctx context.Context, accountID int, slug string) (domain.Rating, error) {
	query, args := psql.Select(
		sm.Columns(r.ratingColumns(ratingColumnsOptions{
			includeAccount: false,
		})...),
		sm.From(domain.RatingTable),
		sm.LeftJoin(domain.RoadmapTable).OnEQ(
			psql.Quote(domain.RoadmapTable, "id"),
			psql.Quote(domain.RatingTable, "roadmap_id")),
		sm.Where(psql.And(
			psql.Quote(domain.RatingTable, "account_id").EQ(psql.Arg(accountID)),
			psql.Quote(domain.RoadmapTable, "slug").EQ(psql.Arg(slug)),
			psql.Quote(domain.RoadmapTable, "deleted_at").IsNull()),
		),
	).MustBuild(ctx)

	ctx, span := spanWithSelectQuery(ctx, r.tracer, "(*RatingRepository.GetRoadmapRatingByAccountID)", query)
	defer span.End()

	ratings, err := r.fetch(ctx, ratingFetchConfig{
		query:          query,
		args:           args,
		includeAccount: false,
	})
	if err != nil {
		return domain.Rating{}, err
	}

	if len(ratings) == 0 {
		return domain.Rating{}, domain.ErrRatingNotFound
	}

	return ratings[0], nil
}

func (r *RatingRepository) GetRoadmapRatings(ctx context.Context, filters filter.Filters) ([]domain.Rating, error) {
	selectQuery := psql.Select(
		sm.Columns(r.ratingColumns(ratingColumnsOptions{
			includeAccount: true,
		})...),
		sm.From(domain.RatingTable),
		sm.LeftJoin(domain.RoadmapTable).OnEQ(
			psql.Quote(domain.RoadmapTable, "id"),
			psql.Quote(domain.RatingTable, "roadmap_id")),
		sm.LeftJoin(domain.AccountTable).OnEQ(
			psql.Quote(domain.AccountTable, "id"),
			psql.Quote(domain.RatingTable, "account_id")),
		sm.LeftJoin(domain.ProfileTable).OnEQ(
			psql.Quote(domain.ProfileTable, "id"),
			psql.Quote(domain.RatingTable, "account_id")),
	)

	if filters.Search != "" {
		selectQuery.Apply(
			sm.Where(psql.And(
				psql.Or(
					psql.Quote(domain.RatingTable, "comment").ILike(psql.Arg("%"+filters.Search+"%")),
					psql.Quote(domain.AccountTable, "email").ILike(psql.Arg("%"+filters.Search+"%")),
					psql.Quote(domain.ProfileTable, "name").ILike(psql.Arg("%"+filters.Search+"%")),
				),
				psql.Quote(domain.RoadmapTable, "id").EQ(psql.Arg(filters.ID)),
				psql.Quote(domain.RoadmapTable, "deleted_at").IsNull())))
	} else {
		selectQuery.Apply(
			sm.Where(psql.And(
				psql.Quote(domain.RoadmapTable, "id").EQ(psql.Arg(filters.ID)),
				psql.Quote(domain.RoadmapTable, "deleted_at").IsNull())))
	}

	if filters.OrderBy == filter.OrderByOldest {
		selectQuery.Apply(sm.OrderBy(psql.Quote(domain.RatingTable, "created_at")).Asc())
	} else {
		selectQuery.Apply(sm.OrderBy(psql.Quote(domain.RatingTable, "created_at")).Desc())
	}

	selectQuery.Apply(
		sm.Offset(psql.Arg(filters.Paginator.Skip)),
		sm.Limit(psql.Arg(filters.Paginator.Limit)),
	)

	query, args := selectQuery.MustBuild(ctx)

	return r.fetch(ctx, ratingFetchConfig{
		query:          query,
		args:           args,
		includeAccount: true,
	})
}

func (r *RatingRepository) AverageRating(ctx context.Context, roadmapID int) (float64, error) {
	query, args := psql.Select(
		sm.Columns(psql.F("AVG", psql.Quote(domain.RatingTable, "rating"))),
		sm.From(domain.RatingTable),
		sm.LeftJoin(domain.RoadmapTable).OnEQ(
			psql.Quote(domain.RoadmapTable, "id"),
			psql.Quote(domain.RatingTable, "roadmap_id")),
		sm.Where(psql.And(
			psql.Quote(domain.RoadmapTable, "id").EQ(psql.Arg(roadmapID)),
			psql.Quote(domain.RoadmapTable, "deleted_at").IsNull()),
		),
	).MustBuild(ctx)

	ctx, span := spanWithSelectQuery(ctx, r.tracer, "(*RatingRepository.AverageRating)", query)
	defer span.End()

	var average float64
	if err := r.db.QueryRow(ctx, query, args...).Scan(&average); err != nil {
		span.SetStatus(codes.Error, "failed to calculate average rating")
		span.RecordError(err)
		return 0, err
	}

	return average, nil
}

func (r *RatingRepository) CountRoadmapRatings(ctx context.Context, id int) (uint64, error) {
	query, args := psql.Select(
		sm.Columns(psql.F("COUNT", psql.Quote(domain.RatingTable, "roadmap_id"))),
		sm.From(domain.RatingTable),
		sm.LeftJoin(domain.RoadmapTable).OnEQ(
			psql.Quote(domain.RoadmapTable, "id"),
			psql.Quote(domain.RatingTable, "roadmap_id")),
		sm.Where(psql.And(
			psql.Quote(domain.RoadmapTable, "id").EQ(psql.Arg(id)),
			psql.Quote(domain.RoadmapTable, "deleted_at").IsNull()),
		),
	).MustBuild(ctx)

	ctx, span := spanWithSelectQuery(ctx, r.tracer, "(*RatingRepository.CountRoadmapRatings)", query)
	defer span.End()

	var total uint64
	if err := r.db.QueryRow(ctx, query, args...).Scan(&total); err != nil {
		span.SetStatus(codes.Error, "failed to count roadmap ratings")
		span.RecordError(err)
		return 0, err
	}

	return total, nil
}

func (r *RatingRepository) CountRoadmapRatingsBySearching(ctx context.Context, roadmapID int, search string) (uint64, error) {
	query, args := psql.Select(
		sm.Columns(psql.F("COUNT", psql.Quote(domain.RatingTable, "roadmap_id"))),
		sm.From(domain.RatingTable),
		sm.LeftJoin(domain.RoadmapTable).OnEQ(
			psql.Quote(domain.RoadmapTable, "id"),
			psql.Quote(domain.RatingTable, "roadmap_id")),
		sm.LeftJoin(domain.AccountTable).OnEQ(
			psql.Quote(domain.AccountTable, "id"),
			psql.Quote(domain.RatingTable, "account_id")),
		sm.LeftJoin(domain.ProfileTable).OnEQ(
			psql.Quote(domain.ProfileTable, "id"),
			psql.Quote(domain.RatingTable, "account_id")),
		sm.Where(psql.And(
			psql.Or(
				psql.Quote(domain.RatingTable, "comment").ILike(psql.Arg("%"+search+"%")),
				psql.Quote(domain.AccountTable, "email").ILike(psql.Arg("%"+search+"%")),
				psql.Quote(domain.ProfileTable, "name").ILike(psql.Arg("%"+search+"%")),
			),
			psql.Quote(domain.RoadmapTable, "id").EQ(psql.Arg(roadmapID)),
			psql.Quote(domain.RoadmapTable, "deleted_at").IsNull()),
		),
	).MustBuild(ctx)

	ctx, span := spanWithSelectQuery(ctx, r.tracer, "(*RatingRepository.CountRoadmapRatingsBySearching)", query)
	defer span.End()

	var total uint64
	if err := r.db.QueryRow(ctx, query, args...).Scan(&total); err != nil {
		span.SetStatus(codes.Error, "failed to count ratings by search")
		span.RecordError(err)
		return 0, err
	}

	return total, nil
}

func (r *RatingRepository) RateRoadmap(ctx context.Context, input *domain.Rating) error {
	query, args := psql.Insert(
		im.Into(domain.RatingTable,
			"roadmap_id", "account_id", "progression_total_topics", "progression_total_finished_topics",
			"rating", "comment", "created_at", "updated_at"),
		im.Values(psql.Arg(
			input.RoadmapID, input.AccountID, input.ProgressionTotalTopics, input.ProgressionTotalFinishedTopics,
			input.Rating, input.Comment, input.CreatedAt, input.UpdatedAt)),
		im.OnConflict("roadmap_id", "account_id").DoUpdate(
			im.SetCol("progression_total_topics").ToArg(input.ProgressionTotalTopics),
			im.SetCol("progression_total_finished_topics").ToArg(input.ProgressionTotalFinishedTopics),
			im.SetCol("rating").ToArg(input.Rating),
			im.SetCol("comment").ToArg(input.Comment),
			im.SetCol("updated_at").ToArg(input.UpdatedAt),
		),
	).MustBuild(ctx)

	ctx, span := spanWithInsertQuery(ctx, r.tracer, "(*RatingRepository.Save)", query)
	defer span.End()

	_, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		span.SetStatus(codes.Error, "failed to insert rating")
		span.RecordError(err)
		return err
	}

	return nil
}
