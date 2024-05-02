package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/GTedya/shortener/internal/app/storage/dbstorage"
	"github.com/GTedya/shortener/internal/app/tokenutils"
)

// getURLByID получает оригинальный URL по его сокращенной версии.
func (h *handler) getURLByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	shortenURL, err := h.store.GetURL(r.Context(), id)
	if err != nil && errors.Is(err, dbstorage.ErrDeletedURL) {
		w.WriteHeader(http.StatusGone)
		return
	}
	if err != nil {
		h.log.Errorw("ID not found", id, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Add(contentType, "text/plain; application/json")

	http.Redirect(w, r, shortenURL, http.StatusTemporaryRedirect)
}

// userUrls получает список сокращенных URL, принадлежащих текущему пользователю.
func (h *handler) userURLS(w http.ResponseWriter, r *http.Request) {
	userID := tokenutils.GetUserID(r)
	w.Header().Add(contentType, appJSON)

	urls, err := h.db.UserURLS(r.Context(), userID)
	if err != nil {
		h.log.Errorw("URL getting error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(urls) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	marshal, err := json.Marshal(urls)
	if err != nil {
		h.log.Errorw("Json marshalling error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	h.log.Debug(urls)

	_, err = w.Write(marshal)
	if err != nil {
		h.log.Errorw("data writing error:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}
