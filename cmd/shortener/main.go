package main

import (
	_ "net/http/pprof"

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

	err = database.RunMigrations(conf.DatabaseDSN)
	if err != nil {
		log.Errorw("database migration error", err)
	}

	err = server.Start(conf, log, db)
	if err != nil {
		log.Errorw("server starting error", err)
	}
}
