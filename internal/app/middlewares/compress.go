package middlewares

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/GTedya/shortener/internal/app/logger"
)

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	writer, err := w.Writer.Write(b)
	if err != nil {
		return 0, fmt.Errorf("error in gzipWriter method Write: %w", err)
	}
	return writer, nil
}

func GzipCompressHandle(next http.Handler) http.Handler {
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
		log := logger.CreateLogger()

		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			log.Error(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer func() {
			if err := gz.Close(); err != nil {
				log.Error(err)
			}
		}()

		w.Header().Add("Content-Encoding", "gzip")
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
}
