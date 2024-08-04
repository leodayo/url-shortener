package file

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/leodayo/url-shortener/internal/app/config"
	"github.com/leodayo/url-shortener/internal/app/entity"
	"github.com/leodayo/url-shortener/internal/app/storage/memory"
)

type ShortenURLFileStorage struct {
	memoryStorage memory.ShortenURLMemoryStorage
	fileWriter    FileWriter
}

type FileWriter struct {
	file    *os.File
	encoder *json.Encoder
}

func (storage *ShortenURLFileStorage) Store(entity entity.ShortenURL) bool {
	ok := storage.memoryStorage.Store(entity)

	if ok {
		storage.fileWriter.encoder.Encode(&entity)
	}

	return ok
}

func (storage *ShortenURLFileStorage) Retrieve(key string) (e entity.ShortenURL, ok bool) {
	return storage.memoryStorage.Retrieve(key)
}

func CreateStorage() (*ShortenURLFileStorage, error) {
	file, err := os.OpenFile(config.FileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	fw := &FileWriter{
		file:    file,
		encoder: json.NewEncoder(file),
	}

	fileStorage := &ShortenURLFileStorage{
		memoryStorage: *memory.CreateStorage(),
		fileWriter:    *fw,
	}

	if err := loadFromDisc(fileStorage); err != nil {
		return nil, err
	}

	return fileStorage, nil
}

func loadFromDisc(fileStorage *ShortenURLFileStorage) error {
	file, err := os.OpenFile(config.FileStoragePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		entry := scanner.Bytes()

		entity := entity.ShortenURL{}
		err := json.Unmarshal(entry, &entity)
		if err != nil {
			return err
		}
		fileStorage.memoryStorage.Store(entity)
	}

	return nil
}
