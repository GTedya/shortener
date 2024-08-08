package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/GTedya/shortener/internal/app/models"
)

type PostgresRepo struct {
	conn *pgx.Conn // connection to the database
	Dsn  string    // data source name for the Postgres database
}

// NewPgRepository creates a new Postgres connection, runs the migrations, and returns a new PostgresRepo.
func NewPgRepository(dsn string, migrationsPath string) (*PostgresRepo, error) {
	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("pgx connection error: %w", err)
	}

	if err = runMigrations(dsn, migrationsPath); err != nil {
		return nil, fmt.Errorf("migration running error: %w", err)
	}
	return &PostgresRepo{
		Dsn:  dsn,
		conn: conn,
	}, nil
}

// runMigrations выполняет миграции базы данных.
func runMigrations(dsn string, migrationsPath string) error {
	m, err := migrate.New(migrationsPath, dsn)
	if err != nil {
		return fmt.Errorf("migrate instance creating error: %w", err)
	}

	err = m.Up()
	if errors.Is(err, migrate.ErrNoChange) {
		fmt.Println("Nothing to migrate")
		return nil
	}
	if err != nil {
		return fmt.Errorf("migrate applying error: %w", err)
	}

	fmt.Println("Migrated successfully")
	return nil
}

// Save inserting a new row into the urls table.
func (repo *PostgresRepo) Save(ctx context.Context, shortURL models.ShortURL) error {
	_, err := repo.conn.Exec(
		ctx,
		"insert into urls (url, short_url, user_token) values ($1, $2, $3)",
		shortURL.OriginalURL,
		shortURL.ShortURL,
		shortURL.CreatedByID,
	)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		return ErrDuplicate
	}
	if err != nil {
		return fmt.Errorf("exec error: %w", err)
	}

	return nil
}

// SaveBatch is a batch insert operation.
func (repo *PostgresRepo) SaveBatch(ctx context.Context, batch []models.ShortURL) error {
	_, err := repo.conn.CopyFrom(
		ctx,
		pgx.Identifier{"urls"},
		[]string{"url", "short_url", "user_token"},
		pgx.CopyFromSlice(len(batch), func(i int) ([]interface{}, error) {
			return []interface{}{batch[i].OriginalURL, batch[i].ShortURL, batch[i].CreatedByID}, nil
		}),
	)
	if err != nil {
		return fmt.Errorf("copy error: %w", err)
	}
	return nil
}

// GetByID gets url by id.
func (repo *PostgresRepo) GetByID(ctx context.Context, id string) (models.ShortURL, error) {
	var model models.ShortURL
	err := repo.conn.QueryRow(
		ctx,
		"select url, short_url, user_token, is_deleted from urls where short_url=$1",
		id,
	).Scan(&model.OriginalURL, &model.ShortURL, &model.CreatedByID, &model.IsDeleted)
	if err != nil {
		return models.ShortURL{}, fmt.Errorf("query error: %w", err)
	}
	return model, nil
}

// ShortenByURL gets id by url.
func (repo *PostgresRepo) ShortenByURL(ctx context.Context, url string) (models.ShortURL, error) {
	var model models.ShortURL
	err := repo.conn.QueryRow(
		ctx,
		"select url, short_url, user_token, is_deleted from urls where url=$1",
		url,
	).Scan(&model.OriginalURL, &model.ShortURL, &model.CreatedByID, &model.IsDeleted)
	if err != nil {
		return models.ShortURL{}, fmt.Errorf("query error: %w", err)
	}
	return model, nil
}

// GetUsersUrls returns all the urls created by a user.
func (repo *PostgresRepo) GetUsersUrls(ctx context.Context, userID string) ([]models.ShortURL, error) {
	var URLs []models.ShortURL

	rows, err := repo.conn.Query(
		ctx,
		"select url, short_url, is_deleted from urls where user_token=$1",
		userID)
	if err != nil {
		return nil, fmt.Errorf("getting user urls error: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		model := models.ShortURL{}
		if err = rows.Scan(&model.OriginalURL, &model.ShortURL, &model.IsDeleted); err != nil {
			return nil, fmt.Errorf("scan row error: %w", err)
		}
		URLs = append(URLs, model)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("rows error :%w", rows.Err())
	}

	return URLs, nil
}

// Close closes the connection to the database.
func (repo *PostgresRepo) Close(ctx context.Context) error {
	err := repo.conn.Close(ctx)
	if err != nil {
		return fmt.Errorf("close error: %w", err)
	}
	return nil
}

// Check checks if the database is up and running.
func (repo *PostgresRepo) Check(ctx context.Context) error {
	err := repo.conn.Ping(ctx)
	if err != nil {
		return fmt.Errorf("ping error: %w", err)
	}
	return nil
}

// DeleteUrls удаляет несколько URL из базы данных, используя указанный токен пользователя.
func (repo *PostgresRepo) DeleteUrls(ctx context.Context, urls []models.ShortURL) (err error) {
	tx, err := repo.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				err = fmt.Errorf("rollback failed: %w, original error: %w", rollbackErr, err)
			}
		} else {
			if commitErr := tx.Commit(ctx); commitErr != nil {
				err = fmt.Errorf("commit failed: %w", commitErr)
			}
		}
	}()

	b := &pgx.Batch{}

	for _, url := range urls {
		sqlStatement := "UPDATE urls SET is_deleted = true WHERE short_url=$1 AND user_token=$2"
		b.Queue(sqlStatement, url.ShortURL, url.CreatedByID)
		if b.Len() >= DeleteBuffer {
			batchResults := tx.SendBatch(ctx, b)
			if err := batchResults.Close(); err != nil {
				return fmt.Errorf("batch closing error: %w", err)
			}
			b = &pgx.Batch{}
		}
	}

	if b.Len() > 0 {
		batchResults := tx.SendBatch(ctx, b)
		if err := batchResults.Close(); err != nil {
			return fmt.Errorf("batch closing error: %w", err)
		}
	}

	return nil
}

func (repo *PostgresRepo) GetUsersAndUrlsCount(ctx context.Context) (int, int, error) {
	var urlsCount int
	var usersCount int
	err := repo.conn.QueryRow(
		ctx,
		"select count('*'), count( distinct user_token) from urls",
	).Scan(&urlsCount, &usersCount)
	if err != nil {
		return 0, 0, fmt.Errorf("query error: %w", err)
	}
	return usersCount, urlsCount, nil
}
