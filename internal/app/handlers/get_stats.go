package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/GTedya/shortener/internal/app/models"
)

func (h *handler) getStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userCount, urlsCount, err := h.repo.GetUsersAndUrlsCount(r.Context())
	if err != nil {
		h.log.Errorw("stats getting error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	marshal, err := json.Marshal(models.Stats{
		UrlsCount:  urlsCount,
		UsersCount: userCount,
	})
	if err != nil {
		h.log.Error(errJSONMarshal)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

	_, err = w.Write(marshal)
	if err != nil {
		h.log.Errorw("data writing error:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}
