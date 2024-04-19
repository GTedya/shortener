package handlers

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/GTedya/shortener/config"
	"github.com/GTedya/shortener/database"
	"github.com/GTedya/shortener/internal/app/middlewares"
	"github.com/GTedya/shortener/internal/app/storage"

	"github.com/go-chi/chi/v5"
)

type handler struct {
	log   *zap.SugaredLogger
	db    database.DB
	store Store
	conf  config.Config
}

const contentType = "Content-Type"
const appJSON = "application/json"

type Store interface {
	GetURL(ctx context.Context, shortID string) (string, error)
	SaveURL(ctx context.Context, id, shortID string) error
	Batch(ctx context.Context, urls map[string]string) error
}

func NewHandler(logger *zap.SugaredLogger, conf config.Config, db database.DB) (Handler, error) {
	store, err := storage.NewStore(conf, db)
	if err != nil {
		return nil, fmt.Errorf("store creation error: %w", err)
	}
	return &handler{log: logger, conf: conf, store: store, db: db}, nil
}

func (h *handler) Register(router *chi.Mux, middleware middlewares.Middleware) {
	router.With(middleware.TokenCreate).Post("/", h.createURL)

	router.Get("/{id}", h.getURLByID)

	router.With(middleware.TokenCreate).Post("/api/shorten", h.urlByJSON)

	router.Get("/ping", h.getPing)

	router.Post("/api/shorten/batch", h.batch)

	router.With(middleware.AuthCheck).Get("/api/user/urls", h.userUrls)

	router.With(middleware.AuthCheck).Delete("/api/user/urls", h.deleteUrls)
}
