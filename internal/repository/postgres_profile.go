package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/roadmap-thesis/backend/internal/domain"
	"github.com/roadmap-thesis/backend/pkg/database"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/stephenafamo/bob/dialect/psql/um"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

type profileRepository struct {
	db     database.Connection
	tracer trace.Tracer
}

var _ domain.ProfileRepository = (*profileRepository)(nil)

func NewPostgresProfileRepository(db database.Connection) domain.ProfileRepository {
	tracer := otel.Tracer("db:postgres:profiles")
	return &profileRepository{
		db:     db,
		tracer: tracer,
	}
}

func (r *profileRepository) fetch(ctx context.Context, query string, args ...any) ([]domain.Profile, error) {
	ctx, span := spanWithQuery(ctx, r.tracer, "(*profileRepository.fetch)", query)
	defer span.End()
	span.SetAttributes(semconv.DBOperationKey.String("SELECT"))

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		span.SetStatus(codes.Error, "failed to fetch profiles")
		span.RecordError(err)
		return nil, err
	}
	defer rows.Close()

	var profiles []domain.Profile
	for rows.Next() {
		var s domain.Profile
		if err = rows.Scan(
			&s.ID,
			&s.Name,
			&s.Avatar,
			&s.CreatedAt,
			&s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		profiles = append(profiles, s)
	}

	return profiles, nil
}

func (r *profileRepository) Update(ctx context.Context, id int, updateFn func(profile *domain.Profile) (bool, error)) error {
	traceCtx, span := r.tracer.Start(ctx, "(*profileRepository.Update)")
	defer span.End()

	err := r.db.InTx(ctx, func(tx pgx.Tx) error {
		fetchProfileQuery, fetchProfileArgs := psql.Select(
			sm.Columns(
				psql.Quote(domain.ProfileTable, "id"),
				psql.Quote(domain.ProfileTable, "name"),
				psql.Quote(domain.ProfileTable, "avatar"),
				psql.Quote(domain.ProfileTable, "created_at"),
				psql.Quote(domain.ProfileTable, "updated_at"),
			),
			sm.From(domain.ProfileTable),
			sm.Where(psql.Quote("id").EQ(psql.Arg(id))),
		).MustBuild(ctx)

		profiles, err := r.fetch(ctx, fetchProfileQuery, fetchProfileArgs...)
		if err != nil {
			return err
		}

		if len(profiles) == 0 {
			return domain.ErrProfileNotFound
		}

		profile := profiles[0]
		updated, err := updateFn(&profile)
		if err != nil {
			return err
		}

		if !updated {
			return nil
		}

		updateProfileQuery, updateProfileArgs := psql.Update(
			um.Table(domain.ProfileTable),
			um.SetCol("name").ToArg(profile.Name),
			um.Where(psql.Quote(domain.ProfileTable, "id").EQ(psql.Arg(id))),
		).MustBuild(ctx)
		_, updateSpan := spanWithQuery(traceCtx, r.tracer, "(*profileRepository.Update)", updateProfileQuery)
		defer updateSpan.End()
		updateSpan.SetAttributes(semconv.DBOperationKey.String("UPDATE"))

		if _, err = tx.Exec(ctx, updateProfileQuery, updateProfileArgs...); err != nil {
			updateSpan.SetStatus(codes.Error, "failed to update profile")
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
