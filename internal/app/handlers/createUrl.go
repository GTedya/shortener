package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/GTedya/shortener/internal/app/middlewares"
	"io"
	"net/http"

	"github.com/google/uuid"

	"github.com/GTedya/shortener/internal/app/storage"
	"github.com/GTedya/shortener/internal/app/storage/dbstorage"
)

var errResponseWrite = errors.New("data writing error")
var errJSONMarshal = errors.New("json marshalling error")

type URL struct {
	URL string `json:"url"`
}

type ShortURL struct {
	URL string `json:"result"`
}

func (h *handler) createURL(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(body) == 0 {
		h.log.Debug("empty request body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var shortID string

	id := string(body)
	w.Header().Add(contentType, "text/plain; application/json")
	shortID = uuid.NewString()

	token, err := middlewares.TokenCreate()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	http.SetCookie(w, token)

	ctx := context.WithValue(r.Context(), "token", token.Value)

	err = h.store.SaveURL(ctx, id, shortID)

	if errors.Is(err, dbstorage.ErrDuplicate) {
		w.WriteHeader(http.StatusConflict)
		shortID, err = h.db.GetShortURL(r.Context(), id)
		if err != nil {
			h.log.Errorw("short url getting error", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if _, err = w.Write([]byte(fmt.Sprintf("%s/%s", h.conf.URL, shortID))); err != nil {
			h.log.Error(errResponseWrite)
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
		h.log.Error(errResponseWrite)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *handler) urlByJSON(w http.ResponseWriter, r *http.Request) {
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
		h.log.Debug("empty request body")
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
	shortID := uuid.NewString()

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
			h.log.Error(errJSONMarshal)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, err = w.Write(marshal)
		if err != nil {
			h.log.Error(errResponseWrite)
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
		h.log.Error(errJSONMarshal)
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

func (h *handler) batch(w http.ResponseWriter, r *http.Request) {
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
		shortID := uuid.NewString()
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
		h.log.Error(errJSONMarshal)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Add(contentType, appJSON)

	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(marshal)
	if err != nil {
		h.log.Error(errResponseWrite)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}
