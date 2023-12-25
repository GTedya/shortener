package database

import (
	"context"
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	pool *pgxpool.Pool
}

func NewDB(dsn string) (*DB, error) {
	if err := runMigrations(dsn); err != nil {
		return nil, fmt.Errorf("failed to run DB migrations: %w", err)
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create a connection pool: %w", err)
	}
	return &DB{pool: pool}, nil
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

func (db *DB) GetBasicURL(shortID string) (string, error) {
	var url string
	err := db.pool.QueryRow(context.TODO(), "SELECT url FROM urls WHERE short_url = $1", shortID).Scan(&url)
	if err != nil {
		return "", fmt.Errorf("query error: %w", err)
	}
	return url, nil
}

func (db *DB) GetShortURL(id string) (string, error) {
	var url string
	err := db.pool.QueryRow(context.TODO(), "SELECT short_url FROM urls WHERE url = $1", id).Scan(&url)
	if err != nil {
		return "", fmt.Errorf("query error: %w", err)
	}
	return url, nil
}

func (db *DB) SaveURL(id, shortID string) (int64, error) {
	ctx := context.TODO()
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("transaction error: %w", err)
	}
	result, err := tx.Exec(ctx, "INSERT INTO urls (short_url, url) VALUES ($1, $2)", shortID, id)
	if err != nil {
		return 0, fmt.Errorf("saving url execution error: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return 0, fmt.Errorf("transaction commit error: %w", err)
	}
	rows := result.RowsAffected()
	return rows, nil
}
