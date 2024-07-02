package memory

import (
	"sync"

	"github.com/leodayo/url-shortener/internal/app/entity"
)

type ShortenUrlMemoryStorage struct {
	syncMap sync.Map
}

func (storage *ShortenUrlMemoryStorage) Store(entity entity.ShortenUrl) bool {
	_, loaded := storage.syncMap.LoadOrStore(entity.Url, entity)
	return !loaded
}

func (storage *ShortenUrlMemoryStorage) Retrieve(key string) (e entity.ShortenUrl, ok bool) {
	v, ok := storage.syncMap.Load(key)
	if !ok {
		return e, ok
	}
	return v.(entity.ShortenUrl), ok
}
