package middlewares

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"strings"
)

// GzipDecompressMiddleware представляет middleware для декомпрессии тела запроса, если оно было сжато с использованием gzip.
// Если заголовок Content-Type указывает на тип application/x-gzip, тело запроса декомпримируется из формата gzip.
// Распакованное тело заменяет исходное тело запроса, а затем запрос передается следующему обработчику.
func (m Middleware) GzipDecompressMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		if strings.Contains(contentType, "application/x-gzip") {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				m.Log.Errorw("Error reading request body", err)
				return
			}

			reader, err := gzip.NewReader(bytes.NewReader(body))
			if err != nil && !errors.Is(err, io.EOF) {
				m.Log.Errorw("Error creating gzip reader", err)
				return
			}

			decodedBody, err := io.ReadAll(reader)
			if err != nil && !errors.Is(err, io.EOF) {
				m.Log.Errorw("Error reading decoded body", err)
				return
			}

			r.Body = io.NopCloser(bytes.NewReader(decodedBody))
		}
		next.ServeHTTP(w, r)
	})
}
