package datastore

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/GTedya/shortener/config"
	"github.com/GTedya/shortener/database"
	"github.com/GTedya/shortener/internal/helpers"
)

const writingPermission = 0600

var ErrDuplicate = errors.New("this url already exists")

type memoryStore struct {
	data map[string]string
	conf config.Config
}

type fileStore struct {
	memoryStore
}

type databaseStore struct {
	db *database.DB
}

type Store interface {
	GetURL(ctx context.Context, shortID string) (string, error)
	SaveURL(ctx context.Context, id, shortID string) error
}

func NewStore(conf config.Config, db *database.DB) (Store, error) {
	var store Store
	data := make(map[string]string)
	store = memoryStore{conf: conf, data: data}

	if len(conf.DatabaseDSN) != 0 {
		store = databaseStore{db: db}
		return store, nil
	}

	if len(conf.FileStoragePath) != 0 {
		err := helpers.FileData(data, conf.FileStoragePath)
		if err != nil {
			return nil, fmt.Errorf("unable to get file data: %w", err)
		}
		store = fileStore{memoryStore: memoryStore{conf: conf, data: data}}
	}
	return store, nil
}

func (ms memoryStore) GetURL(_ context.Context, shortID string) (string, error) {
	url, ok := ms.data[shortID]
	if !ok {
		return "", fmt.Errorf("URL not found in data list")
	}
	return url, nil
}

func (fs fileStore) GetURL(ctx context.Context, shortID string) (string, error) {
	url, err := fs.memoryStore.GetURL(ctx, shortID)
	if err != nil {
		return "", err
	}
	return url, nil
}

func (ds databaseStore) GetURL(_ context.Context, shortID string) (string, error) {
	url, err := ds.db.GetBasicURL(shortID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return "", fmt.Errorf("URL not found in database: %w", err)
	case err != nil:
		return "", fmt.Errorf("query error: %w", err)
	}
	return url, nil
}

func (ms memoryStore) SaveURL(_ context.Context, id, shortID string) error {
	ms.data[shortID] = id
	return nil
}

func (fs fileStore) SaveURL(ctx context.Context, id, shortID string) error {
	err := fs.memoryStore.SaveURL(ctx, id, shortID)
	if err != nil {
		return err
	}
	filePath := fs.conf.FileStoragePath
	jsonFile := helpers.FileStorage{
		UUID:        helpers.GenerateUUID(filePath),
		ShortURL:    shortID,
		OriginalURL: id,
	}

	content, err := os.ReadFile(filePath)

	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("file reading error: %w", err)
	}

	var storage []helpers.FileStorage
	if len(content) > 0 {
		if err = json.Unmarshal(content, &storage); err != nil {
			return fmt.Errorf("json unmarshal error: %w", err)
		}
	}

	storage = append(storage, jsonFile)

	encoded, err := json.MarshalIndent(storage, "", "  ")
	if err != nil {
		return fmt.Errorf("json marshal error: %w", err)
	}

	err = os.WriteFile(filePath, encoded, writingPermission)
	if err != nil {
		return fmt.Errorf("file writing error: %w", err)
	}
	return nil
}

func (ds databaseStore) SaveURL(ctx context.Context, id, shortID string) error {
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
