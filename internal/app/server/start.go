package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/GTedya/shortener/config"
	"github.com/GTedya/shortener/database"
	"github.com/GTedya/shortener/internal/app/handlers"
	"github.com/GTedya/shortener/internal/app/middlewares"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/kabukky/httpscerts"
	"go.uber.org/zap"
)

// Start запускает HTTP-сервер для обслуживания запросов.
// Принимает конфигурацию сервера, логгер и экземпляр базы данных.
// Возвращает ошибку, если сервер не удалось запустить.
func Start(conf config.Config, log *zap.SugaredLogger, db database.DB) error {
	// Создание нового маршрутизатора Chi.
	router := chi.NewRouter()

	// Создание экземпляра посредников.
	middle := middlewares.Middleware{Log: log, SecretKey: conf.SecretKey}

	// Использование посредников для обработки логов, сжатия gzip и декомпрессии gzip.
	router.Use(middle.LogHandle, middle.GzipCompressHandle, middle.GzipDecompressMiddleware)

	// Регистрация профилировщика Chi для мониторинга и отладки.
	router.Mount("/debug", chiMiddleware.Profiler())

	// Создание нового обработчика запросов.
	handler, err := handlers.NewHandler(log, conf, db)
	if err != nil {
		return fmt.Errorf("handler creation error: %w", err)
	}

	// Регистрация обработчика запросов.
	handler.Register(router, middle)

	sigs := make(chan os.Signal, 1)
	idleConnsClosed := make(chan struct{})
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	srv := &http.Server{Addr: conf.Address, Handler: router}

	go handleShutdown(sigs, srv, log, idleConnsClosed)

	// Запуск HTTPS-сервера.
	if conf.EnableHTTPS {
		return runHTTPS(log, srv)
	}

	// Запуск HTTP-сервера.
	if err = srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server serving error: %w", err)
	}

	<-idleConnsClosed
	return nil
}

// handleShutdown gracefully shuts down the HTTP server upon receiving a signal.
// It waits for a signal from the provided channel, then attempts to gracefully
// shut down the server. If the shutdown encounters any errors, they are logged.
// Once the server is shut down, the idleConnsClosed channel is closed.
//
// Parameters:
//   - sigs: A channel to receive OS signals for initiating shutdown.
//   - srv: The HTTP server to shut down.
//   - log: A sugared logger for logging messages.
//   - idleConnsClosed: A channel to signal that all connections are closed.
func handleShutdown(sigs chan os.Signal, srv *http.Server, log *zap.SugaredLogger, idleConnsClosed chan struct{}) {
	<-sigs
	if errs := srv.Shutdown(context.Background()); errs != nil {
		log.Errorw("HTTP server Shutdown: %v", errs)
	}
	log.Debug("Shutting down HTTP server")
	close(idleConnsClosed)
}

// runHTTPS starts the HTTPS server with the provided server configuration.
// If the required certificate and key files are not found, it attempts to
// generate them using the provided address. The server listens and serves
// HTTPS requests using the generated or existing certificate and key files.
//
// Parameters:
//   - log: A sugared logger for logging messages.
//   - srv: The HTTP server to start.
//
// Returns:
//   - error: An error if the server fails to start or serve.
func runHTTPS(log *zap.SugaredLogger, srv *http.Server) error {
	if err := httpscerts.Check("cert.pem", "key.pem"); err != nil {
		if err := httpscerts.Generate("cert.pem", "key.pem", "127.0.0.1:8081"); err != nil {
			log.Fatal("Ошибка: Не можем сгенерировать https сертификат.")
		}
	}

	if err := srv.ListenAndServeTLS("cert.pem", "key.pem"); err != nil {
		return fmt.Errorf("server serving error: %w", err)
	}
	return nil
}
