package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"

	"github.com/GTedya/shortener/internal/app/models"
	"github.com/GTedya/shortener/internal/app/repository"
	"github.com/GTedya/shortener/internal/app/tokenutils"
)

// ReqMultipleURL представляет структуру запроса для создания нескольких коротких URL.
type ReqMultipleURL struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// ResMultipleURL представляет структуру ответа с коротким URL и соответствующим ID запроса.
type ResMultipleURL struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// errResponseWrite представляет ошибку записи данных.
var errResponseWrite = errors.New("data writing error")

// errJSONMarshal представляет ошибку маршалинга JSON.
var errJSONMarshal = errors.New("json marshalling error")

// URL представляет структуру для хранения URL.
type URL struct {
	URL string `json:"url"`
}

// ShortURL представляет структуру для хранения сокращенного URL.
type ShortURL struct {
	URL string `json:"result"`
}

// createURL обрабатывает запрос на создание сокращенного URL.
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

	userID := tokenutils.GetUserID(r)
	err = h.repo.Save(r.Context(), models.ShortURL{
		OriginalURL: id,
		ShortURL:    shortID,
		CreatedByID: userID,
	})

	if errors.Is(err, repository.ErrDuplicate) {
		w.WriteHeader(http.StatusConflict)
		shortURL, err := h.repo.ShortenByURL(r.Context(), id)
		if err != nil {
			h.log.Errorw("short url getting error", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if _, err = w.Write([]byte(fmt.Sprintf("%s/%s", h.conf.URL, shortURL.ShortURL))); err != nil {
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

	if err = tokenutils.AddEncryptedUserIDToCookie(&w, userID); err != nil {
		h.log.Errorw("adding cookie error", err)
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

// urlByJSON обрабатывает запрос на создание сокращенного URL, переданный в формате JSON.
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

	token := r.Header.Get("Authorization")
	err = h.repo.Save(r.Context(), models.ShortURL{
		OriginalURL: id,
		ShortURL:    shortID,
		CreatedByID: token,
	})

	if errors.Is(err, repository.ErrDuplicate) {
		w.WriteHeader(http.StatusConflict)
		shortURL, err := h.repo.ShortenByURL(r.Context(), id)
		if err != nil {
			h.log.Errorw("short url getting error", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		encodedID := ShortURL{URL: fmt.Sprintf("http://%s/%s", h.conf.Address, shortURL.ShortURL)}
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

// batch обрабатывает запрос на пакетное создание сокращенных URL.
func (h *handler) batch(w http.ResponseWriter, r *http.Request) {
	content := r.Header.Get(contentType)
	if content != appJSON {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var reqUrls []ReqMultipleURL
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(body) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resUrls := make([]ResMultipleURL, 0)
	urls := make([]models.ShortURL, 0)

	err = json.Unmarshal(body, &reqUrls)
	if err != nil {
		h.log.Errorw("Json unmarshalling error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	userID := tokenutils.GetUserID(r)

	for _, url := range reqUrls {
		if len(url.OriginalURL) == 0 {
			break
		}
		shortID := uuid.NewString()
		res := ResMultipleURL{
			CorrelationID: url.CorrelationID,
			ShortURL:      fmt.Sprintf("http://%s/%s", h.conf.Address, shortID),
		}

		resUrls = append(resUrls, res)
		urls = append(urls, models.ShortURL{
			OriginalURL: url.OriginalURL,
			ShortURL:    shortID,
			CreatedByID: userID,
		})
	}

	err = h.repo.SaveBatch(r.Context(), urls)
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
