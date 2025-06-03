package repository

import (
	"context"

	"github.com/curiona-org/backend/internal/domain"
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

func (r *RatingRepository) ratingColumns() []any {
	return []any{
		psql.Quote(domain.RatingTable, "roadmap_id"),
		psql.Quote(domain.RatingTable, "account_id"),
		psql.Quote(domain.RatingTable, "progression_total_topics"),
		psql.Quote(domain.RatingTable, "progression_total_finished_topics"),
		psql.Quote(domain.RatingTable, "rating"),
		psql.Quote(domain.RatingTable, "comment"),
		psql.Quote(domain.RatingTable, "created_at"),
		psql.Quote(domain.RatingTable, "updated_at"),
	}
}

type ratingFetchConfig struct {
	query string
	args  []any
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
		if err := rows.Scan(
			&rating.RoadmapID,
			&rating.AccountID,
			&rating.ProgressionTotalTopics,
			&rating.ProgressionTotalFinishedTopics,
			&rating.Rating,
			&rating.Comment,
			&rating.CreatedAt,
			&rating.UpdatedAt,
		); err != nil {
			span.SetStatus(codes.Error, "failed to scan rating row")
			span.RecordError(err)
			return nil, err
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
		sm.Columns(r.ratingColumns()...),
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
		query: query,
		args:  args,
	})
	if err != nil {
		return domain.Rating{}, err
	}

	if len(ratings) == 0 {
		return domain.Rating{}, domain.ErrRatingNotFound
	}

	return ratings[0], nil
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
