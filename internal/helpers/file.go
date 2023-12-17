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

func FileData(data map[string]string, filepath string) error {
	lastUUID := make([]FileStorage, 0)
	bs, err := os.ReadFile(filepath)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("file reading error: %w", err)
	}
	if len(bs) > 0 {
		err = json.Unmarshal(bs, &lastUUID)
		if err != nil {
			return fmt.Errorf("json unmarshalling error: %w", err)
		}
	}

	for _, record := range lastUUID {
		shortURL := record.ShortURL
		originalURL := record.OriginalURL
		data[shortURL] = originalURL
	}
	return nil
}
