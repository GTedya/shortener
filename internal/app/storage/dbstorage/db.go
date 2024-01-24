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

var ErrDuplicate = errors.New("this url already exists")

type DatabaseStore struct {
	DB *database.DB
}

func (ds *DatabaseStore) GetURL(ctx context.Context, shortID string) (string, error) {
	url, err := ds.DB.GetBasicURL(ctx, shortID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return "", fmt.Errorf("URL not found in database: %w", err)
	case err != nil:
		return "", fmt.Errorf("query error: %w", err)
	}
	return url, nil
}

func (ds *DatabaseStore) SaveURL(ctx context.Context, id, shortID string) error {
	var pgError *pgconn.PgError

	rows, err := ds.DB.SaveURL(ctx, id, shortID)
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

func (ds *DatabaseStore) Batch(ctx context.Context, urls map[string]string) error {
	err := ds.DB.Batch(ctx, urls)
	if err != nil {
		return fmt.Errorf("database batch error: %w", err)
	}
	return nil
}
