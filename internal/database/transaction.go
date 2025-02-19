package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

type TransactionFunc func(tx pgx.Tx) error

// InTx starts a new transaction then calls fn().
func (db *DB) InTx(ctx context.Context, fn TransactionFunc) error {
	conn, err := db.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquiring connection: %w", err)
	}
	defer conn.Release()

	log.Debug().Msg("starting transaction")
	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}

	if err = fn(tx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			log.Debug().Msg("rollback transaction")
			return fmt.Errorf("rolling back transaction: %w (original error: %w)", rbErr, err)
		}
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	log.Debug().Msg("transaction committed")
	return nil
}
