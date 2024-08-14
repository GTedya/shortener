package main

import (
	"fmt"
	_ "net/http/pprof"

	"github.com/go-chi/chi/v5"

	"github.com/GTedya/shortener/config"
	"github.com/GTedya/shortener/internal/app/handlers"
	"github.com/GTedya/shortener/internal/app/logger"
	"github.com/GTedya/shortener/internal/app/middlewares"
	pb "github.com/GTedya/shortener/internal/app/proto"
	"github.com/GTedya/shortener/internal/app/repository"
	"github.com/GTedya/shortener/internal/app/server"
	"github.com/GTedya/shortener/internal/app/service"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

// main - основная функция, которая инициализирует конфигурацию, логгер, базу данных,
// запускает миграции и запускает сервер.
func main() {
	// Print environment flags
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	// Получение конфигурации из файла конфигурации.
	conf := config.GetConfig()
	// Создание логгера.
	log := logger.CreateLogger()
	repo := repository.GetRepo(conf)

	handler, err := handlers.NewHandler(log, conf)
	if err != nil {
		log.Errorw("handler creation error", err)
	}
	shortener := service.NewShortener(repo, &conf)

	grpcServer, err := pb.NewGRPCServer(shortener, conf)
	if err != nil {
		return
	}
	err = grpcServer.Run()
	if err != nil {
		return
	}

	middle := middlewares.Middleware{Log: log, SecretKey: conf.SecretKey, TrustedSubnet: conf.TrustedSubnet}

	router := chi.NewRouter()

	// Запуск сервера.
	err = server.Start(conf, log, router, handler, middle)
	if err != nil {
		log.Errorw("server starting error", err)
	}
}
