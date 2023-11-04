package config

import (
	"flag"
	"os"
)

var FlagRunAddr string
var BasicURL string

func ParseFlags() {
	flag.StringVar(&FlagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&BasicURL, "b", "", "basic shorten URL")
	flag.Parse()

	if os.Getenv("SERVER_ADDRESS") != "" {
		FlagRunAddr = os.Getenv("SERVER_ADDRESS")
	}

	if os.Getenv("BASE_URL") != "" {
		BasicURL = os.Getenv("BASE_URL")
	}
}
