package helpers

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/GTedya/shortener/config"
	"go.uber.org/zap"
)

const writingPermission = 0600

type memoryStore struct {
	conf config.Config
}
type fileStore struct {
	conf config.Config
	log  *zap.SugaredLogger
	memoryStore
}

type Store interface {
	Store(id, shortID string, data *URLData)
}

func NewStore(conf config.Config, log *zap.SugaredLogger) Store {
	if len(conf.FileStoragePath) == 0 {
		return memoryStore{conf: conf}
	}
	return fileStore{conf: conf, log: log}
}

func (m memoryStore) Store(id, shortID string, data *URLData) {
	fmt.Println("В памяти")
	data.URLMap[ShortURL{shortID}] = URL{id}
}

func (f fileStore) Store(id, shortID string, data *URLData) {
	f.memoryStore.Store(id, shortID, data)
	filePath := f.conf.FileStoragePath
	jsonFile := FileStorage{
		UUID:        GenerateUUID(filePath),
		ShortURL:    shortID,
		OriginalURL: id,
	}

	content, err := os.ReadFile(filePath)

	if err != nil && !os.IsNotExist(err) {
		f.log.Info(err)
	}

	var storage []FileStorage
	if len(content) > 0 {
		if err := json.Unmarshal(content, &storage); err != nil {
			f.log.Errorw("json unmarshal error", err)
			return
		}
	}

	storage = append(storage, jsonFile)

	encoded, err := json.MarshalIndent(storage, "", "  ")
	if err != nil {
		f.log.Errorw("json marshal error", err)
		return
	}

	err = os.WriteFile(filePath, encoded, writingPermission)
	if err != nil {
		f.log.Errorw("file writing error", err)
		return
	}
}
