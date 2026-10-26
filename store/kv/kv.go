package kv

import (
	"fmt"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/Ayushmangit/redisGO/store"
)

type Store struct {
	mu      sync.RWMutex
	data    map[string]string
	MaxSize int
}

func NewStore(MaxSize int) *Store {
	return &Store{
		data:    make(map[string]string),
		MaxSize: MaxSize,
	}
}

func (s *Store) Clone() *Store {
	cloneMap := make(map[string]string)
	for k, v := range s.data {
		cloneMap[k] = v
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	cloneStore := &Store{
		data:    cloneMap,
		MaxSize: s.MaxSize,
	}
	return cloneStore
}

func (s *Store) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]string, 0, len(s.data))
	for key := range s.data {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}

func (s *Store) Get(key string) (string, error) {
	if key == "" {
		return "", store.ErrInvalidKey
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.data[key]
	if !ok {
		return "", store.ErrKeyNotFound
	}
	return val, nil
}

func (s *Store) Set(key, val string) error {
	if key == "" {
		return store.ErrInvalidKey
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.data[key]

	if s.MaxSize > 0 && len(s.data) >= s.MaxSize && !exists {
		return fmt.Errorf("Set(%q): %w", key, store.ErrCapacityFull)
	}

	s.data[key] = val
	return nil
}

func (s *Store) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
}

func (s *Store) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.data)
}

func (s *Store) Rename(oldKey, newKey string) {
	if oldKey == "" || newKey == "" {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	val, exists := s.data[oldKey]
	if !exists {
		return
	}

	delete(s.data, oldKey)
	s.data[newKey] = val
}

func (s *Store) Pop(key string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	val, ok := s.data[key]
	if ok {
		delete(s.data, key)
	}

	return val, ok
}

func (s *Store) Incr(key string) (int, error) {
	if key == "" {
		return 0, store.ErrInvalidKey
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	current := 0
	if raw, ok := s.data[key]; ok {
		v, err := strconv.Atoi(raw)
		if err != nil {
			return 0, fmt.Errorf("INCR %q: %w", key, err)
		}
		current = v
	}
	current += 1

	time.Sleep(time.Microsecond)

	currentString := strconv.Itoa(current)
	s.data[key] = currentString

	return current, nil
}
