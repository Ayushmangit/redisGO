package main

import (
	"log"
	"os"
	"time"

	"github.com/Ayushmangit/redisGO/store"
)

type LoggingMiddleware struct {
	inner  store.Storer
	logger *log.Logger
}

func NewLoggingMiddleware(inner store.Storer) *LoggingMiddleware {
	return &LoggingMiddleware{
		inner:  inner,
		logger: log.New(os.Stdout, "[log] ", log.Ltime|log.Lmicroseconds),
	}
}

func (l *LoggingMiddleware) Get(key string) (string, error) {
	l.logger.Printf("GET %q", key)
	val, err := l.inner.Get(key)
	if err != nil {
		l.logger.Printf("GET %q -> MISS (%v)", key, err)
	} else {
		l.logger.Printf("GET %q -> HIT", key)
	}
	return val, err
}

func (l *LoggingMiddleware) Set(key, val string, ttl time.Duration) error {
	l.logger.Printf("SET %q", key) // side effect
	return l.inner.Set(key, val, ttl)
}

func (l *LoggingMiddleware) Delete(key string) {
	l.logger.Printf("DELETE %q", key) // side effect
	l.inner.Delete(key)
}

func (l *LoggingMiddleware) Len() int {
	length := l.inner.Len()
	l.logger.Printf("LEN %d", length)
	return length
}

func (l *LoggingMiddleware) Keys() []string {
	got := l.inner.Keys()
	l.logger.Printf("KEYS %v", got)
	return got
}

func (l *LoggingMiddleware) Rename(oldkey, newKey string) {
	l.logger.Printf("RENAME %q -> %q", oldkey, newKey)
	l.inner.Rename(oldkey, newKey)
}

func (l *LoggingMiddleware) Pop(key string) (string, error) {
	poped, err := l.inner.Pop(key)
	l.logger.Printf("POP %q -> VAL: %q ", key, poped)
	return poped, err
}
