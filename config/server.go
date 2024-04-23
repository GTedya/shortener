// Package config предоставляет функциональность для работы с конфигурацией приложения.
package config

import (
	"flag"
	"os"
)

// Config представляет структуру конфигурации приложения.
type Config struct {
	Address         string // Адрес и порт, на котором запускается сервер.
	URL             string // Базовый URL для сокращенных ссылок.
	FileStoragePath string // Путь к файловому хранилищу.
	DatabaseDSN     string // DSN для подключения к базе данных.
}

// GetConfig получает конфигурацию из флагов командной строки и окружения.
func GetConfig() (c Config) {
	flag.StringVar(&c.Address, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&c.URL, "b", "http://localhost:8080", "basic shorten URL")
	flag.StringVar(&c.DatabaseDSN, "d", "", "database dsn")
	flag.StringVar(&c.FileStoragePath, "f", "/tmp/short-url-database.json", "file storage path")
	flag.Parse()

	serverAddress, ok := os.LookupEnv("SERVER_ADDRESS")
	if ok {
		c.Address = serverAddress
	}

	url, ok := os.LookupEnv("BASE_URL")
	if ok {
		c.URL = url
	}

	database, ok := os.LookupEnv("DATABASE_DSN")
	if ok {
		c.DatabaseDSN = database
	}

	filePath, ok := os.LookupEnv("FILE_STORAGE_PATH")
	if ok {
		c.FileStoragePath = filePath
	}
	return c
}
