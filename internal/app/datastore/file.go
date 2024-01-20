package datastore

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/GTedya/shortener/internal/helpers"
)

type fileStore struct {
	memoryStore
}

func (fs *fileStore) GetURL(ctx context.Context, shortID string) (string, error) {
	url, err := fs.memoryStore.GetURL(ctx, shortID)
	if err != nil {
		return "", err
	}
	return url, nil
}

func (fs *fileStore) SaveURL(ctx context.Context, id, shortID string) error {
	err := fs.memoryStore.SaveURL(ctx, id, shortID)
	if err != nil {
		return fmt.Errorf("memory store error: %w", err)
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

func (fs *fileStore) Batch(ctx context.Context, urls map[string]string) error {
	err := fs.memoryStore.Batch(ctx, urls)
	if err != nil {
		return fmt.Errorf("memory store error: %w", err)
	}

	filePath := fs.conf.FileStoragePath
	content, err := os.ReadFile(filePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("file reading error: %w", err)
	}

	storage := make([]helpers.FileStorage, 0)

	if len(content) > 0 {
		if err = json.Unmarshal(content, &storage); err != nil {
			return fmt.Errorf("json unmarshal error: %w", err)
		}
	}

	for id, shortID := range urls {
		jsonFile := helpers.FileStorage{
			UUID:        helpers.GenerateUUID(filePath),
			ShortURL:    shortID,
			OriginalURL: id,
		}
		storage = append(storage, jsonFile)
	}

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
