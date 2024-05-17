package middlewares

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// gzipWriter представляет обертку над http.ResponseWriter для поддержки сжатия gzip.
type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

// Write перенаправляет запись внутреннему Writer'у сжатия.
func (w gzipWriter) Write(b []byte) (int, error) {
	writer, err := w.Writer.Write(b)
	if err != nil {
		return 0, fmt.Errorf("error in gzipWriter method Write: %w", err)
	}
	return writer, nil
}

// GzipCompressHandle возвращает HTTP-обработчик,
// который сжимает ответ с использованием gzip, если клиент поддерживает сжатие.
// Если клиент не поддерживает сжатие или тип содержимого не является application/json или text/html,
// запрос передается следующему обработчику без изменений.
func (m Middleware) GzipCompressHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		content := r.Header.Get("Content-Type")

		if !strings.Contains(content, "application/json") && !strings.Contains(content, "text/html") {
			next.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			m.Log.Errorw("gzip writer error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer func() {
			if err := gz.Close(); err != nil {
				m.Log.Errorw("gzip close error", err)
			}
		}()

		w.Header().Add("Content-Encoding", "gzip")
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
}
