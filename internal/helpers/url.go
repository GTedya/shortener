package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"

	"github.com/GTedya/shortener/internal/app/logger"
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

func CreateURLMap(filepath string) URLData {
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

	data := URLData{
		URLMap: make(map[ShortURL]URL),
	}

	for _, record := range lastUUID {
		shortURL := ShortURL{URL: record.ShortURL}
		originalURL := URL{URL: record.OriginalURL}
		data.URLMap[shortURL] = originalURL
	}
	return data
}

func (u URLData) GetByShortenURL(url string) (URL, error) {
	link, ok := u.URLMap[ShortURL{url}]
	if !ok {
		return URL{}, fmt.Errorf("неверный адресс URL")
	}
	return link, nil
}

func GenerateURL(n int) string {
	var charset = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
