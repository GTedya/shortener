package repository

import (
	"context"
	"errors"
	"sync"

	"github.com/GTedya/shortener/internal/app/models"
)

// InMemoryRepository is repository that uses memory for storage.
type InMemoryRepository struct {
	storage map[string]models.ShortURL // map that will store urls
	mutex   sync.RWMutex               // read-write mutex that will be used to synchronize access to the storage map
}

// NewInMemoryRepository creates a new InMemoryRepository and returns a pointer to it.
func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		storage: make(map[string]models.ShortURL),
		mutex:   sync.RWMutex{},
	}
}

// SaveBatch saves multiple urls.
// Checks if the urls are unique and then saving them.
func (repo *InMemoryRepository) SaveBatch(_ context.Context, batch []models.ShortURL) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	for _, shortURL := range batch {
		_, ok := repo.storage[shortURL.ShortURL]
		if ok {
			return ErrDuplicate
		}
	}

	for _, shortURL := range batch {
		repo.storage[shortURL.ShortURL] = shortURL
	}

	return nil
}

// Save checks if the url is unique and then saving it to the memory.
func (repo *InMemoryRepository) Save(_ context.Context, shortURL models.ShortURL) error {
	repo.mutex.RLock()
	_, ok := repo.storage[shortURL.ShortURL]
	repo.mutex.RUnlock()

	if ok {
		return ErrDuplicate
	}

	repo.mutex.Lock()
	repo.storage[shortURL.ShortURL] = shortURL
	repo.mutex.Unlock()

	return nil
}

// GetByID gets the url by id.
func (repo *InMemoryRepository) GetByID(_ context.Context, id string) (models.ShortURL, error) {
	repo.mutex.RLock()
	url, ok := repo.storage[id]
	repo.mutex.RUnlock()

	if !ok {
		return models.ShortURL{}, errors.New("can't find full url by id")
	}

	return url, nil
}

// ShortenByURL ищет и возвращает сокращенный URL по оригинальному URL.
func (repo *InMemoryRepository) ShortenByURL(_ context.Context, originalURL string) (models.ShortURL, error) {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	for _, entry := range repo.storage {
		if entry.OriginalURL == originalURL {
			return entry, nil
		}
	}
	return models.ShortURL{}, errors.New("can't find shortened URL by original URL")
}

// GetUsersUrls gets all the urls that were created by the user with the given id.
func (repo *InMemoryRepository) GetUsersUrls(_ context.Context, userID string) ([]models.ShortURL, error) {
	repo.mutex.RLock()
	var URLs []models.ShortURL
	for _, URL := range repo.storage {
		if URL.CreatedByID == userID {
			URLs = append(URLs, URL)
		}
	}
	repo.mutex.RUnlock()
	return URLs, nil
}

// Close clears map.
func (repo *InMemoryRepository) Close(_ context.Context) error {
	repo.storage = make(map[string]models.ShortURL)
	return nil
}

// Check is just a stub.
func (repo *InMemoryRepository) Check(_ context.Context) error {
	return nil
}

// DeleteUrls deletes all given urls.
func (repo *InMemoryRepository) DeleteUrls(_ context.Context, urls []models.ShortURL) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	for _, urlToDelete := range urls {
		foundURL, ok := repo.storage[urlToDelete.ShortURL]
		if ok && foundURL.CreatedByID == urlToDelete.CreatedByID {
			foundURL.IsDeleted = true
			repo.storage[urlToDelete.ShortURL] = foundURL
		}
	}

	return nil
}

func (repo *InMemoryRepository) GetUsersAndUrlsCount(_ context.Context) (int, int, error) {
	uniqueUsersIds := make(map[string]bool)

	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	for _, shortURL := range repo.storage {
		uniqueUsersIds[shortURL.CreatedByID] = true
	}

	return len(uniqueUsersIds), len(repo.storage), nil
}
