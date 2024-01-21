package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/GTedya/shortener/internal/app/storage/dbstorage"
	"io"
	"net/http"

	"github.com/GTedya/shortener/config"
	"github.com/GTedya/shortener/database"
	"github.com/GTedya/shortener/internal/app/storage"
	"github.com/go-chi/chi/v5"
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
	GetURL(ctx context.Context, shortID string) (string, error)
	SaveURL(ctx context.Context, id, shortID string) error
	Batch(ctx context.Context, urls map[string]string) error
}

func NewHandler(logger *zap.SugaredLogger, conf config.Config, db *database.DB) (Handler, error) {
	store, err := storage.NewStore(conf, db)
	if err != nil {
		return nil, fmt.Errorf("store creation error: %w", err)
	}
	return &handler{log: logger, conf: conf, store: store, db: db}, nil
}

func (h *handler) Register(router *chi.Mux) {
	router.Post("/", h.CreateURL)

	router.Get("/{id}", h.GetURLByID)

	router.Post("/api/shorten", h.URLByJSON)

	router.Get("/ping", h.GetPing)

	router.Post("/api/shorten/batch", h.Batch)
}

func (h *handler) CreateURL(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	var shortID string
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
	shortID = createUniqueID(r.Context(), h.store.GetURL, urlLen)

	err = h.store.SaveURL(r.Context(), id, shortID)

	if errors.Is(err, dbstorage.ErrDuplicate) {
		w.WriteHeader(http.StatusConflict)
		shortID, err = h.db.GetShortURL(r.Context(), id)
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

	shortenURL, err := h.store.GetURL(r.Context(), id)
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

	w.Header().Set(contentType, appJSON)

	var u URL
	err = json.Unmarshal(body, &u)
	if err != nil {
		h.log.Errorw("Json unmarshalling error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	id := u.URL
	shortID := createUniqueID(r.Context(), h.store.GetURL, urlLen)

	err = h.store.SaveURL(r.Context(), id, shortID)

	if errors.Is(err, dbstorage.ErrDuplicate) {
		w.WriteHeader(http.StatusConflict)
		shortID, err = h.db.GetShortURL(r.Context(), id)
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
	err := h.db.Ping(r.Context())
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
	var reqUrls []storage.ReqMultipleURL
	body, err := io.ReadAll(r.Body)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(body) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	resUrls := make([]storage.ResMultipleURL, 0)
	urls := make(map[string]string)

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
		shortID := createUniqueID(r.Context(), h.store.GetURL, urlLen)
		res := storage.ResMultipleURL{CorrelationID: url.CorrelationID,
			ShortURL: fmt.Sprintf("http://%s/%s", h.conf.Address, shortID)}

		resUrls = append(resUrls, res)
		urls[url.OriginalURL] = shortID
	}

	err = h.store.Batch(r.Context(), urls)
	if err != nil {
		h.log.Errorw("data saving error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
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
