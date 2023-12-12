package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/GTedya/shortener/config"
	"github.com/GTedya/shortener/internal/helpers"
	"github.com/go-chi/chi/v5"
)

type handler struct {
	log   *zap.SugaredLogger
	store helpers.Store
	conf  config.Config
}

const urlLen = 6
const contentType = "Content-Type"

type Store interface {
	GetURL(shortID string) (string, error)
	SaveURL(id, shortID string) error
}

func NewHandler(logger *zap.SugaredLogger, conf config.Config) (Handler, error) {
	data, err := helpers.CreateURLData(conf.FileStoragePath)
	if err != nil {
		return nil, fmt.Errorf("data creation error: %w", err)
	}
	return &handler{log: logger, store: helpers.NewStore(conf, data)}, nil
}

func (h *handler) Register(router *chi.Mux) {
	router.Post("/", func(writer http.ResponseWriter, request *http.Request) {
		h.CreateURL(writer, request)
	})

	router.Get("/{id}", func(writer http.ResponseWriter, request *http.Request) {
		h.GetURLByID(writer, request)
	})

	router.Post("/api/shorten", func(writer http.ResponseWriter, request *http.Request) {
		h.URLByJSON(writer, request)
	})
}

func (h *handler) CreateURL(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(body) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	id := string(body)

	shortID := helpers.CreateUniqueID(h.store.GetURL, urlLen)

	err = h.store.SaveURL(id, shortID)
	if err != nil {
		h.log.Errorw("data saving error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Add(contentType, "text/plain; application/json")
	w.WriteHeader(http.StatusCreated)

	if _, err := w.Write([]byte(fmt.Sprintf("%s/%s", h.conf.URL, shortID))); err != nil {
		h.log.Errorw("data writing error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *handler) GetURLByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	shortenURL, err := h.store.GetURL(id)
	if err != nil {
		h.log.Errorw("ID not found", id, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Add(contentType, "text/plain; application/json")

	http.Redirect(w, r, shortenURL, http.StatusTemporaryRedirect)
}

func (h *handler) URLByJSON(w http.ResponseWriter, r *http.Request) {
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
		h.log.Errorw("Json unmarshalling error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	id := u.URL
	shortID := helpers.CreateUniqueID(h.store.GetURL, urlLen)

	err = h.store.SaveURL(id, shortID)
	if err != nil {
		h.log.Errorw("data saving error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Add(contentType, "application/json")

	encodedID := helpers.ShortURL{URL: fmt.Sprintf("http://%s/%s", h.conf.Address, shortID)}
	marshal, err := json.Marshal(encodedID)
	if err != nil {
		h.log.Errorw("Json marshalling error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(marshal)
	if err != nil {
		h.log.Errorw("data writing error:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}
