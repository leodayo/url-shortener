package storage

import (
	"github.com/leodayo/url-shortener/internal/app/entity"
	"github.com/leodayo/url-shortener/internal/app/storage/file"
	"github.com/leodayo/url-shortener/internal/app/storage/memory"
)

var Repository Storage[string, entity.ShortenURL]

type Storage[K comparable, E any] interface {
	Store(entity E) bool
	Retrieve(key K) (E, bool)
}

func ItinInMemoryStorage() {
	Repository = memory.CreateStorage()
}

func InitFileStorage() (err error) {
	Repository, err = file.CreateStorage()
	return err
}
