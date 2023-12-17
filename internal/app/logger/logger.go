package logger

import "go.uber.org/zap"

func CreateLogger() *zap.SugaredLogger {
	logger := zap.Must(zap.NewDevelopment())
	return logger.Sugar()
}
