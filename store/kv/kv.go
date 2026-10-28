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

	cleanupStop     chan struct{}
	cleanupDone     chan struct{}
	cleanupDoneOnce sync.Once
	cleanupMu       sync.Mutex
	cleanupStarted  bool
	cleanupStopped  bool
}
type TTLentry struct {
	val    string
	expiry time.Time
}

func NewStore(MaxSize int, ttl time.Duration) *Store {
	return &Store{
		data:        make(map[string]TTLentry),
		MaxSize:     MaxSize,
		ttl:         ttl,
		cleanupStop: make(chan struct{}),
		cleanupDone: make(chan struct{}),
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
		data:        cloneMap,
		MaxSize:     s.MaxSize,
		ttl:         s.ttl,
		cleanupStop: make(chan struct{}),
		cleanupDone: make(chan struct{}),
	}
	return cloneStore
}

func (s *Store) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]string, 0, len(s.data))
	now := time.Now()
	for key, entry := range s.data {
		if !entry.expiry.IsZero() && !now.Before(entry.expiry) {
			continue
		}
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
	if ttl < 0 {
		return fmt.Errorf("TTL: invalid ttl %w", store.ErrInvalidTTL)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for k, entry := range s.data {
		if !entry.expiry.IsZero() && !now.Before(entry.expiry) {
			delete(s.data, k)
		}
	}

	_, exists := s.data[key]

	if s.MaxSize > 0 && len(s.data) >= s.MaxSize && !exists {
		return fmt.Errorf("Set(%q): %w", key, store.ErrCapacityFull)
	}
	var expiry time.Time
	if ttl > 0 {
		expiry = now.Add(ttl)
	} else if ttl == 0 {
		expiry = time.Time{}
	}

	s.data[key] = TTLentry{
		val:    val,
		expiry: expiry,
	}
	return nil
}

func (s *Store) Delete(key string) error {
	if key == "" {
		return store.ErrInvalidKey
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
	return nil
}

func (s *Store) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	len := 0
	now := time.Now()
	for _, entry := range s.data {
		if !entry.expiry.IsZero() && !now.Before(entry.expiry) {
			continue
		}
		len++
	}

	return len
}

func (s *Store) Rename(oldKey, newKey string) error {
	if oldKey == "" || newKey == "" {
		return store.ErrInvalidKey
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.data[oldKey]
	if !exists {
		return store.ErrKeyNotFound
	}
	if !entry.expiry.IsZero() && !time.Now().Before(entry.expiry) {
		delete(s.data, oldKey)
		return store.ErrKeyNotFound
	}

	delete(s.data, oldKey)
	s.data[newKey] = entry
	return nil
}

func (s *Store) Pop(key string) (string, error) {
	if key == "" {
		return "", store.ErrInvalidKey
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.data[key]
	if !ok {
		return "", store.ErrKeyNotFound
	}
	now := time.Now()
	if !entry.expiry.IsZero() && !now.Before(entry.expiry) {
		delete(s.data, key)
		return "", store.ErrKeyNotFound
	}
	if ok {
		delete(s.data, key)
	}

	return entry.val, nil
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
