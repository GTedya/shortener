package handlers

import (
	"context"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/GTedya/shortener/config"
	"github.com/GTedya/shortener/internal/app/middlewares"
	"github.com/GTedya/shortener/internal/app/models"
	"github.com/GTedya/shortener/internal/app/repository"
)

// handler представляет обработчик HTTP-запросов.
type handler struct {
	log  *zap.SugaredLogger
	repo Repository
	conf config.Config
}

// contentType представляет тип контента HTTP.
const contentType = "Content-Type"

// appJSON представляет значение Content-Type для JSON.
const appJSON = "application/json"

// Repository saves and retrieves data from storage.
type Repository interface {
	Save(ctx context.Context, shortURL models.ShortURL) error
	GetByID(ctx context.Context, id string) (models.ShortURL, error)
	ShortenByURL(ctx context.Context, url string) (models.ShortURL, error)
	GetUsersUrls(ctx context.Context, userID string) ([]models.ShortURL, error)
	Close(_ context.Context) error
	Check(ctx context.Context) error
	SaveBatch(ctx context.Context, batch []models.ShortURL) error
	DeleteUrls(ctx context.Context, urls []models.ShortURL) error
	GetUsersAndUrlsCount(ctx context.Context) (int, int, error)
}

// NewHandler создает новый экземпляр обработчика HTTP-запросов.
func NewHandler(logger *zap.SugaredLogger, conf config.Config) (Handler, error) {
	repo := repository.GetRepo(conf)

	return &handler{log: logger, conf: conf, repo: repo}, nil
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
	router.With(middleware.AuthCheck).Get("/api/user/urls", h.userURLS)

	// Удаляет сокращенные URL пользователя.
	router.With(middleware.AuthCheck).Delete("/api/user/urls", h.deleteUrls)

	// Return statistic
	router.With(middleware.IPCheck).Get("/api/internal/stats", h.getStats)
}
