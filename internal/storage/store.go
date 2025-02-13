package store

import (
	"shorturl/internal/storage/memory"
)

type Store interface {
	Get(key string) (string, error)
	Set(key, value string) error
}

func NewStore() (Store, error) {
	return memory.NewStore(), nil
}
