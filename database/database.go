// Package database предоставляет функциональность для работы с базой данных.
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

// db представляет реализацию интерфейса DB для работы с базой данных.
type db struct {
	pool *pgxpool.Pool      // Пул подключений к базе данных.
	log  *zap.SugaredLogger // Логгер для ведения журнала событий.
}

// DB представляет интерфейс для работы с базой данных.
type DB interface {
	Ping(ctx context.Context) error // Проверяет подключение к базе данных.
	Close()                         // Закрывает подключение к базе данных.
	GetURLS
	SaveURLS
	DeleteURLS
}

// BaseURL представляет базовый URL для сокращенных ссылок.
var BaseURL = "http://localhost:8080/"

// ErrQuery представляет ошибку запроса к базе данных.
const ErrQuery = "query error: %w"

// ErrCommitTransaction представляет ошибку фиксации транзакции.
const ErrCommitTransaction = "transaction commit error"

// DeleteBuffer определяет размер буфера для удаления записей из базы данных.
const DeleteBuffer = 10

// NewDB создает новый экземпляр базы данных на основе переданных параметров.
func NewDB(dsn string, logger *zap.SugaredLogger) (DB, error) {
	pool, err := NewPool(dsn)
	if err != nil {
		return nil, fmt.Errorf("NewPool error: %w", err)
	}
	return &db{pool: pool, log: logger}, nil
}

// NewPool создает и возвращает новый пул соединений с базой данных PostgreSQL.
// Он использует переданную строку подключения (DSN) для создания пула соединений.
func NewPool(dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create a connection pool: %w", err)
	}
	return pool, nil
}

//go:embed migrations/*.sql
var migrationsDir embed.FS

// RunMigrations выполняет миграции базы данных.
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

// Close закрывает соединение с базой данных.
func (db *db) Close() {
	db.pool.Close()
}

// Ping проверяет соединение с базой данных.
func (db *db) Ping(ctx context.Context) error {
	err := db.pool.Ping(ctx)
	if err != nil {
		return fmt.Errorf("ping error: %w", err)
	}
	return nil
}
