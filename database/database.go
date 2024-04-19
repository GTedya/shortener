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

	"go.uber.org/zap"
)

type db struct {
	pool *pgxpool.Pool
	log  *zap.SugaredLogger
}

type DB interface {
	Ping(ctx context.Context) error
	Close()
	GetURLS
	SaveURLS
	DeleteURLS
}

var BaseURL = "http://localhost:8080/"

const ErrQuery = "query error: %w"
const ErrCommitTransaction = "transaction commit error"
const DeleteBuffer = 10

func NewDB(dsn string, logger *zap.SugaredLogger) (DB, error) {
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create a connection pool: %w", err)
	}
	return &db{pool: pool, log: logger}, nil
}

//go:embed migrations/*.sql
var migrationsDir embed.FS

func RunMigrations(dsn string) error {
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

func (db *db) Close() {
	db.pool.Close()
}

func (db *db) Ping(ctx context.Context) error {
	err := db.pool.Ping(ctx)
	if err != nil {
		return fmt.Errorf("ping error: %w", err)
	}
	return nil
}
