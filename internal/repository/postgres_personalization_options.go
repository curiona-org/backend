package repository

import (
	"context"

	"github.com/roadmap-thesis/backend/internal/domain"
	"github.com/roadmap-thesis/backend/pkg/database"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type PersonalizationOptionsRepository struct {
	db     database.Connection
	tracer trace.Tracer
}

func NewPostgresPersonalizationOptionsRepository(db database.Connection) *PersonalizationOptionsRepository {
	tracer := otel.Tracer("db:postgres:personalization_options")
	return &PersonalizationOptionsRepository{
		db:     db,
		tracer: tracer,
	}
}

func (r *PersonalizationOptionsRepository) GetByRoadmapID(ctx context.Context, roadmapID int) (domain.PersonalizationOptions, error) {
	query, args := psql.Select(
		sm.Columns(
			"id",
			"account_id",
			"roadmap_id",
			"daily_time_availability",
			"total_duration",
			"skill_level",
			"additional_info",
			"created_at",
			"updated_at",
		),
		sm.From(domain.PersonalizationOptionsTable),
		sm.Where(psql.Quote("roadmap_id").EQ(psql.Arg(roadmapID))),
	).MustBuild(ctx)

	personalizationOpts, err := r.fetch(ctx, query, args...)
	if err != nil {
		return domain.PersonalizationOptions{}, err
	}

	if len(personalizationOpts) == 0 {
		return domain.PersonalizationOptions{}, domain.ErrPersonalizationOptionsNotFound
	}

	return personalizationOpts[0], nil
}

func (r *PersonalizationOptionsRepository) fetch(ctx context.Context, query string, args ...any) ([]domain.PersonalizationOptions, error) {
	ctx, span := spanWithSelectQuery(ctx, r.tracer, "(*PersonalizationOptionsRepository.fetch)", query)
	defer span.End()

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		span.SetStatus(codes.Error, "failed to fetch personalization options")
		span.RecordError(err)
		return nil, err
	}
	defer rows.Close()

	var personalizationOpts []domain.PersonalizationOptions
	for rows.Next() {
		var personalizationOpt domain.PersonalizationOptions
		err = rows.Scan(
			&personalizationOpt.ID,
			&personalizationOpt.AccountID,
			&personalizationOpt.RoadmapID,
			&personalizationOpt.DailyTimeAvailability,
			&personalizationOpt.TotalDuration,
			&personalizationOpt.SkillLevel,
			&personalizationOpt.AdditionalInfo,
			&personalizationOpt.CreatedAt,
			&personalizationOpt.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		personalizationOpts = append(personalizationOpts, personalizationOpt)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return personalizationOpts, nil
}
