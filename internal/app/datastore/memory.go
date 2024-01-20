package datastore

import (
	"context"
	"fmt"

	"github.com/GTedya/shortener/config"
)

type memoryStore struct {
	data map[string]string
	conf config.Config
}

func (ms *memoryStore) Batch(_ context.Context, urls map[string]string) error {
	for id, shortID := range urls {
		ms.data[shortID] = id
	}
	return nil
}

func (ms *memoryStore) GetURL(_ context.Context, shortID string) (string, error) {
	url, ok := ms.data[shortID]
	if !ok {
		return "", fmt.Errorf("URL not found in data list")
	}
	return url, nil
}

func (ms *memoryStore) SaveURL(_ context.Context, id, shortID string) error {
	ms.data[shortID] = id
	return nil
}
