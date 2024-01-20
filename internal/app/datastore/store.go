package datastore

import (
	"context"
	"errors"
	"fmt"

	"github.com/GTedya/shortener/config"
	"github.com/GTedya/shortener/database"
	"github.com/GTedya/shortener/internal/helpers"
)

const writingPermission = 0600

var ErrDuplicate = errors.New("this url already exists")

type ReqMultipleURL struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type MultipleURL struct {
	OriginalURL string
	ShortURL    string
}

type ResMultipleURL struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type Store interface {
	GetURL(ctx context.Context, shortID string) (string, error)
	SaveURL(ctx context.Context, id, shortID string) error
	Batch(ctx context.Context, urls map[string]string) error
}

func NewStore(conf config.Config, db *database.DB) (Store, error) {
	var store Store
	data := make(map[string]string)
	store = &memoryStore{conf: conf, data: data}

	if len(conf.DatabaseDSN) != 0 {
		store = &databaseStore{db: db}
		return store, nil
	}

	if len(conf.FileStoragePath) != 0 {
		err := helpers.FileData(data, conf.FileStoragePath)
		if err != nil {
			return nil, fmt.Errorf("unable to get file data: %w", err)
		}
		store = &fileStore{memoryStore: memoryStore{conf: conf, data: data}}
	}
	return store, nil
}
