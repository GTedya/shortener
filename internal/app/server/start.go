package server

import (
	"github.com/GTedya/shortener/config"
	"github.com/GTedya/shortener/internal/app/handlers"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

func Start(conf config.Config) error {
	router := chi.NewRouter()

	handler := handlers.NewHandler()
	handler.Register(router, conf)

	err := http.ListenAndServe(conf.Address, router)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}
