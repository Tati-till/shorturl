package memory

import (
	"fmt"
	"sync"
)

type Store struct {
	mux *sync.Mutex
	s   map[string]string
}

func NewStore() *Store {
	return &Store{
		mux: &sync.Mutex{},
		s:   make(map[string]string),
	}
}

func (s *Store) Get(key string) (string, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	res, ok := s.s[key]
	if !ok {
		return "", fmt.Errorf("can't find related URL %s in storage", key)
	}
	return res, nil
}

func (s *Store) Set(key, value string) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.s[key] = value
	return nil
}
