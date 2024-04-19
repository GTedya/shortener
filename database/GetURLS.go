package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/GTedya/shortener/internal/helpers"
)

type GetURLS interface {
	GetShortURL(ctx context.Context, url string) (string, error)
	GetBasicURL(ctx context.Context, shortID string) (string, bool, error)
	UserURLS(ctx context.Context, token string) ([]helpers.UserURL, error)
	IsUserURL(ctx context.Context, token string, shortURL string) (bool, error)
}

func (db *db) GetBasicURL(ctx context.Context, shortID string) (string, bool, error) {
	var (
		url       string
		isDeleted bool
		err       error
	)

	err = db.pool.QueryRow(ctx, "SELECT url, is_deleted FROM urls WHERE short_url = $1", shortID).Scan(&url, &isDeleted)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, fmt.Errorf("record not found")
		}
		return "", false, fmt.Errorf(ErrQuery, err)
	}

	return url, isDeleted, nil
}

func (db *db) GetShortURL(ctx context.Context, id string) (string, error) {
	var url string
	err := db.pool.QueryRow(ctx, "SELECT short_url FROM urls WHERE url = $1", id).Scan(&url)
	if err != nil {
		return "", fmt.Errorf(ErrQuery, err)
	}
	return url, nil
}

func (db *db) UserURLS(ctx context.Context, token string) ([]helpers.UserURL, error) {
	rows, err := db.pool.Query(ctx, "SELECT short_url, url FROM urls WHERE user_token = $1 AND is_deleted=false", token)
	if err != nil {
		return nil, fmt.Errorf(ErrQuery, err)
	}
	var urls []helpers.UserURL

	for rows.Next() {
		var url helpers.UserURL
		if err = rows.Scan(&url.ShortURL, &url.OriginalURL); err != nil {
			return nil, fmt.Errorf("rows scan error: %w", err)
		}
		url.ShortURL = BaseURL + url.ShortURL
		urls = append(urls, url)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error in query all urls: %w", err)
	}

	defer rows.Close()

	return urls, nil
}

func (db *db) IsUserURL(ctx context.Context, token string, shortURL string) (bool, error) {
	var isOwner bool
	err := db.pool.QueryRow(ctx, "select exists (select true from urls WHERE short_url = $1 AND user_token = $2);",
		shortURL, token).Scan(&isOwner)
	if err != nil {
		return false, fmt.Errorf(ErrQuery, err)
	}
	return isOwner, nil
}
