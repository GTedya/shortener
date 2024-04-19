package storage

import (
	"context"
	"fmt"

	"github.com/GTedya/shortener/config"
	"github.com/GTedya/shortener/database"
	"github.com/GTedya/shortener/internal/app/storage/dbstorage"
	"github.com/GTedya/shortener/internal/app/storage/filestorage"
	"github.com/GTedya/shortener/internal/app/storage/inmemory"
	"github.com/GTedya/shortener/internal/helpers"
)

type ReqMultipleURL struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
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

func NewStore(conf config.Config, db database.DB) (Store, error) {
	var store Store
	if len(conf.DatabaseDSN) != 0 {
		store = &dbstorage.DatabaseStore{DB: db}
		return store, nil
	}

	data := make(map[string]string)
	if len(conf.FileStoragePath) != 0 {
		err := helpers.FileData(data, conf.FileStoragePath)
		if err != nil {
			return nil, fmt.Errorf("unable to get file data: %w", err)
		}
		store = &filestorage.FileStore{Memory: inmemory.MemoryStore{Conf: conf, Data: data}}
		return store, nil
	}

	store = &inmemory.MemoryStore{Conf: conf, Data: data}
	return store, nil
}
