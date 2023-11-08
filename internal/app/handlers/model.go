package handlers

import (
	"github.com/GTedya/shortener/config"
	"github.com/go-chi/chi/v5"
)

type Handler interface {
	Register(router *chi.Mux, conf config.Config)
}
