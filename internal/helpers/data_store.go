package helpers

import (
	"encoding/json"
	"net/url"
	"os"

	"github.com/GTedya/shortener/internal/app/logger"
)

const urlLen = 6
const writingPermission = 0600

func MemoryStore(data *URLData, body URL, basicURL string) ShortURL {
	id := basicURL + GenerateURL(urlLen)
	uniqueID := false
	for !uniqueID {
		_, ok := data.URLMap[ShortURL{URL: id}]
		if ok {
			id = basicURL + GenerateURL(urlLen)
			continue
		}
		uniqueID = true
	}
	encodedID := ShortURL{URL: url.PathEscape(id)}
	originalURL := body

	data.URLMap[encodedID] = originalURL
	return encodedID
}

func FileStore(data *URLData, body URL, basicURL string, filePath string) string {
	short := MemoryStore(data, body, basicURL)
	jsonFile := FileStorage{
		UUID:        GenerateUUID(filePath),
		ShortURL:    short.URL,
		OriginalURL: data.URLMap[short].URL,
	}

	log := logger.CreateLogger()
	content, err := os.ReadFile(filePath)

	if err != nil && !os.IsNotExist(err) {
		log.Error(err)
		return ""
	}

	var storage []FileStorage
	if len(content) > 0 {
		if err := json.Unmarshal(content, &storage); err != nil {
			log.Error(err)
			return ""
		}
	}

	storage = append(storage, jsonFile)

	encoded, err := json.MarshalIndent(storage, "", "  ")
	if err != nil {
		log.Error(err)
		return ""
	}

	err = os.WriteFile(filePath, encoded, writingPermission)
	if err != nil {
		log.Error(err)
		return ""
	}
	return short.URL
}
