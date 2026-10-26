package ttl

import (
	"fmt"
	"time"

	"github.com/Ayushmangit/redisGO/store"
)

type TTLStore struct {
	data map[string]ttlEntry
	ttl  time.Duration
}

type ttlEntry struct {
	value     string
	expiresAt time.Time
}

func NewTTLStore(ttl time.Duration) *TTLStore {
	return &TTLStore{
		data: make(map[string]ttlEntry),
		ttl:  ttl,
	}
}

func (t *TTLStore) Set(key string, value string) error {
	if key == "" {
		return store.ErrInvalidKey
	}
	var entry ttlEntry
	entry.value = value
	entry.expiresAt = time.Now().Add(t.ttl)
	t.data[key] = entry

	return nil
}

func (t *TTLStore) Get(key string) (string, error) {
	if key == "" {
		return "", store.ErrInvalidKey
	}

	entry, ok := t.data[key]
	if !ok || time.Now().After(entry.expiresAt) {
		delete(t.data, key)
		return "", fmt.Errorf("Key %s does not exist", key)
	}

	return entry.value, nil
}

func (t *TTLStore) Delete(key string) { delete(t.data, key) }
func (t *TTLStore) Len() int          { return len(t.data) }
