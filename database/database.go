package database

import (
	"database/sql"
	"fmt"

	"go.uber.org/zap"

	"github.com/GTedya/shortener/config"
)

func CreateDB(config config.Config, log *zap.SugaredLogger) (*sql.DB, error) {
	db, err := sql.Open("pgx", config.DatabaseDSN)
	if err != nil {
		log.Errorw("Unable to connect to database", err)
		return nil, fmt.Errorf("database connection error: %w", err)
	}
	return db, nil
}
