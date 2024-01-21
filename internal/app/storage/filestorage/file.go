package filestorage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/GTedya/shortener/internal/app/storage/inmemory"
	"github.com/GTedya/shortener/internal/helpers"
)

const writingPermission = 0600

type FileStore struct {
	Memory inmemory.MemoryStore
}

func (fs *FileStore) GetURL(ctx context.Context, shortID string) (string, error) {
	url, err := fs.Memory.GetURL(ctx, shortID)
	if err != nil {
		return "", err
	}
	return url, nil
}

func (fs *FileStore) SaveURL(ctx context.Context, id, shortID string) error {
	err := fs.Memory.SaveURL(ctx, id, shortID)
	if err != nil {
		return fmt.Errorf("memory store error: %w", err)
	}
	filePath := fs.Memory.Conf.FileStoragePath
	jsonFile := helpers.FileStorage{
		UUID:        helpers.GenerateUUID(filePath),
		ShortURL:    shortID,
		OriginalURL: id,
	}

	content, err := os.ReadFile(filePath)

	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("file reading error: %w", err)
	}

	var fileStorage []helpers.FileStorage
	if len(content) > 0 {
		if err = json.Unmarshal(content, &fileStorage); err != nil {
			return fmt.Errorf("json unmarshal error: %w", err)
		}
	}

	fileStorage = append(fileStorage, jsonFile)

	encoded, err := json.MarshalIndent(fileStorage, "", "  ")
	if err != nil {
		return fmt.Errorf("json marshal error: %w", err)
	}

	err = os.WriteFile(filePath, encoded, writingPermission)
	if err != nil {
		return fmt.Errorf("file writing error: %w", err)
	}
	return nil
}

func (fs *FileStore) Batch(ctx context.Context, urls map[string]string) error {
	err := fs.Memory.Batch(ctx, urls)
	if err != nil {
		return fmt.Errorf("Memory store error: %w", err)
	}

	filePath := fs.Memory.Conf.FileStoragePath
	content, err := os.ReadFile(filePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("file reading error: %w", err)
	}

	fileStorage := make([]helpers.FileStorage, 0)

	if len(content) > 0 {
		if err = json.Unmarshal(content, &fileStorage); err != nil {
			return fmt.Errorf("json unmarshal error: %w", err)
		}
	}

	for id, shortID := range urls {
		jsonFile := helpers.FileStorage{
			UUID:        helpers.GenerateUUID(filePath),
			ShortURL:    shortID,
			OriginalURL: id,
		}
		fileStorage = append(fileStorage, jsonFile)
	}

	encoded, err := json.MarshalIndent(fileStorage, "", "  ")
	if err != nil {
		return fmt.Errorf("json marshal error: %w", err)
	}

	err = os.WriteFile(filePath, encoded, writingPermission)
	if err != nil {
		return fmt.Errorf("file writing error: %w", err)
	}

	return nil
}
