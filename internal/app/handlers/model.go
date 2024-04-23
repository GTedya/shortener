package handlers

import (
	"github.com/GTedya/shortener/internal/app/middlewares"
	"github.com/go-chi/chi/v5"
)

// Handler представляет интерфейс для обработчика HTTP-запросов.
type Handler interface {
	// Register регистрирует обработчики маршрутов HTTP в маршрутизаторе chi.
	Register(router *chi.Mux, middleware middlewares.Middleware)
}
