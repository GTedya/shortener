package helpers

import (
	"fmt"
	"math/rand"
)

type URLData struct {
	URLMap map[string]string
}

func (u URLData) GetByShortenURL(url string) (string, error) {
	link, ok := u.URLMap[url]
	if !ok {
		return "", fmt.Errorf("неверный адресс URL")
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
