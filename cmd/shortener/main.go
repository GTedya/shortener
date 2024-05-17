package main

import (
	_ "net/http/pprof"

	"github.com/GTedya/shortener/database"

	"github.com/GTedya/shortener/config"
	"github.com/GTedya/shortener/internal/app/logger"
	"github.com/GTedya/shortener/internal/app/server"
)

// main - основная функция, которая инициализирует конфигурацию, логгер, базу данных,
// запускает миграции и запускает сервер.
func main() {
	// Получение конфигурации из файла конфигурации.
	conf := config.GetConfig()
	// Создание логгера.
	log := logger.CreateLogger()

	// Инициализация базы данных.
	db, err := database.NewDB(conf.DatabaseDSN, log)
	if err != nil {
		log.Errorw("database creation error", err)
	}
	defer db.Close()

	// Запуск миграций базы данных.
	err = database.RunMigrations(conf.DatabaseDSN)
	if err != nil {
		log.Errorw("database migration error", err)
	}

	// Запуск сервера.
	err = server.Start(conf, log, db)
	if err != nil {
		log.Errorw("server starting error", err)
	}
}
