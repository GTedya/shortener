package config

import (
	"database/sql"
	"fmt"

	"go.uber.org/zap"
)

func CreateDBConn(config Config, log *zap.SugaredLogger) (*sql.DB, error) {
	db, err := sql.Open("pgx", config.DatabaseDSN)
	if err != nil {
		log.Errorw("Unable to connect to database", err)
		return nil, fmt.Errorf("database connection error: %w", err)
	}
	return db, nil
}
