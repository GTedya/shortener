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

// handler представляет обработчик HTTP-запросов.
type handler struct {
	log   *zap.SugaredLogger
	db    database.DB
	store Store
	conf  config.Config
}

// contentType представляет тип контента HTTP.
const contentType = "Content-Type"

// appJSON представляет значение Content-Type для JSON.
const appJSON = "application/json"

// Store предоставляет интерфейс для хранилища URL.
type Store interface {
	// GetURL получает оригинальный URL по его сокращенной версии.
	GetURL(ctx context.Context, shortID string) (string, error)

	// SaveURL сохраняет URL в хранилище и связывает его с сокращенной версией.
	SaveURL(ctx context.Context, token, id, shortID string) error

	// Batch пакетно сохраняет URL в хранилище и связывает их с сокращенными версиями.
	Batch(ctx context.Context, urls map[string]string) error
}

// NewHandler создает новый экземпляр обработчика HTTP-запросов.
func NewHandler(logger *zap.SugaredLogger, conf config.Config, db database.DB) (Handler, error) {
	store, err := storage.NewStore(conf, db)
	if err != nil {
		return nil, fmt.Errorf("store creation error: %w", err)
	}
	return &handler{log: logger, conf: conf, store: store, db: db}, nil
}

// Register регистрирует обработчики маршрутов HTTP в маршрутизаторе chi.
func (h *handler) Register(router *chi.Mux, middleware middlewares.Middleware) {
	// Создает сокращенный URL.
	router.Post("/", h.createURL)

	// Получает оригинальный URL по его сокращенной версии.
	router.Get("/{id}", h.getURLByID)

	// Создает сокращенный URL из JSON-данных.
	router.Post("/api/shorten", h.urlByJSON)

	// Проверяет доступность сервера.
	router.Get("/ping", h.getPing)

	// Пакетно создает сокращенные URL.
	router.Post("/api/shorten/batch", h.batch)

	// Получает все сокращенные URL пользователя.
	router.With(middleware.AuthCheck).Get("/api/user/urls", h.userUrls)

	// Удаляет сокращенные URL пользователя.
	router.With(middleware.AuthCheck).Delete("/api/user/urls", h.deleteUrls)
}
