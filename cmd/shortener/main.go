package main

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/GTedya/shortener/config"
	"github.com/GTedya/shortener/database"
	"github.com/GTedya/shortener/internal/app/logger"
	"github.com/GTedya/shortener/internal/app/server"
)

func main() {
	conf := config.GetConfig()
	log := logger.CreateLogger()

	db, err := database.CreateDB(conf, log)
	if err != nil {
		log.Errorw("database creation error", err)
	}
	defer func(db *sql.DB) {
		err = db.Close()
		if err != nil {
			log.Errorw("Database close connection error", err)
		}
	}(db)

	err = server.Start(conf, log, db)
	if err != nil {
		log.Errorw("server starting error", err)
	}
}
