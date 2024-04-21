package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type SaveURLS interface {
	SaveURL(ctx context.Context, token, id, shortID string) (int64, error)
	Batch(ctx context.Context, records map[string]string) error
}

func (db *db) SaveURL(ctx context.Context, token, id, shortID string) (int64, error) {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("transaction start error: %w", err)
	}

	defer func() {
		if err != nil {
			if txErr := tx.Rollback(ctx); txErr != nil {
				db.log.Error("transaction rollback error: ", txErr)
				return
			}
		}
		if txErr := tx.Commit(ctx); txErr != nil && !errors.Is(txErr, pgx.ErrTxClosed) {
			db.log.Errorw(ErrCommitTransaction, "error", txErr)
		}
	}()
	result, err := tx.Exec(ctx, "INSERT INTO urls (short_url, url, user_token) VALUES ($1, $2, $3)",
		shortID, id, token)
	if err != nil {
		return 0, fmt.Errorf("saving url execution error: %w", err)
	}

	rows := result.RowsAffected()
	return rows, nil
}

func (db *db) Batch(ctx context.Context, records map[string]string) error {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("transaction error: %w", err)
	}

	defer func() {
		if err != nil {
			if txErr := tx.Rollback(ctx); txErr != nil {
				db.log.Error("transaction rollback error: ", txErr)
				return
			}
		}
		if txErr := tx.Commit(ctx); txErr != nil && !errors.Is(txErr, pgx.ErrTxClosed) {
			db.log.Errorw(ErrCommitTransaction, "err", txErr)
		}
	}()

	b := &pgx.Batch{}

	for id, subID := range records {
		sqlStatement := "INSERT INTO urls (url,short_url) VALUES ($1, $2)"
		b.Queue(sqlStatement, id, subID)
	}

	batchResults := tx.SendBatch(ctx, b)
	err = batchResults.Close()
	if err != nil {
		return fmt.Errorf("batch closing error: %w", err)
	}

	return nil
}
