package repository

import (
	"context"

	"github.com/curiona-org/backend/internal/domain"
	"github.com/curiona-org/backend/pkg/database"
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

// Currently unused.
func NewPostgresPersonalizationOptionsRepository(db database.Connection) *PersonalizationOptionsRepository {
	tracer := otel.Tracer("db:postgres:personalization_options")
	return &PersonalizationOptionsRepository{
		db:     db,
		tracer: tracer,
	}
}

func (r *PersonalizationOptionsRepository) columns() []any {
	return []any{
		psql.Quote(domain.PersonalizationOptionsTable, "id"),
		psql.Quote(domain.PersonalizationOptionsTable, "account_id"),
		psql.Quote(domain.PersonalizationOptionsTable, "roadmap_id"),
		psql.Quote(domain.PersonalizationOptionsTable, "daily_time_availability"),
		psql.Quote(domain.PersonalizationOptionsTable, "total_duration"),
		psql.Quote(domain.PersonalizationOptionsTable, "skill_level"),
		psql.Quote(domain.PersonalizationOptionsTable, "additional_info"),
		psql.Quote(domain.PersonalizationOptionsTable, "created_at"),
		psql.Quote(domain.PersonalizationOptionsTable, "updated_at")}
}

func (r *PersonalizationOptionsRepository) fetch(
	ctx context.Context,
	query string,
	args ...any,
) ([]domain.PersonalizationOptions, error) {
	ctx, span := spanWithSelectQuery(ctx, r.tracer, "(*PersonalizationOptionsRepository.fetch)", query)
	defer span.End()

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		span.SetStatus(codes.Error, "failed to fetch roadmaps")
		span.RecordError(err)
		return nil, err
	}
	defer rows.Close()

	var personalizationOptions []domain.PersonalizationOptions
	for rows.Next() {
		var po domain.PersonalizationOptions
		if err := rows.Scan(&po.ID, &po.AccountID, &po.RoadmapID, &po.DailyTimeAvailability,
			&po.TotalDuration, &po.SkillLevel, &po.AdditionalInfo, &po.CreatedAt, &po.UpdatedAt); err != nil {
			span.SetStatus(codes.Error, "failed to scan row")
			span.RecordError(err)
			return nil, err
		}
		personalizationOptions = append(personalizationOptions, po)
	}
	if err := rows.Err(); err != nil {
		span.SetStatus(codes.Error, "error encountered during row iteration")
		span.RecordError(err)
		return nil, err
	}

	return personalizationOptions, nil
}

func (r *PersonalizationOptionsRepository) GetByRoadmapID(
	ctx context.Context,
	accountID, roadmapID int,
) (domain.PersonalizationOptions, error) {
	query, args := psql.Select(
		sm.Columns(r.columns()...),
		sm.From(psql.Quote(domain.PersonalizationOptionsTable)),
		sm.Where(psql.And(
			psql.Quote(domain.PersonalizationOptionsTable, "account_id").EQ(psql.Arg(accountID)),
			psql.Quote(domain.PersonalizationOptionsTable, "roadmap_id").EQ(psql.Arg(roadmapID)),
		)),
		sm.Limit(1),
	).MustBuild(ctx)

	pos, err := r.fetch(ctx, query, args...)
	if err != nil {
		return domain.PersonalizationOptions{}, err
	}

	if len(pos) == 0 {
		return domain.PersonalizationOptions{}, domain.ErrPersonalizationOptionsNotFound
	}

	return pos[0], nil
}
