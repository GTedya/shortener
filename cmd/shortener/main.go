package main

import (
	"github.com/GTedya/shortener/config"
	"github.com/GTedya/shortener/internal/app/server"
)

func main() {
	conf := config.GetConfig()

	server.Start(conf)
}
