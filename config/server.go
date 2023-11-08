package config

import (
	"flag"
	"os"
)

type Config struct {
	Address string
	URL     string
}

func GetConfig() (c Config) {
	flag.StringVar(&c.Address, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&c.URL, "b", "short", "basic shorten URL")
	flag.Parse()

	serverAddress, ok := os.LookupEnv("SERVER_ADDRESS")
	if ok {
		c.Address = serverAddress
	}

	url, ok := os.LookupEnv("BASE_URL")
	if ok {
		c.URL = url
	}
	return c
}
