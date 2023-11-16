package logger

import (
	"net/http"
	"time"
)

type Logger struct {
	handler http.Handler
}

func (l *Logger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	responseData := &responseData{
		status: 0,
		size:   0,
	}

	lw := loggingResponseWriter{
		ResponseWriter: w,
		responseData:   responseData,
	}

	l.handler.ServeHTTP(&lw, r)

	logger := CreateLogger()
	defer logger.Sync()

	logger.Infoln(
		"uri", r.RequestURI,
		"method", r.Method,
		"status", responseData.status,
		"duration", time.Since(start),
		"size", responseData.size,
	)
}

func CreateLogHandler(handlerToWrap http.Handler) *Logger {
	return &Logger{handler: handlerToWrap}
}
