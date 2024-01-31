package server

import (
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/GTedya/shortener/config"
	"github.com/GTedya/shortener/database"
	"github.com/GTedya/shortener/internal/app/handlers"
	"github.com/GTedya/shortener/internal/app/middlewares"
	"github.com/go-chi/chi/v5"
)

func Start(conf config.Config, log *zap.SugaredLogger, db *database.DB) error {
	router := chi.NewRouter()
	middleware := middlewares.Middleware{Log: log}
	router.Use(middleware.LogHandle, middleware.GzipCompressHandle, middleware.GzipDecompressMiddleware)

	handler, err := handlers.NewHandler(log, conf, db)
	if err != nil {
		return fmt.Errorf("handler creation error: %w", err)
	}

	handler.Register(router, middleware)

	err = http.ListenAndServe(conf.Address, router)
	if err != nil {
		return fmt.Errorf("server serving error: %w", err)
	}
	return nil
}
