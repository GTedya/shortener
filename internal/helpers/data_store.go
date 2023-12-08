package helpers

import (
	"encoding/json"
	"os"

	"github.com/GTedya/shortener/config"
	"go.uber.org/zap"
)

const writingPermission = 0600

type memoryStore struct {
	data *URLData
	conf config.Config
}
type fileStore struct {
	data *URLData
	conf config.Config
	log  *zap.SugaredLogger
	memoryStore
}

type Store interface {
	Store(id, shortID string)
}

func NewStore(conf config.Config, log *zap.SugaredLogger, data *URLData) Store {
	if len(conf.FileStoragePath) == 0 {
		return memoryStore{conf: conf, data: data}
	}
	return fileStore{conf: conf, log: log, data: data}
}

func (m memoryStore) Store(id, shortID string) {
	m.data.URLMap[ShortURL{shortID}] = URL{id}
}

func (f fileStore) Store(id, shortID string) {
	f.memoryStore.data = f.data
	f.memoryStore.Store(id, shortID)
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
