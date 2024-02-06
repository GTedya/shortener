package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func (h *handler) deleteUrls(w http.ResponseWriter, r *http.Request) {
	token, err := r.Cookie("token")
	if err != nil {
		h.log.Errorw("token receiving error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Add(contentType, appJSON)

	var shortURLS []string
	body, err := io.ReadAll(r.Body)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(body) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(body, &shortURLS)
	if err != nil {
		h.log.Errorw("Json unmarshalling error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	gen := generator(ctx, shortURLS)
	out := make(chan string)

	go func() {
		defer close(out)
		for url := range gen {
			isUser, err := h.db.IsUserURL(ctx, token.Value, url)
			if err != nil {
				h.log.Errorw("Is user error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if isUser {
				out <- url
			}
		}
	}()

	err = h.db.DeleteURLS(r.Context(), out)
	if err != nil {
		h.log.Errorw("User deleting error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func generator(ctx context.Context, input []string) chan string {
	inputCh := make(chan string)

	go func() {
		defer close(inputCh)

		for _, data := range input {
			select {
			case <-ctx.Done():
				return
			case inputCh <- data:
			}
		}
	}()

	return inputCh
}
