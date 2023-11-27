package handlers

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/GTedya/shortener/config"
	"github.com/GTedya/shortener/internal/app/logger"
	"github.com/GTedya/shortener/internal/helpers"
	"github.com/go-chi/chi/v5"
)

type handler struct {
}

const urlLen = 6
const contentType = "Content-Type"

func NewHandler() Handler {
	return &handler{}
}

func (h *handler) Register(router *chi.Mux, conf config.Config) {
	data := helpers.CreateURLMap(conf.FileStoragePath)

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
	body, err := io.ReadAll(r.Body)
	log := logger.CreateLogger()

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(body) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	content := r.Header.Get(contentType)

	if strings.Contains(content, "application/x-gzip") {
		reader, err := gzip.NewReader(bytes.NewReader(body))
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		body, err = io.ReadAll(reader)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	id := conf.URL + helpers.GenerateURL(urlLen)
	encodedID := helpers.ShortURL{URL: url.PathEscape(id)}
	originalURL := helpers.URL{URL: string(body)}
	data.URLMap[encodedID] = originalURL

	if conf.FileStoragePath != "" {
		jsonFile := helpers.FileStorage{
			UUID:        helpers.GenerateUUID(conf.FileStoragePath),
			ShortURL:    encodedID.URL,
			OriginalURL: originalURL.URL,
		}
		err = helpers.AppendToFile(conf.FileStoragePath, jsonFile)
		if err != nil {
			log.Info(err)
		}
	}

	w.Header().Add(contentType, "text/plain; application/json")
	w.WriteHeader(http.StatusCreated)

	if _, err := w.Write([]byte(fmt.Sprintf("http://%s/%s", conf.Address, encodedID.URL))); err != nil {
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

func (h *handler) URLByJSON(w http.ResponseWriter, r *http.Request, conf config.Config, data *helpers.URLData) {
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
	log := logger.CreateLogger()

	var u helpers.URL
	err = json.Unmarshal(body, &u)
	if err != nil {
		log.Error(err)
	}
	id := conf.URL + helpers.GenerateURL(urlLen)

	encodedID := helpers.ShortURL{URL: url.PathEscape(id)}
	data.URLMap[encodedID] = u
	if conf.FileStoragePath != "" {
		jsonFile := helpers.FileStorage{
			UUID:        helpers.GenerateUUID(conf.FileStoragePath),
			ShortURL:    encodedID.URL,
			OriginalURL: u.URL,
		}
		err = helpers.AppendToFile(conf.FileStoragePath, jsonFile)
		if err != nil {
			log.Info(err)
		}
	}
	w.Header().Set(contentType, "application/json")
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
