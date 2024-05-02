package handlers

import (
	"context"
	"encoding/json"
	"github.com/GTedya/shortener/internal/app/tokenutils"
	"io"
	"net/http"
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
		return
	}
	w.WriteHeader(http.StatusAccepted)

	ctx := context.Background()
	gen := generator(ctx, shortURLs)

	err = h.db.DeleteURLS(ctx, token, gen)
	if err != nil {
		h.log.Errorw("User deleting error", err)
		return
	}
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
