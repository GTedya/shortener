package database

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"sync"

	"github.com/GTedya/shortener/internal/app/middlewares"
	"github.com/GTedya/shortener/internal/helpers"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"go.uber.org/zap"
)

type DB struct {
	pool *pgxpool.Pool
	log  *zap.SugaredLogger
}

var BaseURL = "http://localhost:8080/"

const ErrQuery = "query error: %w"
const ErrCommitTransaction = "transaction commit error"
const DeleteBuffer = 10

func NewDB(dsn string, logger *zap.SugaredLogger) (*DB, error) {
	if err := runMigrations(dsn); err != nil {
		return nil, fmt.Errorf("failed to run DB migrations: %w", err)
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create a connection pool: %w", err)
	}
	return &DB{pool: pool, log: logger}, nil
}

//go:embed migrations/*.sql
var migrationsDir embed.FS

func runMigrations(dsn string) error {
	d, err := iofs.New(migrationsDir, "migrations")
	if err != nil {
		return fmt.Errorf("failed to return an iofs driver: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dsn)
	if err != nil {
		return fmt.Errorf("failed to get a new migrate instance: %w", err)
	}
	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("failed to apply migrations to the DB: %w", err)
		}
	}
	return nil
}

func (db *DB) Close() {
	db.pool.Close()
}

func (db *DB) Ping(ctx context.Context) error {
	err := db.pool.Ping(ctx)
	if err != nil {
		return fmt.Errorf("ping error: %w", err)
	}
	return nil
}

func (db *DB) GetBasicURL(ctx context.Context, shortID string) (string, bool, error) {
	var url string
	var isDeleted bool

	err := db.pool.QueryRow(ctx, "SELECT url, is_deleted FROM urls WHERE short_url = $1", shortID).Scan(&url, &isDeleted)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, fmt.Errorf("record not found")
		}
		return "", false, fmt.Errorf(ErrQuery, err)
	}

	return url, isDeleted, nil
}

func (db *DB) GetShortURL(ctx context.Context, id string) (string, error) {
	var url string
	err := db.pool.QueryRow(ctx, "SELECT short_url FROM urls WHERE url = $1", id).Scan(&url)
	if err != nil {
		return "", fmt.Errorf(ErrQuery, err)
	}
	return url, nil
}

func (db *DB) SaveURL(ctx context.Context, id, shortID string) (int64, error) {
	tx, err := db.pool.Begin(ctx)
	var token sql.NullString
	if err != nil {
		return 0, fmt.Errorf("transaction start error: %w", err)
	}
	value, ok := ctx.Value(middlewares.ContextKey("token")).(string)
	if ok {
		token.String = value
	}

	defer func() {
		if err != nil {
			if txErr := tx.Rollback(ctx); txErr != nil {
				db.log.Error("transaction rollback error: ", txErr)
				return
			}
		}
		if txErr := tx.Commit(ctx); txErr != nil {
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

func (db *DB) Batch(ctx context.Context, records map[string]string) error {
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
		if txErr := tx.Commit(ctx); txErr != nil {
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

func (db *DB) UserURLS(ctx context.Context, token string) ([]helpers.UserURL, error) {
	rows, err := db.pool.Query(ctx, "SELECT short_url, url FROM urls WHERE user_token = $1 AND is_deleted=false", token)
	if err != nil {
		return nil, fmt.Errorf(ErrQuery, err)
	}
	var urls []helpers.UserURL

	for rows.Next() {
		var url helpers.UserURL
		if err = rows.Scan(&url.ShortURL, &url.OriginalURL); err != nil {
			return nil, fmt.Errorf("rows scan error: %w", err)
		}
		url.ShortURL = BaseURL + url.ShortURL
		urls = append(urls, url)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error in query all urls: %w", err)
	}

	defer rows.Close()

	return urls, nil
}

func (db *DB) DeleteURLS(ctx context.Context, wg *sync.WaitGroup, shortURLS chan string) error {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("transaction err: %w", err)
	}

	b := &pgx.Batch{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			url, ok := <-shortURLS
			if !ok {
				if b.Len() != 0 {
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
			sqlStatement := "UPDATE urls SET is_deleted = true WHERE short_url=$1"
			b.Queue(sqlStatement, url)
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

func (db *DB) IsUserURL(ctx context.Context, token string, shortURL string) (bool, error) {
	var isOwner bool
	err := db.pool.QueryRow(ctx, "select exists (select true from urls WHERE short_url = $1 AND user_token = $2);",
		shortURL, token).Scan(&isOwner)
	if err != nil {
		return false, fmt.Errorf(ErrQuery, err)
	}
	return isOwner, nil
}
