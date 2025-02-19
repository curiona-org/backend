package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/roadmap-thesis/backend/internal/apperrors"
	"github.com/roadmap-thesis/backend/internal/database"
	"github.com/roadmap-thesis/backend/internal/domain"
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

func NewPostgresSessionRepository(db database.Connection) *SessionRepository {
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
	ctx, span := spanWithSelectQuery(ctx, r.tracer, "(*SessionRepository.fetch)", query)
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
		if err = rows.Scan(
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

	ctx, span := spanWithInsertQuery(ctx, r.tracer, "(*SessionRepository.Save)", query)
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

	ctx, span := spanWithDeleteQuery(ctx, r.tracer, "(*SessionRepository.Delete)", query)
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

// Renew renews a session by updating the existing session and creating a new one.
// It performs the following steps:
//  1. Starts a new transaction.
//  2. Fetches the session associated with the given refresh token.
//  3. Calls updateFn() to update the session details.
//  4. If the session is blocked, deletes the session.
//  5. If the session was updated, marks the old session as blocked and creates a new session with the updated details.
func (r *SessionRepository) Renew(ctx context.Context, refreshToken string, updateFn func(*domain.Session) (bool, error)) error {
	traceCtx, span := r.tracer.Start(ctx, "(*SessionRepository.Renew)")
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

		updated, err := updateFn(&session)
		if err != nil && errors.Is(apperrors.Unwrap(err), domain.ErrSessionIsBlocked) {
			query, args := psql.Delete(
				dm.From(domain.SessionTable),
				dm.Where(psql.Quote(domain.SessionTable, "id").EQ(psql.Arg(session.ID))),
			).MustBuild(ctx)

			_, span = spanWithDeleteQuery(traceCtx, r.tracer, "(*SessionRepository.Renew)", query)
			defer span.End()

			commandTag, execErr := r.db.Exec(ctx, query, args...)
			if execErr != nil {
				span.SetStatus(codes.Error, "failed to delete session")
				span.RecordError(execErr)
				return execErr
			}
			if commandTag.RowsAffected() == 0 {
				return err
			}

			return err
		} else if err != nil {
			return err
		}

		if !updated {
			return nil
		}

		updateOldSessionQuery, updateOldSessionArgs := psql.Update(
			um.Table(domain.SessionTable),
			um.SetCol("blocked").ToArg(true),
			um.Where(psql.Quote(domain.SessionTable, "refresh_token").EQ(psql.Arg(refreshToken))),
		).MustBuild(ctx)
		updateTraceCtx, updateSpan := spanWithUpdateQuery(traceCtx, r.tracer, "(*SessionRepository.Renew)", updateOldSessionQuery)
		defer updateSpan.End()

		if _, err = tx.Exec(ctx, updateOldSessionQuery, updateOldSessionArgs...); err != nil {
			updateSpan.SetStatus(codes.Error, "failed to update session")
			updateSpan.RecordError(err)
			return err
		}

		newSessionQuery, newSessionArgs := psql.Insert(
			im.Into(domain.SessionTable, "account_id", "refresh_token", "user_agent", "client_ip", "blocked", "expires_at"),
			im.Values(psql.Arg(session.AccountID, session.RefreshToken, session.UserAgent, session.ClientIP, session.Blocked, session.ExpiresAt)),
		).MustBuild(ctx)
		_, insertSpan := spanWithInsertQuery(updateTraceCtx, r.tracer, "(*SessionRepository.Renew)", newSessionQuery)
		defer insertSpan.End()

		if _, err = tx.Exec(ctx, newSessionQuery, newSessionArgs...); err != nil {
			insertSpan.SetStatus(codes.Error, "failed to update session")
			insertSpan.RecordError(err)
			return err
		}

		return nil
	})
	return err
}
