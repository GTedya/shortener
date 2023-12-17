package handlers

import (
	"math/rand"
)

type URL struct {
	URL string `json:"url"`
}

type ShortURL struct {
	URL string `json:"result"`
}

func generateURL(n int) string {
	var charset = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func createUniqueID(check func(shortID string) (string, error), urlLen int) string {
	id := generateURL(urlLen)
	uniqueID := false
	for !uniqueID {
		_, err := check(id)
		if err != nil {
			id = generateURL(urlLen)
			uniqueID = true
		}
	}
	return id
}
