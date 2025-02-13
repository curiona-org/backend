package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/roadmap-thesis/backend/internal/domain"
	"github.com/roadmap-thesis/backend/pkg/database"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/im"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type accountRepository struct {
	db     database.Connection
	tracer trace.Tracer
}

var _ domain.AccountRepository = (*accountRepository)(nil)

func NewAccountRepository(db database.Connection) domain.AccountRepository {
	tracer := otel.Tracer("db:postgres:accounts")
	return &accountRepository{
		db:     db,
		tracer: tracer,
	}
}

func (r *accountRepository) GetByID(ctx context.Context, id int) (domain.Account, error) {
	query, args := psql.Select(
		sm.Columns(
			psql.Quote(domain.AccountTable, "id"),
			psql.Quote(domain.AccountTable, "email"),
			psql.Quote(domain.AccountTable, "password"),
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

func (r *accountRepository) GetByEmail(ctx context.Context, email string) (domain.Account, error) {
	query, args := psql.Select(
		sm.Columns(
			psql.Quote(domain.AccountTable, "id"),
			psql.Quote(domain.AccountTable, "email"),
			psql.Quote(domain.AccountTable, "password"),
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

func (r *accountRepository) fetch(ctx context.Context, query string, args ...any) ([]domain.Account, error) {
	ctx, span := spanWithQuery(ctx, r.tracer, "(*accountRepository.fetch)", query)
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
		err := rows.Scan(
			&account.ID,
			&account.Email,
			&account.Password,
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

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return accounts, nil
}

func (r *accountRepository) Save(ctx context.Context, input *domain.Account) (domain.Account, error) {
	var account domain.Account
	var profile domain.Profile
	err := r.db.InTx(ctx, func(tx pgx.Tx) error {
		saveAccountQuery, saveAccountArgs := psql.Insert(
			im.Into(domain.AccountTable, "email", "password", "created_at", "updated_at"),
			im.Values(psql.Arg(input.Email, input.Password, input.CreatedAt, input.UpdatedAt)),
			im.Returning("id", "email", "created_at", "updated_at"),
		).MustBuild(ctx)

		ctx, span := spanWithQuery(ctx, r.tracer, "(*accountRepository.Save)", saveAccountQuery)
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

		ctx, span = spanWithQuery(ctx, r.tracer, "(*accountRepository.Save)", saveProfileQuery)
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
