package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/Ayushmangit/redisGO/store"
)

type MetricsMiddleware struct {
	inner       store.Storer
	getCalls    int
	setCalls    int
	DeleteCalls int
	lenCalls    int
	keysCalls   int
	popCalls    int
	renameCalls int

	getMisses       int
	totalGetLatency time.Duration
	totalSetLatency time.Duration
}

func NewMetricsMiddleware(inner store.Storer) *MetricsMiddleware {
	return &MetricsMiddleware{
		inner: inner,
	}
}

func (m *MetricsMiddleware) Get(key string) (string, error) {
	start := time.Now()
	val, err := m.inner.Get(key)
	if err != nil {
		if errors.Is(err, store.ErrKeyNotFound) {
			m.getMisses++
		}
	}
	m.getCalls += 1
	m.totalGetLatency += time.Since(start)
	return val, err
}

func (m *MetricsMiddleware) Set(key, val string, ttl time.Duration) error {
	start := time.Now()
	m.setCalls += 1
	m.totalSetLatency += time.Since(start)
	return m.inner.Set(key, val, ttl)
}

func (m *MetricsMiddleware) Delete(key string) {
	m.DeleteCalls += 1
	m.inner.Delete(key)
}

func (m *MetricsMiddleware) Len() int {
	m.lenCalls += 1
	return m.inner.Len()
}

func (m *MetricsMiddleware) Keys() []string {
	m.keysCalls += 1
	return m.inner.Keys()
}

func (m *MetricsMiddleware) Rename(oldkey, newKey string) {
	m.renameCalls += 1
	m.inner.Rename(oldkey, newKey)
}

func (m *MetricsMiddleware) Pop(key string) (string, bool) {
	m.popCalls += 1
	return m.inner.Pop(key)
}

func (m *MetricsMiddleware) Report() {
	fmt.Printf("GET calls: %d. (missed %d)\n", m.getCalls, m.getMisses)
	fmt.Printf("SET calls: %d\n", m.setCalls)
	fmt.Printf("DELETE calls: %d\n", m.DeleteCalls)
	fmt.Printf("LEN calls: %d\n", m.lenCalls)
	fmt.Printf("KEYS calls: %d\n", m.keysCalls)
	fmt.Printf("RENAME calls: %d\n", m.renameCalls)
	fmt.Printf("POP calls: %d\n", m.popCalls)

	if m.getCalls > 0 {
		avg := m.totalGetLatency / time.Duration(m.getCalls)
		fmt.Printf("Avg GET latency : %v\n", avg)
	}
	if m.setCalls > 0 {
		avg := m.totalSetLatency / time.Duration(m.setCalls)
		fmt.Printf("Avg SET latency : %v\n", avg)
	}
}
