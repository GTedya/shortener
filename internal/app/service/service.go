package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/GTedya/shortener/config"
	"github.com/GTedya/shortener/internal/app/models"
	"github.com/GTedya/shortener/internal/app/repository"
)

var timeout = 5 * time.Second

type ShortenerInterface interface {
	Shorten(ctx context.Context, url string, userID string) (models.ShortURL, error)
	Expand(ctx context.Context, id string) (models.ShortURL, error)
	FormatShortURL(urlID string) string
	GetUrlsCreatedBy(ctx context.Context, userID string) ([]models.ShortURL, error)
	HealthCheck(ctx context.Context) error
	ShortenBatch(ctx context.Context, batch []models.ShortURL, userID string) ([]models.ShortURL, error)
	GenerateNewUserID() string
	DeleteUrls(ctx context.Context, ids []string, userID string)
	GetStats(ctx context.Context) (models.Stats, error)
}

// Shortener is main service of application.
type Shortener struct {
	repository repository.Repository
	config     *config.Config
}

// NewShortener creates new service.
func NewShortener(
	repository repository.Repository,
	config *config.Config,
) *Shortener {
	return &Shortener{
		repository: repository,
		config:     config}
}

// shorteningError is error wrapper of any error occurred in service.
type shorteningError struct {
	Err      error
	ShortURL models.ShortURL
}

func (err *shorteningError) Error() string {
	return fmt.Sprintf("error while shortening: %v", err.Err)
}

func (err *shorteningError) Unwrap() error {
	return err.Err
}

// NewShorteningError wraps error err with additional info about url.
func NewShorteningError(shortURL models.ShortURL, err error) error {
	return &shorteningError{
		Err:      err,
		ShortURL: shortURL,
	}
}

// Shorten shortens full url and returns filled struct ShortURL.
func (service *Shortener) Shorten(ctx context.Context, url string, userID string) (models.ShortURL, error) {
	urlID := uuid.NewString()

	shortURL := models.ShortURL{
		OriginalURL: url,
		ShortURL:    urlID,
		CreatedByID: userID,
	}

	err := service.repository.Save(ctx, shortURL)
	if errors.Is(err, repository.ErrDuplicate) {
		return shortURL, NewShorteningError(shortURL, err)
	}
	if err != nil {
		return models.ShortURL{}, fmt.Errorf("error while saving shortening URL: %w", err)
	}

	return shortURL, nil
}

// Expand expands full url from given id. Returns filled ShortURL struct.
func (service *Shortener) Expand(ctx context.Context, id string) (models.ShortURL, error) {
	origURL, err := service.repository.GetByID(ctx, id)
	if err != nil {
		return models.ShortURL{}, fmt.Errorf("error while getting shortening URL: %w", err)
	}
	return origURL, nil
}

// FormatShortURL formats url id to full url.
func (service *Shortener) FormatShortURL(urlID string) string {
	return fmt.Sprintf("%s/%s", service.config.URL, urlID)
}

// GetUrlsCreatedBy returns array of all urs that was shortened by given userID.
// It's just a wrapper for repository.GetUsersUrls.
func (service *Shortener) GetUrlsCreatedBy(ctx context.Context, userID string) ([]models.ShortURL, error) {
	urls, err := service.repository.GetUsersUrls(ctx, userID)
	if err != nil {
		return []models.ShortURL{}, fmt.Errorf("error while getting user's shortening URLs: %w", err)
	}
	return urls, nil
}

// HealthCheck checks if service is working correctly.
func (service *Shortener) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err := service.repository.Check(ctx)
	if err != nil {
		return fmt.Errorf("error while healt check: %w", err)
	}
	return nil
}

// ShortenBatch shortens array of urls.
// All entries of batch must contain OriginalURL.
func (service *Shortener) ShortenBatch(ctx context.Context,
	batch []models.ShortURL, userID string) ([]models.ShortURL, error) {
	for i := range batch {
		urlID := uuid.NewString()
		batch[i].ShortURL = urlID
		batch[i].CreatedByID = userID
	}

	if err := service.repository.SaveBatch(ctx, batch); err != nil {
		return nil, fmt.Errorf("error while batch: %w", err)
	}

	return batch, nil
}

// GenerateNewUserID generates new user id.
// It's just a wrapper for random.GenerateNewUserID().
func (service *Shortener) GenerateNewUserID() string {
	return uuid.NewString()
}

// DeleteUrls deletes all urls with given ids that was created by userID.
// Implements fan-in and fan-out pattern for learning purposes.
func (service *Shortener) DeleteUrls(ctx context.Context, ids []string, userID string) {
	done := make(chan struct{})
	defer close(done)

	workersCount := runtime.NumCPU()
	inputCh := make(chan string)
	modelsToDelete := make([]models.ShortURL, 0, len(ids))

	go func() {
		for _, id := range ids {
			inputCh <- id
		}

		close(inputCh)
	}()

	workerChs := make([]chan models.ShortURL, 0, workersCount)
	for urlID := range inputCh {
		workerCh := make(chan models.ShortURL)
		newWorker(urlID, userID, workerCh)
		workerChs = append(workerChs, workerCh)
	}

	for v := range fanIn(done, workerChs...) {
		modelsToDelete = append(modelsToDelete, v)
	}

	err := service.repository.DeleteUrls(ctx, modelsToDelete)
	if err != nil {
		fmt.Printf("couldn't delete urls: %v\n", err)
	}
}

func (service *Shortener) GetStats(ctx context.Context) (models.Stats, error) {
	usersCount, urlsCount, err := service.repository.GetUsersAndUrlsCount(ctx)
	if err != nil {
		return models.Stats{}, fmt.Errorf("getting info error: %w", err)
	}

	return models.Stats{UsersCount: usersCount, UrlsCount: urlsCount}, nil
}

func newWorker(urlID string, userID string, out chan models.ShortURL) {
	go func() {
		defer func() {
			if x := recover(); x != nil {
				newWorker(urlID, userID, out)
				log.Printf("run time panic: %v, %v", x, out)
			}
		}()

		out <- models.ShortURL{ShortURL: urlID, CreatedByID: userID}
		close(out)
	}()
}

func fanIn(done <-chan struct{}, channels ...chan models.ShortURL) chan models.ShortURL {
	var wg sync.WaitGroup
	multiplexedStream := make(chan models.ShortURL)

	multiplex := func(c <-chan models.ShortURL) {
		defer wg.Done()
		for v := range c {
			select {
			case <-done:
				return
			case multiplexedStream <- v:
			}
		}
	}

	wg.Add(len(channels))
	for _, c := range channels {
		go multiplex(c)
	}

	go func() {
		wg.Wait()
		close(multiplexedStream)
	}()

	return multiplexedStream
}
