package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/GTedya/shortener/internal/app/logger"
)

type FileStorage struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

const writingPermission = 0600

func GenerateUUID(filepath string) string {
	lastUUID := make([]FileStorage, 0)
	log := logger.CreateLogger()

	bs, err := os.ReadFile(filepath)
	if err != nil && !errors.Is(err, io.EOF) {
		log.Info(err)
	}
	if len(bs) > 0 {
		err = json.Unmarshal(bs, &lastUUID)
		if err != nil {
			log.Info(err)
		}
	}

	if len(lastUUID) == 0 {
		return "1"
	}

	id, err := strconv.Atoi(lastUUID[len(lastUUID)-1].UUID)
	if err != nil {
		log.Info(err)
	}
	return strconv.Itoa(id + 1)
}

func AppendToFile(fileStoragePath string, jsonFile FileStorage) error {
	log := logger.CreateLogger()
	content, err := os.ReadFile(fileStoragePath)
	if err != nil && !os.IsNotExist(err) {
		log.Error(err)
		return fmt.Errorf("error during file opening: %w", err)
	}

	var storage []FileStorage
	if len(content) > 0 {
		if err := json.Unmarshal(content, &storage); err != nil {
			log.Error(err)
			return fmt.Errorf("error during json.Unmarshaling: %w", err)
		}
	}

	storage = append(storage, jsonFile)

	encoded, err := json.MarshalIndent(storage, "", "  ")
	if err != nil {
		log.Error(err)
		return fmt.Errorf("error during json.MarshalIndent: %w", err)
	}

	err = os.WriteFile(fileStoragePath, encoded, writingPermission)
	if err != nil {
		log.Error(err)
		return fmt.Errorf("error during writing in file: %w", err)
	}

	return nil
}
