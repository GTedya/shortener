package handlers

import (
	"github.com/GTedya/shortener/internal/app/middlewares"
	"github.com/go-chi/chi/v5"
)

type Handler interface {
	Register(router *chi.Mux, middleware middlewares.Middleware)
}
