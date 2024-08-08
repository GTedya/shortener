// Package config предоставляет функциональность для работы с конфигурацией приложения.
package config

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
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
	TrustedSubnet   string `json:"trusted_subnet"` // TrustedSubnet
	MigrationPath   string // migration directory path
	EnableHTTPS     bool   `json:"enable_https"` // enable HTTPS on server
}

// GetConfig initializes the configuration from command-line flags, environment variables, or a JSON file.
func GetConfig() Config {
	var c Config
	var path string

	flag.StringVar(&path, "config", "", "server config file path")
	if path != "" {
		if conf, err := loadConfigFromJSON(path); err == nil {
			c = conf
		} else {
			log.Printf("Warning: failed to load config from JSON: %v", err)
		}
	}

	flag.StringVar(&c.Address, "a", ":8080", "address and port to run server")
	flag.StringVar(&c.URL, "b", "http://localhost:8080", "basic shorten URL")
	flag.StringVar(&c.DatabaseDSN, "d", "postgres://root:root@localhost:5432/shortener?sslmode=disable", "database dsn")
	flag.StringVar(&c.FileStoragePath, "f", "/tmp/short-url-database.json", "file storage path")
	flag.StringVar(&c.MigrationPath, "m", "file://internal/app/repository/migrations", "migration directory path")
	flag.StringVar(&c.SecretKey, "sk", "secret_key", "secret key")
	flag.BoolVar(&c.EnableHTTPS, "s", false, "enable HTTPS on server")
	flag.StringVar(&c.TrustedSubnet, "t", "172.17.0.0/16", "TrustedSubnet")
	flag.Parse()

	overrideConfigWithEnvVars(&c)

	return c
}

// loadConfigFromJSON loads the server configuration from a JSON file.
func loadConfigFromJSON(path string) (c Config, err error) {
	filePath, ok := os.LookupEnv("CONFIG")
	if ok {
		path = filePath
	}

	if path == "" {
		return Config{}, fmt.Errorf("config path is empty")
	}

	file, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	if err := json.Unmarshal(file, &c); err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return c, nil
}

// overrideConfigWithEnvVars overrides configuration values with environment variables if they are set.
func overrideConfigWithEnvVars(c *Config) {
	envVars := map[string]*string{
		"SERVER_ADDRESS":    &c.Address,
		"BASE_URL":          &c.URL,
		"DATABASE_DSN":      &c.DatabaseDSN,
		"FILE_STORAGE_PATH": &c.FileStoragePath,
		"SECRET_KEY":        &c.SecretKey,
		"TRUSTED_SUBNET":    &c.TrustedSubnet,
	}
	for env, ptr := range envVars {
		if value, ok := os.LookupEnv(env); ok {
			*ptr = value
		}
	}

	if enableHTTPS, ok := os.LookupEnv("ENABLE_HTTPS"); ok {
		if boolValue, err := strconv.ParseBool(enableHTTPS); err == nil {
			c.EnableHTTPS = boolValue
		}
	}
}
