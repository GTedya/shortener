package handlers

import (
	"database/sql"
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
	Log *zap.SugaredLogger
}

const urlLen = 6
const contentType = "Content-Type"

func NewHandler(logger *zap.SugaredLogger) Handler {
	return &handler{Log: logger}
}

func (h *handler) Register(router *chi.Mux, conf config.Config, db *sql.DB) {
	data := helpers.CreateURLMap(conf.FileStoragePath, h.Log)
	router.Post("/", func(writer http.ResponseWriter, request *http.Request) {
		h.CreateURL(writer, request, conf, &data)
	})

	router.Get("/{id}", func(writer http.ResponseWriter, request *http.Request) {
		h.GetURLByID(writer, request, data)
	})

	router.Post("/api/shorten", func(writer http.ResponseWriter, request *http.Request) {
		h.URLByJSON(writer, request, conf, &data)
	})

	router.Get("/ping", func(writer http.ResponseWriter, request *http.Request) {
		h.GetPing(writer, request, db)
	})
}

func (h *handler) CreateURL(w http.ResponseWriter, r *http.Request,
	conf config.Config, data *helpers.URLData) {
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

	shortID := helpers.CreateUniqueID(*data, urlLen)

	store := helpers.NewStore(conf, h.Log)
	store.Store(id, shortID, data)

	w.Header().Add(contentType, "text/plain; application/json")
	w.WriteHeader(http.StatusCreated)

	if _, err := w.Write([]byte(fmt.Sprintf("%s/%s", conf.URL, shortID))); err != nil {
		h.Log.Errorw("data writing error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *handler) GetURLByID(w http.ResponseWriter, r *http.Request, data helpers.URLData) {
	id := chi.URLParam(r, "id")

	shortenURL, err := data.GetByShortenURL(id)
	if err != nil {
		h.Log.Errorw("ID not found", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Add(contentType, "text/plain; application/json")

	http.Redirect(w, r, shortenURL.URL, http.StatusTemporaryRedirect)
}

func (h *handler) URLByJSON(w http.ResponseWriter, r *http.Request,
	conf config.Config, data *helpers.URLData) {
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
		h.Log.Errorw("Json unmarshalling error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	id := u.URL
	shortID := helpers.CreateUniqueID(*data, urlLen)

	store := helpers.NewStore(conf, h.Log)
	store.Store(id, shortID, data)

	w.Header().Add(contentType, "application/json")

	encodedID := helpers.ShortURL{URL: fmt.Sprintf("http://%s/%s", conf.Address, shortID)}
	marshal, err := json.Marshal(encodedID)
	if err != nil {
		h.Log.Errorw("Json marshalling error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(marshal)
	if err != nil {
		h.Log.Errorw("data writing error:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (h *handler) GetPing(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	err := db.Ping()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
