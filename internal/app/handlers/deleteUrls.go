package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"
)

func (h *handler) deleteUrls(w http.ResponseWriter, r *http.Request) {
	token, err := r.Cookie("token")
	if err != nil {
		h.log.Errorw("token receiving error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

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
		return
	}
	w.WriteHeader(http.StatusAccepted)

	ctx := r.Context()
	gen := generator(ctx, shortURLs)
	out := make(chan string)
	var wg sync.WaitGroup

	go func() {
		defer close(out)

		for url := range gen {
			isUser, er := h.db.IsUserURL(ctx, token.Value, url)
			if er != nil {
				h.log.Errorw("Is user error", er)
				return
			}
			if isUser {
				out <- url
			}
		}
	}()

	err = h.db.DeleteURLS(ctx, &wg, out)
	if err != nil {
		h.log.Errorw("User deleting error", err)
		return
	}

	wg.Wait()
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
