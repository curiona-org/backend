package repository

import (
	"context"

	"github.com/curiona-org/backend/internal/domain"
	"github.com/curiona-org/backend/pkg/database"
	"github.com/jackc/pgx/v5"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/stephenafamo/bob/dialect/psql/um"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type ProfileRepository struct {
	db     database.Connection
	tracer trace.Tracer
}

func NewPostgresProfileRepository(db database.Connection) *ProfileRepository {
	tracer := otel.Tracer("db:postgres:profiles")
	return &ProfileRepository{
		db:     db,
		tracer: tracer,
	}
}

func (r *ProfileRepository) profileColumns() []any {
	return []any{
		psql.Quote(domain.ProfileTable, "id"),
		psql.Quote(domain.ProfileTable, "name"),
		psql.Quote(domain.ProfileTable, "avatar"),
		psql.Quote(domain.ProfileTable, "created_at"),
		psql.Quote(domain.ProfileTable, "updated_at"),
	}
}

func (r *ProfileRepository) Update(ctx context.Context, id int, updateFn func(profile *domain.Profile) (bool, error)) error {
	traceCtx, span := r.tracer.Start(ctx, "(*ProfileRepository.Update)", trace.WithAttributes(
		attribute.Int("id", id),
	))
	defer span.End()

	err := r.db.InTx(ctx, func(tx pgx.Tx) error {
		fetchProfileQuery, fetchProfileArgs := psql.Select(
			sm.Columns(r.profileColumns()...),
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
		_, updateSpan := spanWithUpdateQuery(traceCtx, r.tracer, "(*ProfileRepository.Update)", updateProfileQuery)
		defer updateSpan.End()

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

func (r *ProfileRepository) fetch(ctx context.Context, query string, args ...any) ([]domain.Profile, error) {
	ctx, span := spanWithSelectQuery(ctx, r.tracer, "(*ProfileRepository.fetch)", query)
	defer span.End()

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
