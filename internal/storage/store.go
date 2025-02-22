package store

import (
	"shorturl/internal/storage/file"
	"shorturl/internal/storage/memory"
)

type Store interface {
	Get(key string) (string, error)
	Set(key, value string) error
	Load() error
}

func NewStore(consumer *file.Consumer, producer *file.Producer) (Store, error) {
	return memory.NewStore(consumer, producer), nil
}
