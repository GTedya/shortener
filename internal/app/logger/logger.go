package logger

import "go.uber.org/zap"

// CreateLogger создает новый экземпляр логгера и возвращает его в виде упрощенного логгера zap.SugaredLogger.
func CreateLogger() *zap.SugaredLogger {
	logger := zap.Must(zap.NewDevelopment())
	return logger.Sugar()
}
