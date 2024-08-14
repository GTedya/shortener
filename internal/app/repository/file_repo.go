package repository

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/GTedya/shortener/internal/app/models"
)

var ErrDecoding = errors.New("decoding error")
var ErrFileSeek = errors.New("file seek error")

// FileRepository is repository that uses files for storage.
type FileRepository struct {
	file   *os.File      // file that we will be writing to
	writer *bufio.Writer // buffered writer that will write to the file
	mutex  sync.RWMutex  // mutex that will be used to synchronize access to the file
}

// NewFileRepository creates new file repository. Creates file at filePath if it doesn't exist.
// It opens a file, creates a buffered writer, and returns a pointer to a FileRepository.
func NewFileRepository(filePath string) (*FileRepository, error) {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600) //nolint:gomnd // permission
	if err != nil {
		return nil, fmt.Errorf("file opening error: %w", err)
	}

	return &FileRepository{
		mutex:  sync.RWMutex{},
		file:   file,
		writer: bufio.NewWriter(file),
	}, nil
}

// SaveBatch saves multiple urls.
// Checks if the urls are unique and then saving them.
func (repo *FileRepository) SaveBatch(ctx context.Context, batch []models.ShortURL) error {
	for _, shortURL := range batch {
		_, err := repo.GetByID(ctx, shortURL.ShortURL)
		if err == nil {
			return ErrDuplicate
		}
	}

	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	for _, shortURL := range batch {
		data, err := json.Marshal(shortURL)
		if err != nil {
			return fmt.Errorf("unmarshalling error: %w", err)
		}

		if _, errWrite := repo.writer.Write(data); errWrite != nil {
			return fmt.Errorf("writer error: %w", errWrite)
		}

		if errWriteByte := repo.writer.WriteByte('\n'); errWriteByte != nil {
			return fmt.Errorf("byte writing error: %w", err)
		}
	}
	if err := repo.writer.Flush(); err != nil {
		return fmt.Errorf("flush error: %w", err)
	}

	return nil
}

// Save checks if the url is unique and then saving it to the file.
func (repo *FileRepository) Save(ctx context.Context, shortURL models.ShortURL) error {
	_, err := repo.GetByID(ctx, shortURL.ShortURL)
	if err == nil {
		return ErrDuplicate
	}

	data, err := json.Marshal(shortURL)
	if err != nil {
		return fmt.Errorf("marshalling error: %w", err)
	}

	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	if _, errWrite := repo.writer.Write(data); errWrite != nil {
		return fmt.Errorf("writer error: %w", errWrite)
	}

	if errWriteByte := repo.writer.WriteByte('\n'); errWriteByte != nil {
		return fmt.Errorf("byte writing error: %w", err)
	}

	if errFlush := repo.writer.Flush(); errFlush != nil {
		return fmt.Errorf("flush error: %w", errFlush)
	}

	return nil
}

// readEntries scans the file and applies a callback to each decoded entry.
func (repo *FileRepository) readEntries(callback func(entry models.ShortURL) (bool, error)) error {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	if _, err := repo.file.Seek(0, io.SeekStart); err != nil {
		return ErrFileSeek
	}

	scanner := bufio.NewScanner(repo.file)
	for scanner.Scan() {
		line := scanner.Bytes()
		var entry models.ShortURL
		if err := json.NewDecoder(bytes.NewReader(line)).Decode(&entry); err != nil {
			return ErrDecoding
		}
		stop, err := callback(entry)
		if err != nil {
			return fmt.Errorf("callback error: %w", err)
		}
		if stop {
			break
		}
	}

	return nil
}

// GetByID gets a URL by its ID.
func (repo *FileRepository) GetByID(_ context.Context, id string) (models.ShortURL, error) {
	var result models.ShortURL
	err := repo.readEntries(func(entry models.ShortURL) (bool, error) {
		if entry.ShortURL == id {
			result = entry
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return models.ShortURL{}, err
	}
	if result.ShortURL == "" {
		return models.ShortURL{}, errors.New("can't find full URL by id")
	}
	return result, nil
}

// ShortenByURL gets a shortened URL by its original URL.
func (repo *FileRepository) ShortenByURL(_ context.Context, originalURL string) (models.ShortURL, error) {
	var result models.ShortURL
	err := repo.readEntries(func(entry models.ShortURL) (bool, error) {
		if entry.OriginalURL == originalURL {
			result = entry
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return models.ShortURL{}, err
	}
	if result.ShortURL == "" {
		return models.ShortURL{}, errors.New("can't find shortened URL by original URL")
	}
	return result, nil
}

// GetUsersUrls reads the file line by line and returning all the urls that were created by user with id userID.
func (repo *FileRepository) GetUsersUrls(_ context.Context, userID string) ([]models.ShortURL, error) {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	if _, err := repo.file.Seek(0, io.SeekStart); err != nil {
		return nil, ErrFileSeek
	}

	var entry models.ShortURL
	var URLs []models.ShortURL

	scanner := bufio.NewScanner(repo.file)

	for scanner.Scan() {
		line := scanner.Bytes()
		if err := json.NewDecoder(bytes.NewReader(line)).Decode(&entry); err != nil {
			return nil, ErrDecoding
		}
		if entry.CreatedByID == userID {
			URLs = append(URLs, entry)
		}
	}

	return URLs, nil
}

// Close closes file.
func (repo *FileRepository) Close(_ context.Context) error {
	if err := repo.file.Close(); err != nil {
		return fmt.Errorf("file close error: %w", err)
	}
	return nil
}

// Check checks if file is ok.
func (repo *FileRepository) Check(_ context.Context) error {
	if _, err := repo.file.Stat(); err != nil {
		return fmt.Errorf("file stat error: %w", err)
	}
	return nil
}

// DeleteUrls deletes all given urls.
func (repo *FileRepository) DeleteUrls(_ context.Context, urls []models.ShortURL) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	existingURLs, err := repo.readFileToMap()
	if err != nil {
		return fmt.Errorf("readFileToMap error: %w", err)
	}

	// mark deleted urls in memory
	for _, urlToDelete := range urls {
		foundURL, ok := existingURLs[urlToDelete.ShortURL]
		if ok && foundURL.CreatedByID == urlToDelete.CreatedByID {
			foundURL.IsDeleted = true
			existingURLs[urlToDelete.ShortURL] = foundURL
		}
	}

	// write back in memory map to file
	err = repo.writeMapToFile(existingURLs)
	if err != nil {
		return fmt.Errorf("writeMapToFile error: %w", err)
	}

	return nil
}

// readFileToMap reads the file and returns a map of all the urls in the file.
func (repo *FileRepository) readFileToMap() (map[string]models.ShortURL, error) {
	if _, err := repo.file.Seek(0, io.SeekStart); err != nil {
		return nil, ErrFileSeek
	}
	var entry models.ShortURL
	existingURLs := make(map[string]models.ShortURL)

	scanner := bufio.NewScanner(repo.file)

	for scanner.Scan() {
		line := scanner.Bytes()
		if err := json.NewDecoder(bytes.NewReader(line)).Decode(&entry); err != nil {
			return nil, ErrDecoding
		}
		existingURLs[entry.ShortURL] = entry
	}
	return existingURLs, nil
}

// writeMapToFile writes the map to the file.
func (repo *FileRepository) writeMapToFile(existingURLs map[string]models.ShortURL) error {
	if err := repo.file.Truncate(0); err != nil {
		return fmt.Errorf("file truncate error: %w", err)
	}
	if _, err := repo.file.Seek(0, 0); err != nil {
		return ErrFileSeek
	}

	for _, url := range existingURLs {
		data, err := json.Marshal(url)
		if err != nil {
			return fmt.Errorf("marshalling error: %w", err)
		}

		if _, errWrite := repo.writer.Write(data); errWrite != nil {
			return fmt.Errorf("writer error: %w", errWrite)
		}

		if errWriteByte := repo.writer.WriteByte('\n'); errWriteByte != nil {
			return fmt.Errorf("byte writing error: %w", err)
		}
	}
	if err := repo.writer.Flush(); err != nil {
		return fmt.Errorf("flush error: %w", err)
	}
	return nil
}

func (repo *FileRepository) GetUsersAndUrlsCount(_ context.Context) (int, int, error) {
	if _, err := repo.file.Seek(0, io.SeekStart); err != nil {
		return 0, 0, ErrFileSeek
	}

	uniqueUsersIds := make(map[string]bool)
	urlsCount := 0

	var entry models.ShortURL
	scanner := bufio.NewScanner(repo.file)

	for scanner.Scan() {
		line := scanner.Bytes()
		if err := json.NewDecoder(bytes.NewReader(line)).Decode(&entry); err != nil {
			return 0, 0, ErrDecoding
		}
		urlsCount++
		uniqueUsersIds[entry.CreatedByID] = true
	}

	return len(uniqueUsersIds), urlsCount, nil
}
