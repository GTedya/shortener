package main

import (
	"github.com/GTedya/shortener/database"

	"github.com/GTedya/shortener/config"
	"github.com/GTedya/shortener/internal/app/logger"
	"github.com/GTedya/shortener/internal/app/server"
)

func main() {
	conf := config.GetConfig()
	log := logger.CreateLogger()

	db, err := database.NewDB(conf.DatabaseDSN, log)
	if err != nil {
		log.Errorw("database creation error", err)
	}
	defer db.Close()

	err = server.Start(conf, log, db)
	if err != nil {
		log.Errorw("server starting error", err)
	}
}
