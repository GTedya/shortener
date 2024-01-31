package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *handler) getURLByID(w http.ResponseWriter, r *http.Request) {
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

func (h *handler) userUrls(w http.ResponseWriter, r *http.Request) {
	token, err := r.Cookie("token")
	if err != nil {
		h.log.Errorw("token receiving error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Add(contentType, appJSON)

	urls, err := h.db.UserURLS(r.Context(), token.Value)
	if err != nil {
		h.log.Errorw("URL getting error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(urls) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	for _, url := range urls {
		url.ShortURL = h.conf.URL + url.ShortURL
	}

	marshal, err := json.Marshal(urls)
	if err != nil {
		h.log.Errorw("Json marshalling error", err)
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