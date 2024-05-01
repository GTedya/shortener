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
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

// Start запускает HTTP-сервер для обслуживания запросов.
// Принимает конфигурацию сервера, логгер и экземпляр базы данных.
// Возвращает ошибку, если сервер не удалось запустить.
func Start(conf config.Config, log *zap.SugaredLogger, db database.DB) error {
	// Создание нового маршрутизатора Chi.
	router := chi.NewRouter()

	// Создание экземпляра посредников.
	middleware := middlewares.Middleware{Log: log, SecretKey: conf.SecretKEY}

	// Использование посредников для обработки логов, сжатия gzip и декомпрессии gzip.
	router.Use(middleware.LogHandle, middleware.GzipCompressHandle, middleware.GzipDecompressMiddleware)

	// Регистрация профилировщика Chi для мониторинга и отладки.
	router.Mount("/debug", chiMiddleware.Profiler())

	// Создание нового обработчика запросов.
	handler, err := handlers.NewHandler(log, conf, db)
	if err != nil {
		return fmt.Errorf("handler creation error: %w", err)
	}

	// Регистрация обработчика запросов.
	handler.Register(router, middleware)

	// Запуск HTTP-сервера.
	err = http.ListenAndServe(conf.Address, router)
	if err != nil {
		return fmt.Errorf("server serving error: %w", err)
	}

	return nil
}
