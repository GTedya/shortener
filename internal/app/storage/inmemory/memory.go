package inmemory

import (
	"context"
	"fmt"

	"github.com/GTedya/shortener/config"
)

// MemoryStore представляет хранилище данных в памяти.
type MemoryStore struct {
	Data map[string]string // Data содержит отображение коротких URL на их полные версии.
	Conf config.Config     // Conf содержит конфигурацию приложения.
}

// Batch пакетно добавляет короткие URL в хранилище в памяти.
func (ms *MemoryStore) Batch(_ context.Context, urls map[string]string) error {
	for id, shortID := range urls {
		ms.Data[shortID] = id
	}
	return nil
}

// GetURL возвращает полный URL по его короткой версии из хранилища в памяти.
func (ms *MemoryStore) GetURL(_ context.Context, shortID string) (string, error) {
	url, ok := ms.Data[shortID]
	if !ok {
		return "", fmt.Errorf("URL not found in data list")
	}
	return url, nil
}

// SaveURL сохраняет полный URL и его короткую версию в хранилище в памяти.
func (ms *MemoryStore) SaveURL(_ context.Context, _, id, shortID string) error {
	ms.Data[shortID] = id
	return nil
}
