package store

import (
	"encoding/base64"
	"errors"
)

var (
	ErrKeyNotFound  = errors.New("key not found")
	ErrInvalidKey   = errors.New("invalid key ' ' is not allowed ")
	ErrCapacityFull = errors.New("reached max capacity")
)

type Storer interface {
	Get(key string) (string, error)
	Set(key string, value string) error
	Keys() []string
	Delete(key string)
	Len() int
	Rename(oldkey, newKey string)
	Pop(key string) (string, bool)
	Incr(key string) (int, error)
}

func SetKeyWithEncryption(store Storer, key, val string) (string, error) {
	encoded := base64.StdEncoding.EncodeToString([]byte(val))
	if err := store.Set(key, encoded); err != nil {
		return "", err
	}
	store.Len()
	store.Get("key")
	return store.Get("a")
}
