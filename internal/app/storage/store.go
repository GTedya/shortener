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

// ReqMultipleURL представляет структуру запроса для создания нескольких коротких URL.
type ReqMultipleURL struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// ResMultipleURL представляет структуру ответа с коротким URL и соответствующим ID запроса.
type ResMultipleURL struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// Store определяет интерфейс хранилища для работы с короткими URL.
type Store interface {
	GetURL(ctx context.Context, shortID string) (string, error)
	SaveURL(ctx context.Context, id, shortID string) error
	Batch(ctx context.Context, urls map[string]string) error
}

// NewStore создает новый экземпляр хранилища в зависимости от конфигурации.
func NewStore(conf config.Config, db database.DB) (Store, error) {
	var store Store

	// Если задана конфигурация для базы данных, используется хранилище базы данных.
	if len(conf.DatabaseDSN) != 0 {
		store = &dbstorage.DatabaseStore{DB: db}
		return store, nil
	}

	data := make(map[string]string)

	// Если задан путь к файловому хранилищу, данные загружаются из файла и используется файловое хранилище.
	if len(conf.FileStoragePath) != 0 {
		err := helpers.FileData(data, conf.FileStoragePath)
		if err != nil {
			return nil, fmt.Errorf("unable to get file data: %w", err)
		}
		store = &filestorage.FileStore{Memory: inmemory.MemoryStore{Conf: conf, Data: data}}
		return store, nil
	}

	// Иначе используется встроенное в память хранилище.
	store = &inmemory.MemoryStore{Conf: conf, Data: data}
	return store, nil
}
