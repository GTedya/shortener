package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// SaveURLS предоставляет методы для сохранения URL в базе данных.
type SaveURLS interface {
	// SaveURL сохраняет URL в базе данных и возвращает количество измененных строк.
	SaveURL(ctx context.Context, token, id, shortID string) (int64, error)

	// Batch выполняет пакетное сохранение URL в базе данных.
	Batch(ctx context.Context, records map[string]string) error
}

// SaveURL сохраняет URL в базе данных и возвращает количество измененных строк.
func (db *db) SaveURL(ctx context.Context, token, id, shortID string) (int64, error) {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("transaction start error: %w", err)
	}

	var txCommitted bool
	var result pgconn.CommandTag

	defer func() {
		if !txCommitted {
			if txErr := tx.Rollback(ctx); txErr != nil {
				db.log.Error("transaction rollback error: ", txErr)
			}
		}
	}()

	db.log.Info(shortID, id, "123"+token)
	result, err = tx.Exec(ctx, "INSERT INTO urls (short_url, url, user_token) VALUES ($1, $2, $3)", shortID, id, token)
	if err != nil {
		return 0, fmt.Errorf("saving url execution error: %w", err)
	}

	// Commit the transaction if everything went well
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit transaction error: %w", err)
	}
	txCommitted = true

	rows := result.RowsAffected()
	return rows, nil
}

// Batch выполняет пакетное сохранение URL в базе данных.
func (db *db) Batch(ctx context.Context, records map[string]string) error {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("transaction error: %w", err)
	}

	var txCommitted bool

	defer func() {
		if !txCommitted {
			if txErr := tx.Rollback(ctx); txErr != nil {
				db.log.Error("transaction rollback error: ", txErr)
			}
		}
	}()

	b := &pgx.Batch{}

	for id, subID := range records {
		sqlStatement := "INSERT INTO urls (url,short_url) VALUES ($1, $2)"
		b.Queue(sqlStatement, id, subID)
	}
	err = tx.SendBatch(ctx, b).Close()
	if err != nil {
		return fmt.Errorf("SendBatch error: %w", err)
	}

	// Commit the transaction if everything went well
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction error: %w", err)
	}

	txCommitted = true

	return nil
}
