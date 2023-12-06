package main

import (
	"github.com/GTedya/shortener/config"
	"github.com/GTedya/shortener/internal/app/logger"
	"github.com/GTedya/shortener/internal/app/server"
)

func main() {
	conf := config.GetConfig()
	log := logger.CreateLogger()

	err := server.Start(conf, log)
	if err != nil {
		log.Errorw("server starting error", err)
	}
}
