package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/roadmap-thesis/backend/internal/domain"
	"github.com/roadmap-thesis/backend/pkg/database"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/dm"
	"github.com/stephenafamo/bob/dialect/psql/im"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/stephenafamo/bob/dialect/psql/um"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type SessionRepository struct {
	db     database.Connection
	tracer trace.Tracer
}

func NewSessionRepository(db database.Connection) *SessionRepository {
	tracer := otel.Tracer("db:postgres:sessions")
	return &SessionRepository{
		db:     db,
		tracer: tracer,
	}
}

func (r *SessionRepository) GetByAccountID(ctx context.Context, accountID int) (domain.Session, error) {
	query, args := psql.Select(
		sm.Columns(
			psql.Quote(domain.SessionTable, "id"),
			psql.Quote(domain.SessionTable, "account_id"),
			psql.Quote(domain.SessionTable, "refresh_token"),
			psql.Quote(domain.SessionTable, "user_agent"),
			psql.Quote(domain.SessionTable, "client_ip"),
			psql.Quote(domain.SessionTable, "blocked"),
			psql.Quote(domain.SessionTable, "expires_at"),
			psql.Quote(domain.SessionTable, "created_at"),
		),
		sm.From(domain.SessionTable),
		sm.Where(psql.Quote(domain.SessionTable, "account_id").EQ(psql.Arg(accountID))),
	).MustBuild(ctx)

	accounts, err := r.fetch(ctx, query, args...)
	if err != nil {
		return domain.Session{}, err
	}

	if len(accounts) == 0 {
		return domain.Session{}, domain.ErrSessionNotFound
	}

	return accounts[0], nil
}

func (r *SessionRepository) fetch(ctx context.Context, query string, args ...any) ([]domain.Session, error) {
	ctx, span := spanWithQuery(ctx, r.tracer, "(*SessionRepository.fetch)", query)
	defer span.End()

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		span.SetStatus(codes.Error, "failed to fetch sessions")
		span.RecordError(err)
		return nil, err
	}
	defer rows.Close()

	var sessions []domain.Session
	for rows.Next() {
		var s domain.Session
		if err := rows.Scan(
			&s.ID,
			&s.AccountID,
			&s.RefreshToken,
			&s.UserAgent,
			&s.ClientIP,
			&s.Blocked,
			&s.ExpiresAt,
			&s.CreatedAt,
		); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}

	return sessions, nil
}

func (r *SessionRepository) Save(ctx context.Context, input *domain.Session) (domain.Session, error) {
	query, args := psql.Insert(
		im.Into(domain.SessionTable, "account_id", "refresh_token", "user_agent", "client_ip", "blocked", "expires_at"),
		im.Values(psql.Arg(input.AccountID, input.RefreshToken, input.UserAgent, input.ClientIP, input.Blocked, input.ExpiresAt)),
		im.Returning("id", "created_at"),
	).MustBuild(ctx)

	ctx, span := spanWithQuery(ctx, r.tracer, "(*SessionRepository.Save)", query)
	defer span.End()

	var id int
	var createdAt time.Time
	if err := r.db.QueryRow(ctx, query, args...).Scan(&id, &createdAt); err != nil {
		span.SetStatus(codes.Error, "failed to save session")
		span.RecordError(err)
		return domain.Session{}, err
	}

	input.ID = id
	input.CreatedAt = createdAt
	return *input, nil
}

func (r *SessionRepository) Delete(ctx context.Context, id int) error {
	query, args := psql.Delete(
		dm.From(domain.SessionTable),
		dm.Where(psql.Quote(domain.SessionTable, "id").EQ(psql.Arg(id))),
	).MustBuild(ctx)

	ctx, span := spanWithQuery(ctx, r.tracer, "(*SessionRepository.Delete)", query)
	defer span.End()

	commandTag, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		span.SetStatus(codes.Error, "failed to delete session")
		span.RecordError(err)
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return domain.ErrSessionNotFound
	}
	return nil
}

func (r *SessionRepository) RotateRefreshToken(ctx context.Context, refreshToken string, updateFn func(context.Context, *domain.Session) (bool, error)) error {
	traceCtx, span := r.tracer.Start(ctx, "(*SessionRepository.RotateRefreshToken)")
	defer span.End()

	err := r.db.InTx(ctx, func(tx pgx.Tx) error {
		fetchSessionQuery, fetchSessionArgs := psql.Select(
			sm.Columns(
				psql.Quote(domain.SessionTable, "id"),
				psql.Quote(domain.SessionTable, "account_id"),
				psql.Quote(domain.SessionTable, "refresh_token"),
				psql.Quote(domain.SessionTable, "user_agent"),
				psql.Quote(domain.SessionTable, "client_ip"),
				psql.Quote(domain.SessionTable, "blocked"),
				psql.Quote(domain.SessionTable, "expires_at"),
				psql.Quote(domain.SessionTable, "created_at"),
			),
			sm.From(domain.SessionTable),
			sm.Where(psql.Quote(domain.SessionTable, "refresh_token").EQ(psql.Arg(refreshToken))),
		).MustBuild(ctx)

		sessions, err := r.fetch(traceCtx, fetchSessionQuery, fetchSessionArgs...)
		if err != nil {
			return err
		}

		if len(sessions) == 0 {
			return domain.ErrSessionNotFound
		}

		session := sessions[0]

		updated, err := updateFn(traceCtx, &session)
		if err != nil {
			return err
		}

		if !updated {
			return nil
		}

		query, args := psql.Update(
			um.Table(domain.SessionTable),
			um.SetCol("account_id").ToArg(session.AccountID),
			um.SetCol("user_agent").ToArg(session.UserAgent),
			um.SetCol("client_ip").ToArg(session.ClientIP),
			um.SetCol("blocked").ToArg(session.Blocked),
			um.SetCol("expires_at").ToArg(session.ExpiresAt),
			um.Where(psql.Quote(domain.SessionTable, "refresh_token").EQ(psql.Arg(refreshToken))),
		).MustBuild(ctx)

		ctx, span := spanWithQuery(traceCtx, r.tracer, "(*SessionRepository.RotateRefreshToken)", query)
		defer span.End()

		if _, err := tx.Exec(ctx, query, args...); err != nil {
			span.SetStatus(codes.Error, "failed to update session")
			span.RecordError(err)
			return err
		}

		return nil
	})
	return err
}
