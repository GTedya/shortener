package helpers

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/GTedya/shortener/config"
)

const writingPermission = 0600

type memoryStore struct {
	data map[string]string
	conf config.Config
}
type fileStore struct {
	memoryStore
}

type Store interface {
	GetURL(shortID string) (string, error)
	SaveURL(id, shortID string) error
}

func NewStore(conf config.Config, data map[string]string) Store {
	var store Store
	store = memoryStore{conf: conf, data: data}
	if len(conf.FileStoragePath) != 0 {
		store = fileStore{memoryStore: memoryStore{conf: conf, data: data}}
	}
	return store
}

func (ms memoryStore) GetURL(shortID string) (string, error) {
	url, ok := ms.data[shortID]
	if !ok {
		return "", fmt.Errorf("URL not found in data list")
	}
	return url, nil
}

func (fs fileStore) GetURL(shortID string) (string, error) {
	url, err := fs.memoryStore.GetURL(shortID)
	if err != nil {
		return "", err
	}
	return url, nil
}

func (ms memoryStore) SaveURL(id, shortID string) error {
	ms.data[shortID] = id
	return nil
}

func (fs fileStore) SaveURL(id, shortID string) error {
	err := fs.memoryStore.SaveURL(id, shortID)
	if err != nil {
		return err
	}
	filePath := fs.conf.FileStoragePath
	jsonFile := FileStorage{
		UUID:        GenerateUUID(filePath),
		ShortURL:    shortID,
		OriginalURL: id,
	}

	content, err := os.ReadFile(filePath)

	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("file reading error: %w", err)
	}

	var storage []FileStorage
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
