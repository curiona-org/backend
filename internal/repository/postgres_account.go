package repository

import (
	"context"

	"github.com/curiona-org/backend/internal/domain"
	"github.com/curiona-org/backend/pkg/database"
	"github.com/curiona-org/backend/pkg/pagination"
	"github.com/jackc/pgx/v5"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/im"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type AccountRepository struct {
	db     database.Connection
	tracer trace.Tracer
}

func NewPostgresAccountRepository(db database.Connection) *AccountRepository {
	tracer := otel.Tracer("db:postgres:accounts")
	return &AccountRepository{
		db:     db,
		tracer: tracer,
	}
}

func (r *AccountRepository) GetAll(ctx context.Context, pagination pagination.Paginator) ([]domain.Account, error) {
	query, args := psql.Select(
		sm.Columns(
			psql.Quote(domain.AccountTable, "id"),
			psql.Quote(domain.AccountTable, "provider"),
			psql.Quote(domain.AccountTable, "email"),
			psql.Quote(domain.AccountTable, "password"),
			psql.Quote(domain.AccountTable, "is_suspended"),
			psql.Quote(domain.AccountTable, "is_admin"),
			psql.Quote(domain.AccountTable, "created_at"),
			psql.Quote(domain.AccountTable, "updated_at"),
			psql.Quote(domain.ProfileTable, "id"),
			psql.Quote(domain.ProfileTable, "name"),
			psql.Quote(domain.ProfileTable, "avatar"),
			psql.Quote(domain.ProfileTable, "created_at"),
			psql.Quote(domain.ProfileTable, "updated_at"),
		),
		sm.From(domain.AccountTable),
		sm.LeftJoin(domain.ProfileTable).Using("id"),
		sm.OrderBy(psql.Quote(domain.ProfileTable, "created_at")).Desc(),
		sm.Offset(psql.Arg(pagination.Skip)),
		sm.Limit(psql.Arg(pagination.Limit)),
	).MustBuild(ctx)

	return r.fetch(ctx, query, args...)
}

func (r *AccountRepository) GetByID(ctx context.Context, id int) (domain.Account, error) {
	query, args := psql.Select(
		sm.Columns(
			psql.Quote(domain.AccountTable, "id"),
			psql.Quote(domain.AccountTable, "provider"),
			psql.Quote(domain.AccountTable, "email"),
			psql.Quote(domain.AccountTable, "password"),
			psql.Quote(domain.AccountTable, "is_suspended"),
			psql.Quote(domain.AccountTable, "is_admin"),
			psql.Quote(domain.AccountTable, "created_at"),
			psql.Quote(domain.AccountTable, "updated_at"),
			psql.Quote(domain.ProfileTable, "id"),
			psql.Quote(domain.ProfileTable, "name"),
			psql.Quote(domain.ProfileTable, "avatar"),
			psql.Quote(domain.ProfileTable, "created_at"),
			psql.Quote(domain.ProfileTable, "updated_at"),
		),
		sm.From(domain.AccountTable),
		sm.LeftJoin(domain.ProfileTable).Using("id"),
		sm.Where(psql.Quote("id").EQ(psql.Arg(id))),
	).MustBuild(ctx)

	accounts, err := r.fetch(ctx, query, args...)
	if err != nil {
		return domain.Account{}, err
	}

	if len(accounts) == 0 {
		return domain.Account{}, domain.ErrAccountNotFound
	}

	return accounts[0], nil
}

func (r *AccountRepository) GetByEmail(ctx context.Context, email string) (domain.Account, error) {
	query, args := psql.Select(
		sm.Columns(
			psql.Quote(domain.AccountTable, "id"),
			psql.Quote(domain.AccountTable, "provider"),
			psql.Quote(domain.AccountTable, "email"),
			psql.Quote(domain.AccountTable, "password"),
			psql.Quote(domain.AccountTable, "is_suspended"),
			psql.Quote(domain.AccountTable, "is_admin"),
			psql.Quote(domain.AccountTable, "created_at"),
			psql.Quote(domain.AccountTable, "updated_at"),
			psql.Quote(domain.ProfileTable, "id"),
			psql.Quote(domain.ProfileTable, "name"),
			psql.Quote(domain.ProfileTable, "avatar"),
			psql.Quote(domain.ProfileTable, "created_at"),
			psql.Quote(domain.ProfileTable, "updated_at"),
		),
		sm.From(domain.AccountTable),
		sm.LeftJoin(domain.ProfileTable).Using("id"),
		sm.Where(psql.Quote("email").EQ(psql.Arg(email))),
	).MustBuild(ctx)

	accounts, err := r.fetch(ctx, query, args...)
	if err != nil {
		return domain.Account{}, err
	}

	if len(accounts) == 0 {
		return domain.Account{}, domain.ErrAccountNotFound
	}

	return accounts[0], nil
}

func (r *AccountRepository) fetch(ctx context.Context, query string, args ...any) ([]domain.Account, error) {
	ctx, span := spanWithSelectQuery(ctx, r.tracer, "(*AccountRepository.fetch)", query)
	defer span.End()

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		span.SetStatus(codes.Error, "failed to fetch accounts")
		span.RecordError(err)
		return nil, err
	}
	defer rows.Close()

	var accounts []domain.Account
	for rows.Next() {
		var account domain.Account
		var profile domain.Profile
		err = rows.Scan(
			&account.ID,
			&account.Method,
			&account.Email,
			&account.PasswordDigest,
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
		if err != nil {
			return nil, err
		}

		account.SetProfile(&profile)
		accounts = append(accounts, account)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return accounts, nil
}

func (r *AccountRepository) Count(ctx context.Context) (uint64, error) {
	query, args := psql.Select(
		sm.Columns(psql.F("COUNT", "*")),
		sm.From(domain.AccountTable),
	).MustBuild(ctx)

	var count uint64
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *AccountRepository) Save(ctx context.Context, input *domain.Account) (domain.Account, error) {
	var account domain.Account
	var profile domain.Profile
	err := r.db.InTx(ctx, func(tx pgx.Tx) error {
		saveAccountQuery, saveAccountArgs := psql.Insert(
			im.Into(domain.AccountTable, "email", "password", "provider", "created_at", "updated_at"),
			im.Values(psql.Arg(input.Email, input.PasswordDigest, input.Method, input.CreatedAt, input.UpdatedAt)),
			im.Returning("id", "email", "created_at", "updated_at"),
		).MustBuild(ctx)

		traceCtx, span := spanWithInsertQuery(ctx, r.tracer, "(*AccountRepository.Save)", saveAccountQuery)
		defer span.End()

		err := tx.QueryRow(ctx, saveAccountQuery, saveAccountArgs...).Scan(
			&account.ID,
			&account.Email,
			&account.CreatedAt,
			&account.UpdatedAt,
		)
		if err != nil {
			span.RecordError(err)
			return err
		}

		saveProfileQuery, saveProfileArgs := psql.Insert(
			im.Into(domain.ProfileTable, "id", "name", "avatar", "created_at", "updated_at"),
			im.Values(psql.Arg(account.ID, input.Profile.Name, input.Profile.Avatar, input.CreatedAt, input.UpdatedAt)),
			im.Returning("name", "avatar"),
		).MustBuild(ctx)

		_, span = spanWithInsertQuery(traceCtx, r.tracer, "(*AccountRepository.Save)", saveProfileQuery)
		defer span.End()

		err = tx.QueryRow(ctx, saveProfileQuery, saveProfileArgs...).Scan(
			&profile.Name,
			&profile.Avatar,
		)
		if err != nil {
			span.SetStatus(codes.Error, "failed to save account")
			span.RecordError(err)
			return err
		}

		return nil
	})
	if err != nil {
		return domain.Account{}, err
	}

	account.SetProfile(&profile)
	return account, nil
}
