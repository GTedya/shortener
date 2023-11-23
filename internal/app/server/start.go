package server

import (
	"github.com/GTedya/shortener/config"
	"github.com/GTedya/shortener/internal/app/handlers"
	"github.com/GTedya/shortener/internal/app/logger"
	"github.com/GTedya/shortener/internal/app/middlewares"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func Start(conf config.Config) {
	router := chi.NewRouter()

	handler := handlers.NewHandler()
	handler.Register(router, conf)

	middleware := logger.CreateLogHandler(router)
	squeeze := middlewares.CompressHandle(middleware)
	sugaredLogger := logger.CreateLogger()
	defer sugaredLogger.Sync()

	err := http.ListenAndServe(conf.Address, squeeze)
	if err != nil {
		sugaredLogger.Fatal(err)
	}
}
