package middlewares

import (
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Middleware представляет middleware для логирования HTTP-запросов.
type Middleware struct {
	Log       *zap.SugaredLogger
	SecretKey string
}

// loggerWriter представляет структуру для перехвата записи в ответ.
type loggerWriter struct {
	http.ResponseWriter
}

// responseData представляет данные ответа для логирования.
type responseData struct {
	status int
	size   int
}

// loggingResponseWriter представляет структуру для перехвата записи и записи кода ответа.
type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

// Write перехватывает запись в ответ и обновляет размер ответа.
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	if err != nil {
		return 0, fmt.Errorf("error in loggingResponseWriter method Write: %w", err)
	}
	r.responseData.size += size
	return size, nil
}

// WriteHeader перехватывает запись кода ответа.
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

// LogHandle возвращает HTTP-обработчик, который выполняет логирование запроса и ответа.
func (m Middleware) LogHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		resData := &responseData{
			status: 0,
			size:   0,
		}

		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   resData,
		}
		next.ServeHTTP(loggerWriter{ResponseWriter: &lw}, r)

		m.Log.Infoln(
			"uri", r.RequestURI,
			"method", r.Method,
			"status", resData.status,
			"duration", time.Since(start),
			"size", resData.size,
		)
	})
}
