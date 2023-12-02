package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/GTedya/shortener/config"
	"github.com/GTedya/shortener/internal/app/logger"
	"github.com/GTedya/shortener/internal/helpers"
	"github.com/go-chi/chi/v5"
)

type handler struct {
}

const contentType = "Content-Type"

func NewHandler() Handler {
	return &handler{}
}

func (h *handler) Register(router *chi.Mux, conf config.Config) {
	data := helpers.CreateURLMap(conf.FileStoragePath)
	log := logger.CreateLogger()

	router.Post("/", func(writer http.ResponseWriter, request *http.Request) {
		h.CreateURL(writer, request, conf, &data, log)
	})

	router.Get("/{id}", func(writer http.ResponseWriter, request *http.Request) {
		h.GetURLByID(writer, request, data)
	})

	router.Post("/api/shorten", func(writer http.ResponseWriter, request *http.Request) {
		h.URLByJSON(writer, request, conf, &data, log)
	})
}

func (h *handler) CreateURL(w http.ResponseWriter, r *http.Request,
	conf config.Config, data *helpers.URLData, log *zap.SugaredLogger) {
	body, err := io.ReadAll(r.Body)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(body) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var store string
	var u = helpers.URL{URL: string(body)}

	if conf.FileStoragePath != "" {
		store = helpers.FileStore(data, u, conf.URL, conf.FileStoragePath)
	} else {
		store = helpers.MemoryStore(data, u, conf.URL).URL
	}
	w.Header().Add(contentType, "text/plain; application/json")
	w.WriteHeader(http.StatusCreated)

	if _, err := w.Write([]byte(fmt.Sprintf("http://%s/%s", conf.Address, store))); err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
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

	w.Header().Add(contentType, "text/plain; application/json")

	http.Redirect(w, r, shortenURL.URL, http.StatusTemporaryRedirect)
}

func (h *handler) URLByJSON(w http.ResponseWriter, r *http.Request,
	conf config.Config, data *helpers.URLData, log *zap.SugaredLogger) {
	content := r.Header.Get(contentType)
	if content != "application/json" {
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
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var store string

	if conf.FileStoragePath != "" {
		store = helpers.FileStore(data, u, conf.URL, conf.FileStoragePath)
	} else {
		store = helpers.MemoryStore(data, u, conf.URL).URL
	}
	w.Header().Add(contentType, "application/json")

	encodedID := helpers.ShortURL{URL: fmt.Sprintf("http://%s/%s", conf.Address, store)}
	marshal, err := json.Marshal(encodedID)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(marshal)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}
