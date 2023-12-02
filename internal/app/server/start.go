package server

import (
	"log"
	"net/http"

	"go.uber.org/zap"

	"github.com/GTedya/shortener/config"
	"github.com/GTedya/shortener/internal/app/handlers"
	"github.com/GTedya/shortener/internal/app/logger"
	"github.com/GTedya/shortener/internal/app/middlewares"
	"github.com/go-chi/chi/v5"
)

func Start(conf config.Config) {
	router := chi.NewRouter()
	router.Use(middlewares.LogHandle, middlewares.GzipCompressHandle, middlewares.GzipDecompressMiddleware)

	handler := handlers.NewHandler()
	handler.Register(router, conf)

	sugaredLogger := logger.CreateLogger()
	defer func(sugaredLogger *zap.SugaredLogger) {
		err := sugaredLogger.Sync()
		if err != nil {
			log.Fatal(err)
		}
	}(sugaredLogger)

	err := http.ListenAndServe(conf.Address, router)
	if err != nil {
		sugaredLogger.Info(err)
	}
}
