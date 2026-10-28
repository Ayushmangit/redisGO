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
	data    map[string]TTLentry
	ttl     time.Duration
	MaxSize int
}
type TTLentry struct {
	val    string
	expiry time.Time
}

func NewStore(MaxSize int, ttl time.Duration) *Store {
	return &Store{
		data:    make(map[string]TTLentry),
		MaxSize: MaxSize,
		ttl:     ttl,
	}
}

func (s *Store) Clone() *Store {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cloneMap := make(map[string]TTLentry)
	for k, v := range s.data {
		cloneMap[k] = v
	}

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
	s.mu.Lock()
	defer s.mu.Unlock()
	if key == "" {
		return "", store.ErrInvalidKey
	}

	entry, ok := s.data[key]
	if !ok || time.Now().After(entry.expiry) && !entry.expiry.IsZero() {
		delete(s.data, key)
		return "", store.ErrKeyNotFound
	}

	return entry.val, nil
}

func (s *Store) Set(key, val string, ttl time.Duration) error {
	if key == "" {
		return store.ErrInvalidKey
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.data[key]

	if s.MaxSize > 0 && len(s.data) >= s.MaxSize && !exists {
		return fmt.Errorf("Set(%q): %w", key, store.ErrCapacityFull)
	}
	var expiry time.Time
	if ttl > 0 {
		expiry = time.Now().Add(ttl)
	} else if ttl == 0 {
		expiry = time.Time{}
	} else {
		return fmt.Errorf("TTL: invalid ttl %w", store.ErrInvalidTTL)
	}

	s.data[key] = TTLentry{
		val:    val,
		expiry: expiry,
	}
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

	entry, ok := s.data[key]
	if ok {
		delete(s.data, key)
	}

	return entry.val, ok
}

func (s *Store) Incr(key string) (int, error) {
	if key == "" {
		return 0, store.ErrInvalidKey
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	current := 0
	var expiry time.Time
	if s.ttl > 0 {
		expiry = time.Now().Add(s.ttl)
	}

	if entry, ok := s.data[key]; ok {
		if !entry.expiry.IsZero() && !time.Now().Before(entry.expiry) {
			delete(s.data, key)
		} else {
			v, err := strconv.Atoi(entry.val)
			if err != nil {
				return 0, fmt.Errorf("INCR %q: %w", key, err)
			}

			current = v
			expiry = entry.expiry
		}
	}

	current++

	s.data[key] = TTLentry{
		val:    strconv.Itoa(current),
		expiry: expiry,
	}

	return current, nil
}
