// Package store provides a concurrency-safe, in-memory string store.
package store

import "sync"

type Store struct {
	mu     sync.RWMutex
	values map[string]string
}

func New() *Store { return &Store{values: make(map[string]string)} }

func (s *Store) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values[key] = value
}

func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.values[key]
	return value, ok
}

// Delete atomically removes keys and counts each existing key only once.
func (s *Store) Delete(keys ...string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	deleted := 0
	for _, key := range keys {
		if _, exists := s.values[key]; exists {
			delete(s.values, key)
			deleted++
		}
	}
	return deleted
}
