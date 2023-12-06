package middlewares

import (
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type (
	Middleware struct {
		Log *zap.SugaredLogger
	}
	loggerWriter struct {
		http.ResponseWriter
	}
	responseData struct {
		status int
		size   int
	}
	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	if err != nil {
		return 0, fmt.Errorf("error in loggingResponseWriter method Write: %w", err)
	}
	r.responseData.size += size
	return size, nil
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func (m Middleware) LogHandle(next http.Handler) http.Handler {
	start := time.Now()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
