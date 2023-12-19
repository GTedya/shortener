package main

import (
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/GTedya/shortener/config"
	"github.com/GTedya/shortener/internal/app/logger"
	"github.com/GTedya/shortener/internal/app/server"
)

func main() {
	conf := config.GetConfig()
	log := logger.CreateLogger()

	db, err := config.CreateDBConn(conf, log)
	if err != nil {
		log.Errorw("database creation error", err)
	}
	defer func() {
		err = db.Close()
		if err != nil {
			log.Errorw("Database close connection error", err)
		}
	}()

	err = server.Start(conf, log, db)
	if err != nil {
		log.Errorw("server starting error", err)
	}
}
