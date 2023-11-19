package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/GTedya/shortener/config"
	"github.com/GTedya/shortener/internal/app/logger"
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
		URLMap: make(map[helpers.ShortURL]helpers.URL),
	}
	router.Post("/", func(writer http.ResponseWriter, request *http.Request) {
		h.CreateURL(writer, request, conf, &data)
	})

	router.Get("/{id}", func(writer http.ResponseWriter, request *http.Request) {
		h.GetURLByID(writer, request, data)
	})

	router.Post("/api/shorten", func(writer http.ResponseWriter, request *http.Request) {
		h.URLByJSON(writer, request, conf, &data)
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
	encodedID := helpers.ShortURL{URL: url.PathEscape(id)}

	data.URLMap[encodedID] = helpers.URL{URL: string(body)}

	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)

	_, err = w.Write([]byte(fmt.Sprintf("http://%s/%s", conf.Address, encodedID.URL)))
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
	http.Redirect(w, r, shortenURL.URL, http.StatusTemporaryRedirect)
}

func (h *handler) URLByJSON(w http.ResponseWriter, r *http.Request, conf config.Config, data *helpers.URLData) {
	log := logger.CreateLogger()

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
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
	var u helpers.URL
	err = json.Unmarshal(body, &u)
	if err != nil {
		log.Fatal(err)
	}
	id := conf.URL + helpers.GenerateURL(6)

	encodedID := helpers.ShortURL{URL: url.PathEscape(id)}
	data.URLMap[encodedID] = u

	w.WriteHeader(http.StatusCreated)

	encodedID = helpers.ShortURL{URL: fmt.Sprintf("http://%s/%s", conf.Address, url.PathEscape(id))}
	marshal, err := json.Marshal(encodedID)
	if err != nil {
		return
	}

	_, err = w.Write(marshal)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}
