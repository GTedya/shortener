package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// DeleteURLS представляет интерфейс для удаления нескольких URL из базы данных.
type DeleteURLS interface {
	DeleteURLS(ctx context.Context, token string, shortURLS chan string) error
}

// DeleteURLS удаляет несколько URL из базы данных, используя указанный токен пользователя и канал URL-адресов.
func (db *db) DeleteURLS(ctx context.Context, token string, shortURLS chan string) error {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("transaction err: %w", err)
	}

	b := &pgx.Batch{}

	go func() {
		for {
			url, ok := <-shortURLS
			if !ok {
				if b.Len() != 0 {
					db.log.Debug("finish")
					batchResults := tx.SendBatch(ctx, b)
					er := batchResults.Close()
					if er != nil {
						db.log.Errorw("batch closing error", "error", er)
						return
					}
					er = tx.Commit(ctx)
					if er != nil {
						db.log.Error(fmt.Errorf("%s: %w", ErrCommitTransaction, err))
					}
				}
				break
			}
			sqlStatement := "UPDATE urls SET is_deleted = true WHERE short_url=$1 AND user_token=$2"
			b.Queue(sqlStatement, url, token)
			if b.Len() >= DeleteBuffer {
				batchResults := tx.SendBatch(ctx, b)
				er := batchResults.Close()
				if er != nil {
					db.log.Errorw("batch closing error", "error", er)
					return
				}
				b = &pgx.Batch{}
			}
		}
	}()

	return nil
}
