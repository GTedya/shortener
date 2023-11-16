package server

import (
	"github.com/GTedya/shortener/config"
	"github.com/GTedya/shortener/internal/app/handlers"
	"github.com/GTedya/shortener/internal/app/logger"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func Start(conf config.Config) {
	router := chi.NewRouter()

	handler := handlers.NewHandler()
	handler.Register(router, conf)

	middleware := logger.CreateLogHandler(router)

	sugaredLogger := logger.CreateLogger()
	defer sugaredLogger.Sync()

	err := http.ListenAndServe(conf.Address, middleware)
	if err != nil {
		sugaredLogger.Fatal(err)
	}
}
