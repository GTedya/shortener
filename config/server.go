// Package config предоставляет функциональность для работы с конфигурацией приложения.
package config

import (
	"encoding/json"
	"flag"
	"os"
	"strconv"
)

// Config представляет структуру конфигурации приложения.
type Config struct {
	Address         string `json:"server_address"`    // Адрес и порт, на котором запускается сервер.
	URL             string `json:"base_url"`          // Базовый URL для сокращенных ссылок.
	FileStoragePath string `json:"file_storage_path"` // Путь к файловому хранилищу.
	DatabaseDSN     string `json:"database_dsn"`      // DSN для подключения к базе данных.
	SecretKey       string // Секретный клюя для токена
	EnableHTTPS     bool   `json:"enable_https"` // enable HTTPS on server

}

// GetConfig получает конфигурацию из флагов командной строки и окружения.
func GetConfig() (c Config) {
	c.getConfigFromJSON()

	flag.StringVar(&c.Address, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&c.URL, "b", "http://localhost:8080", "basic shorten URL")
	flag.StringVar(&c.DatabaseDSN, "d", "postgres://root:root@localhost:5432/shortener?sslmode=disable", "database dsn")
	flag.StringVar(&c.FileStoragePath, "f", "/tmp/short-url-database.json", "file storage path")
	flag.StringVar(&c.SecretKey, "sk", "secret_key", "secret key")
	flag.BoolVar(&c.EnableHTTPS, "s", false, "enable HTTPS on server")
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

	EnableHTTPS, ok := os.LookupEnv("ENABLE_HTTPS")
	if ok {
		boolValue, err := strconv.ParseBool(EnableHTTPS)
		if err != nil {
			return c
		}
		c.EnableHTTPS = boolValue
	}
	return c
}

// getConfigFromJSON loads the server configuration from a JSON file.
// It first checks if a command-line flag "-config" is provided to specify
// the configuration file path. If not, it checks the environment variable
// "CONFIG". If the file path is found, it reads the file and unmarshals
// its content into the Config struct.
func (c *Config) getConfigFromJSON() {
	var path string

	flag.StringVar(&path, "config", "", "server config file path")
	flag.Parse()

	filePath, ok := os.LookupEnv("CONFIG")
	if ok {
		path = filePath
	}

	file, err := os.ReadFile(path)
	if err != nil {
		return
	}

	err = json.Unmarshal(file, c)
	if err != nil {
		return
	}
}
