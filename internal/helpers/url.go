package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
)

type URLData struct {
	URLMap map[ShortURL]URL
}

type URL struct {
	URL string `json:"url"`
}

type ShortURL struct {
	URL string `json:"result"`
}

func CreateURLData(filepath string) (map[string]string, error) {
	lastUUID := make([]FileStorage, 0)

	bs, err := os.ReadFile(filepath)
	if os.IsNotExist(err) {
		_, er := os.Create(filepath)
		if er != nil {
			return nil, fmt.Errorf("file creation error: %w", er)
		}
	}
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("file reading error: %w", err)
	}
	if len(bs) > 0 {
		err = json.Unmarshal(bs, &lastUUID)
		if err != nil {
			return nil, fmt.Errorf("json unmarshalling error: %w", err)
		}
	}

	data := make(map[string]string)

	for _, record := range lastUUID {
		shortURL := record.ShortURL
		originalURL := record.OriginalURL
		data[shortURL] = originalURL
	}
	return data, nil
}

func (u URLData) GetByShortenURL(url string) (URL, error) {
	link, ok := u.URLMap[ShortURL{url}]
	if !ok {
		return URL{}, fmt.Errorf("неверный адресс URL")
	}
	return link, nil
}

func CreateUniqueID(check func(shortID string) (string, error), urlLen int) string {
	id := GenerateURL(urlLen)
	uniqueID := false
	for !uniqueID {
		_, err := check(id)
		if err != nil {
			id = GenerateURL(urlLen)
			uniqueID = true
		}
	}
	return id
}

func GenerateURL(n int) string {
	var charset = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
