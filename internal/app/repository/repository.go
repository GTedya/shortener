package repository

import (
	"context"
	"errors"

	"github.com/GTedya/shortener/config"
	"github.com/GTedya/shortener/internal/app/models"
)

const DeleteBuffer = 10

// ErrDuplicate возвращается при попытке сохранить URL, который уже существует в базе данных.
var ErrDuplicate = errors.New("this url already exists")

// Repository saves and retrieves data from storage.
type Repository interface {
	Save(ctx context.Context, shortURL models.ShortURL) error
	GetByID(ctx context.Context, id string) (models.ShortURL, error)
	ShortenByURL(ctx context.Context, url string) (models.ShortURL, error)
	GetUsersUrls(ctx context.Context, userID string) ([]models.ShortURL, error)
	Close(_ context.Context) error
	Check(ctx context.Context) error
	SaveBatch(ctx context.Context, batch []models.ShortURL) error
	DeleteUrls(ctx context.Context, urls []models.ShortURL) error
	GetUsersAndUrlsCount(ctx context.Context) (int, int, error)
}

func GetRepo(cfg config.Config) Repository {
	if cfg.DatabaseDSN != "" {
		repo, err := NewPgRepository(cfg.DatabaseDSN, cfg.MigrationPath)
		if err != nil {
			panic(err)
		}
		return repo
	}
	if cfg.FileStoragePath != "" {
		repo, err := NewFileRepository(cfg.FileStoragePath)
		if err != nil {
			panic(err)
		}
		return repo
	}

	return NewInMemoryRepository()
}
