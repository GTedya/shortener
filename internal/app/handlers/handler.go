package handlers

import (
	"fmt"
	"github.com/GTedya/shortener/config"
	"github.com/GTedya/shortener/internal/helpers"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"net/url"
)

type handler struct {
}

func NewHandler() Handler {
	return &handler{}
}

func (h *handler) Register(router *chi.Mux, conf config.Config) {
	data := helpers.URLData{
		URLMap: make(map[string]string),
	}
	router.Post("/", func(writer http.ResponseWriter, request *http.Request) {
		h.CreateURL(writer, request, conf, &data)
	})

	router.Get("/{id}", func(writer http.ResponseWriter, request *http.Request) {
		h.GetURLByID(writer, request, data)
	})
}

func (h *handler) CreateURL(w http.ResponseWriter, r *http.Request, conf config.Config, data *helpers.URLData) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "text/plain; charset=utf-8" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(r.Body)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(body) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	id := conf.URL + helpers.GenerateURL(6)
	encodedID := url.PathEscape(id)

	data.URLMap[encodedID] = string(body)

	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)

	_, err = w.Write([]byte(fmt.Sprintf("http://%s/%s", conf.Address, encodedID)))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (h *handler) GetURLByID(w http.ResponseWriter, r *http.Request, data helpers.URLData) {
	id := chi.URLParam(r, "id")

	shortenURL, err := data.GetByShortenURL(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Add("Content-Type", "text/plain")
	http.Redirect(w, r, shortenURL, http.StatusTemporaryRedirect)
}
