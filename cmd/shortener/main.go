package main

import (
	"github.com/GTedya/shortener/config"
	"github.com/GTedya/shortener/internal/app"
	"github.com/GTedya/shortener/internal/helpers"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

func main() {
	router := chi.NewRouter()

	conf := new(config.Config)
	conf.AnnounceConfig()

	data := helpers.URLData{
		URLMap: make(map[string]string),
	}

	router.Post("/", func(writer http.ResponseWriter, request *http.Request) {
		app.CreateURL(writer, request, conf, &data)
	})

	router.Get("/{id}", func(writer http.ResponseWriter, request *http.Request) {
		app.GetURLByID(writer, request, data)
	})

	err := http.ListenAndServe(conf.Address, router)
	if err != nil {
		log.Fatal(err)
	}
}
