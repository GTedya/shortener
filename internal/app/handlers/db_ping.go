package handlers

import "net/http"

// getPing выполняет проверку доступности базы данных.
func (h *handler) getPing(w http.ResponseWriter, r *http.Request) {
	err := h.repo.Check(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
