package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/GTedya/shortener/config"
	"github.com/GTedya/shortener/database"
	"github.com/GTedya/shortener/internal/app/datastore"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

type handler struct {
	log   *zap.SugaredLogger
	db    *database.DB
	store Store
	conf  config.Config
}

const urlLen = 6
const contentType = "Content-Type"
const appJSON = "application/json"

type Store interface {
	GetURL(shortID string) (string, error)
	SaveURL(id, shortID string) error
}

type reqMultipleURL struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type resMultipleURL struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

func NewHandler(logger *zap.SugaredLogger, conf config.Config, db *database.DB) (Handler, error) {
	store, err := datastore.NewStore(conf, db)
	if err != nil {
		return nil, fmt.Errorf("store creation error: %w", err)
	}
	return &handler{log: logger, conf: conf, store: store, db: db}, nil
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

	router.Get("/ping", h.GetPing)

	router.Post("/api/shorten/batch", h.Batch)
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
	w.Header().Add(contentType, "text/plain; application/json")

	shortID := createUniqueID(h.store.GetURL, urlLen)

	err = h.store.SaveURL(id, shortID)

	var pqError *pgconn.PgError

	if errors.As(err, &pqError) {
		w.WriteHeader(http.StatusConflict)
		shortID, err = h.db.GetShortURL(id)
		if err != nil {
			h.log.Errorw("short url getting error", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if _, err = w.Write([]byte(fmt.Sprintf("%s/%s", h.conf.URL, shortID))); err != nil {
			h.log.Errorw("data writing error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	}

	if err != nil {
		h.log.Errorw("data saving error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)

	if _, err = w.Write([]byte(fmt.Sprintf("%s/%s", h.conf.URL, shortID))); err != nil {
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
	if content != appJSON {
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

	var u URL
	err = json.Unmarshal(body, &u)
	if err != nil {
		h.log.Errorw("Json unmarshalling error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	id := u.URL
	shortID := createUniqueID(h.store.GetURL, urlLen)

	err = h.store.SaveURL(id, shortID)

	var pqError *pgconn.PgError

	if errors.As(err, &pqError) {
		w.WriteHeader(http.StatusConflict)
		shortID, err = h.db.GetShortURL(id)
		if err != nil {
			h.log.Errorw("short url getting error", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		encodedID := ShortURL{URL: fmt.Sprintf("http://%s/%s", h.conf.Address, shortID)}
		marshal, err := json.Marshal(encodedID)
		if err != nil {
			h.log.Errorw("Json marshalling error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, err = w.Write(marshal)
		if err != nil {
			h.log.Errorw("data writing error:", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		return
	}
	if err != nil {
		h.log.Errorw("data saving error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Add(contentType, appJSON)

	encodedID := ShortURL{URL: fmt.Sprintf("http://%s/%s", h.conf.Address, shortID)}
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

func (h *handler) GetPing(w http.ResponseWriter, r *http.Request) {
	err := h.db.Ping(context.TODO())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *handler) Batch(w http.ResponseWriter, r *http.Request) {
	content := r.Header.Get(contentType)
	if content != appJSON {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var reqUrls []reqMultipleURL
	body, err := io.ReadAll(r.Body)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(body) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	resUrls := make([]resMultipleURL, 0)

	err = json.Unmarshal(body, &reqUrls)
	if err != nil {
		h.log.Errorw("Json unmarshalling error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for _, url := range reqUrls {
		if len(url.OriginalURL) == 0 {
			break
		}
		shortID := createUniqueID(h.store.GetURL, urlLen)
		res := resMultipleURL{CorrelationID: url.CorrelationID,
			ShortURL: fmt.Sprintf("http://%s/%s", h.conf.Address, shortID)}
		err = h.store.SaveURL(url.OriginalURL, shortID)
		if err != nil {
			h.log.Errorw("data saving error", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		resUrls = append(resUrls, res)
	}

	marshal, err := json.Marshal(resUrls)
	if err != nil {
		h.log.Errorw("Json marshalling error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Add(contentType, appJSON)

	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(marshal)
	if err != nil {
		h.log.Errorw("data writing error:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}
