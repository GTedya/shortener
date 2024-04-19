package config

import (
	"flag"
	"os"
)

type Config struct {
	Address         string
	URL             string
	FileStoragePath string
	DatabaseDSN     string
}

func GetConfig() (c Config) {
	flag.StringVar(&c.Address, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&c.URL, "b", "http://localhost:8080", "basic shorten URL")
	flag.StringVar(&c.DatabaseDSN, "d", "postgres://root:root@localhost:5432/shortener?sslmode=disable", "database dsn")
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
