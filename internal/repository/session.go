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
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type SessionRepository struct {
	db database.Connection
}

func NewSessionRepository(db database.Connection) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) GetByAccountID(ctx context.Context, accountID int) (domain.Session, error) {
	ctx, span := tracer.Start(ctx, "(*SessionRepository.GetByAccountID)", trace.WithAttributes(attribute.Int("account_id", accountID)))
	defer span.End()

	accounts, err := r.fetch(ctx, "account_id", accountID)
	if err != nil {
		return domain.Session{}, err
	}

	if len(accounts) == 0 {
		return domain.Session{}, domain.ErrSessionNotFound
	}

	return accounts[0], nil
}

func (r *SessionRepository) fetch(ctx context.Context, col string, args ...any) ([]domain.Session, error) {
	ctx, span := tracer.Start(ctx, "(*SessionRepository.fetch)", trace.WithAttributes(attribute.String("col", col)))
	defer span.End()

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
		sm.Where(psql.Quote(domain.SessionTable, col).EQ(psql.Arg(args...))),
	).MustBuild(ctx)

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
	ctx, span := tracer.Start(ctx, "(*SessionRepository.Save)", trace.WithAttributes(attribute.Int("account_id", input.AccountID)))
	defer span.End()

	query, args := psql.Insert(
		im.Into(domain.SessionTable, "account_id", "refresh_token", "user_agent", "client_ip", "blocked", "expires_at"),
		im.Values(psql.Arg(input.AccountID, input.RefreshToken, input.UserAgent, input.ClientIP, input.Blocked, input.ExpiresAt)),
		im.Returning("id", "created_at"),
	).MustBuild(ctx)

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
	ctx, span := tracer.Start(ctx, "(*SessionRepository.Delete)", trace.WithAttributes(attribute.Int("id", id)))
	defer span.End()

	query, args := psql.Delete(
		dm.From(domain.SessionTable),
		dm.Where(psql.Quote(domain.SessionTable, "id").EQ(psql.Arg(id))),
	).MustBuild(ctx)

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

func (r *SessionRepository) UpdateByRefreshToken(ctx context.Context, refreshToken string, updateFn func(context.Context, *domain.Session) (bool, error)) error {
	traceCtx, span := tracer.Start(ctx, "(*SessionRepository.UpdateByRefreshToken)")
	defer span.End()

	err := r.db.InTx(ctx, func(tx pgx.Tx) error {
		sessions, err := r.fetch(traceCtx, "refresh_token", refreshToken)
		if err != nil {
			span.SetStatus(codes.Error, "failed to fetch session")
			span.RecordError(err)
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

		if _, err := tx.Exec(ctx, query, args...); err != nil {
			span.SetStatus(codes.Error, "failed to update session")
			span.RecordError(err)
			return err
		}

		return nil
	})
	return err
}
