package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/roadmap-thesis/backend/internal/domain"
	"github.com/roadmap-thesis/backend/pkg/database"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/im"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type AccountRepository struct {
	db database.Connection
}

func NewAccountRepository(db database.Connection) *AccountRepository {
	return &AccountRepository{
		db: db,
	}
}

func (r *AccountRepository) GetByID(ctx context.Context, filter int) (domain.Account, error) {
	ctx, span := tracer.Start(ctx, "(*AccountRepository.GetByID)", trace.WithAttributes(attribute.Int("id", filter)))
	defer span.End()

	accounts, err := r.fetch(ctx, "id", filter)
	if err != nil {
		return domain.Account{}, err
	}

	if len(accounts) == 0 {
		return domain.Account{}, domain.ErrAccountNotFound
	}

	return accounts[0], nil
}

func (r *AccountRepository) GetByEmail(ctx context.Context, filter string) (domain.Account, error) {
	ctx, span := tracer.Start(ctx, "(*AccountRepository.GetByEmail)", trace.WithAttributes(attribute.String("email", filter)))
	defer span.End()

	accounts, err := r.fetch(ctx, "email", filter)
	if err != nil {
		return domain.Account{}, err
	}

	if len(accounts) == 0 {
		return domain.Account{}, domain.ErrAccountNotFound
	}

	return accounts[0], nil
}

func (r *AccountRepository) fetch(ctx context.Context, col string, args ...any) ([]domain.Account, error) {
	ctx, span := tracer.Start(ctx, "(*AccountRepository.fetch)", trace.WithAttributes(attribute.String("col", col)))
	defer span.End()

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
		sm.Where(psql.Quote(col).EQ(psql.Arg(args...))),
	).MustBuild(ctx)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
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
			span.RecordError(err)
			return nil, err
		}

		account.SetProfile(&profile)
		accounts = append(accounts, account)
	}

	if err := rows.Err(); err != nil {
		span.RecordError(err)
		return nil, err
	}

	return accounts, nil
}

func (r *AccountRepository) Save(ctx context.Context, input *domain.Account) (domain.Account, error) {
	ctx, span := tracer.Start(ctx, "(*AccountRepository.Save)")
	defer span.End()

	var account domain.Account
	var profile domain.Profile
	err := r.db.InTx(ctx, func(tx pgx.Tx) error {
		saveAccountQuery, saveAccountArgs := psql.Insert(
			im.Into(domain.AccountTable, "email", "password", "created_at", "updated_at"),
			im.Values(psql.Arg(input.Email, input.Password, input.CreatedAt, input.UpdatedAt)),
			im.Returning("id", "email", "created_at", "updated_at"),
		).MustBuild(ctx)

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

		err = tx.QueryRow(ctx, saveProfileQuery, saveProfileArgs...).Scan(
			&profile.Name,
			&profile.Avatar,
		)
		if err != nil {
			span.RecordError(err)
			return err
		}

		return nil
	})
	if err != nil {
		span.RecordError(err)
		return domain.Account{}, err
	}

	account.SetProfile(&profile)
	return account, nil
}
