package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/GTedya/shortener/internal/app/models"
	"github.com/GTedya/shortener/internal/app/tokenutils"
)

// deleteUrls обрабатывает запрос на удаление сокращенных URL, принадлежащих пользователю.
func (h *handler) deleteUrls(w http.ResponseWriter, r *http.Request) {
	token := tokenutils.GetUserID(r)

	var shortURLs []string
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(body) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(body, &shortURLs)
	if err != nil {
		h.log.Errorw("JSON unmarshalling error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)

	ctx := r.Context()

	urls := make([]models.ShortURL, 0)

	for _, shortURL := range shortURLs {
		urls = append(urls, models.ShortURL{
			ShortURL:    shortURL,
			CreatedByID: token,
		})
	}

	err = h.repo.DeleteUrls(ctx, urls)
	if err != nil {
		h.log.Errorw("User deleting error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
