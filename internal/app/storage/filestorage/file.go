package filestorage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/GTedya/shortener/internal/app/storage/inmemory"
	"github.com/GTedya/shortener/internal/helpers"
)

// writingPermission устанавливает права доступа для записи в файл.
const writingPermission = 0600

// FileStore представляет хранилище данных в файле.
type FileStore struct {
	Memory inmemory.MemoryStore // Memory содержит ссылку на хранилище в памяти.
}

// GetURL возвращает полный URL по его короткой версии из хранилища в файле.
func (fs *FileStore) GetURL(ctx context.Context, shortID string) (string, error) {
	url, err := fs.Memory.GetURL(ctx, shortID)
	if err != nil {
		return "", fmt.Errorf("URL getting error: %w", err)
	}
	return url, nil
}

// SaveURL сохраняет полный URL и его короткую версию в хранилище в файле.
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

// Batch добавляет короткие URL в хранилище в файле.
func (fs *FileStore) Batch(ctx context.Context, urls map[string]string) error {
	err := fs.Memory.Batch(ctx, urls)
	if err != nil {
		return fmt.Errorf("memory store error: %w", err)
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
