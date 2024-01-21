package inmemory

import (
	"context"
	"fmt"

	"github.com/GTedya/shortener/config"
)

type MemoryStore struct {
	Data map[string]string
	Conf config.Config
}

func (ms *MemoryStore) Batch(_ context.Context, urls map[string]string) error {
	for id, shortID := range urls {
		ms.Data[shortID] = id
	}
	return nil
}

func (ms *MemoryStore) GetURL(_ context.Context, shortID string) (string, error) {
	url, ok := ms.Data[shortID]
	if !ok {
		return "", fmt.Errorf("URL not found in data list")
	}
	return url, nil
}

func (ms *MemoryStore) SaveURL(_ context.Context, id, shortID string) error {
	ms.Data[shortID] = id
	return nil
}
