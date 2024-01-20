package datastore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/GTedya/shortener/database"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type databaseStore struct {
	db *database.DB
}

func (ds *databaseStore) GetURL(_ context.Context, shortID string) (string, error) {
	url, err := ds.db.GetBasicURL(shortID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return "", fmt.Errorf("URL not found in database: %w", err)
	case err != nil:
		return "", fmt.Errorf("query error: %w", err)
	}
	return url, nil
}

func (ds *databaseStore) SaveURL(ctx context.Context, id, shortID string) error {
	var pgError *pgconn.PgError

	rows, err := ds.db.SaveURL(ctx, id, shortID)
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

func (ds *databaseStore) Batch(ctx context.Context, urls map[string]string) error {
	err := ds.db.Batch(ctx, urls)
	if err != nil {
		return fmt.Errorf("database batch error: %w", err)
	}
	return nil
}
