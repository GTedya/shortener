package helpers

import (
	"fmt"
	"math/rand"
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
