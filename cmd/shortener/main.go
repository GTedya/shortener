package main

import (
	"github.com/GTedya/shortener/config"
	"github.com/GTedya/shortener/internal/app/server"
	"log"
)

func main() {
	conf := config.GetConfig()

	err := server.Start(conf)
	if err != nil {
		log.Fatal(err)
	}
}
