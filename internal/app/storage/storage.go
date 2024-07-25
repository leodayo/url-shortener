package storage

type Storage[K comparable, E any] interface {
	Store(entity E) bool
	Retrieve(key K) (E, bool)
}
