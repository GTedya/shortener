package dbstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/GTedya/shortener/database"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// ErrDuplicate возвращается при попытке сохранить URL, который уже существует в базе данных.
var ErrDuplicate = errors.New("this url already exists")

// ErrDeletedURL возвращается при попытке получить URL, который был удален из базы данных.
var ErrDeletedURL = errors.New("this url deleted")

// DatabaseStore представляет хранилище данных в базе данных.
type DatabaseStore struct {
	DB database.DB // DB содержит интерфейс для выполнения операций с базой данных.
}

// GetURL возвращает полный URL по его короткой версии из базы данных.
func (ds *DatabaseStore) GetURL(ctx context.Context, shortID string) (string, error) {
	url, isDeleted, err := ds.DB.GetBasicURL(ctx, shortID)
	switch {
	case isDeleted:
		return "", ErrDeletedURL
	case errors.Is(err, sql.ErrNoRows):
		return "", fmt.Errorf("URL not found in database: %w", err)
	case err != nil:
		return "", fmt.Errorf("query error: %w", err)
	}
	return url, nil
}

// SaveURL сохраняет полный URL и его короткую версию в базе данных.
func (ds *DatabaseStore) SaveURL(ctx context.Context, token, id, shortID string) error {
	var pgError *pgconn.PgError

	rows, err := ds.DB.SaveURL(ctx, token, id, shortID)
	if errors.As(err, &pgError) && pgError.Code == pgerrcode.UniqueViolation {
		return ErrDuplicate
	}
	if err != nil {
		return fmt.Errorf("saving url query error: %w", err)
	}
	if rows != 1 {
		return fmt.Errorf("expected to affect 1 row, affected %d", rows)
	}
	return nil
}

// Batch добавляет короткие URL в базу данных.
func (ds *DatabaseStore) Batch(ctx context.Context, urls map[string]string) error {
	err := ds.DB.Batch(ctx, urls)
	if err != nil {
		return fmt.Errorf("database batch error: %w", err)
	}
	return nil
}
