package server

import (
	"database/sql"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/GTedya/shortener/config"
	"github.com/GTedya/shortener/internal/app/handlers"
	"github.com/GTedya/shortener/internal/app/middlewares"
	"github.com/go-chi/chi/v5"
)

func Start(conf config.Config, log *zap.SugaredLogger, db *sql.DB) error {
	router := chi.NewRouter()
	middleware := middlewares.Middleware{Log: log}
	router.Use(middleware.LogHandle, middleware.GzipCompressHandle, middleware.GzipDecompressMiddleware)

	handler := handlers.NewHandler(log)
	handler.Register(router, conf, db)

	err := http.ListenAndServe(conf.Address, router)
	if err != nil {
		log.Errorw("server serving error", err)
		return fmt.Errorf("server serving error: %w", err)
	}
	return nil
}
