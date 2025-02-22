package memory

import (
	"fmt"
	"io"
	"strconv"
	"sync"

	"shorturl/internal/storage/file"
)

const initialUUIDValue = 0

type Store struct {
	mux         *sync.Mutex
	s           map[string]string
	currentUUID uint
	consumer    *file.Consumer
	producer    *file.Producer
}

func NewStore(consumer *file.Consumer, producer *file.Producer) *Store {
	return &Store{
		mux:         &sync.Mutex{},
		s:           make(map[string]string),
		currentUUID: initialUUIDValue,
		consumer:    consumer,
		producer:    producer,
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

	// If value already exists, don't add it again.
	_, ok := s.s[key]
	if ok {
		return nil
	}

	s.s[key] = value
	s.currentUUID++

	// Add record to the file storage.
	r := file.Record{
		UUID:        strconv.Itoa(int(s.currentUUID)),
		ShortURL:    key,
		OriginalURL: value,
	}
	err := s.producer.WriteRecord(&r)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) Load() error {
	s.mux.Lock()
	defer s.mux.Unlock()

	for {
		r, err := s.consumer.ReadRecord()
		if err != nil {
			if err == io.EOF { // Stop reading if EOF is reached.
				break
			}
			return err
		}

		if r != nil {
			s.s[r.ShortURL] = r.OriginalURL
			s.currentUUID++
		} else { // Stop reading if there are no more records.
			break
		}
	}
	return nil
}
