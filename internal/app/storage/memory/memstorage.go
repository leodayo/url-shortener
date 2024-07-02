package memory

import (
	"sync"

	"github.com/leodayo/url-shortener/internal/app/entity"
)

type ShortenURLMemoryStorage struct {
	syncMap sync.Map
}

func (storage *ShortenURLMemoryStorage) Store(entity entity.ShortenURL) bool {
	_, loaded := storage.syncMap.LoadOrStore(entity.URL, entity)
	return !loaded
}

func (storage *ShortenURLMemoryStorage) Retrieve(key string) (e entity.ShortenURL, ok bool) {
	v, ok := storage.syncMap.Load(key)
	if !ok {
		return e, ok
	}
	return v.(entity.ShortenURL), ok
}
