package main

import (
	"errors"
	"sync"
	"testing"

	"github.com/Ayushmangit/redisGO/store/kv"
)

func TestExecutor_SetThenGet(t *testing.T) {
	s := kv.NewStore(10, 0)
	executor := NewExecutor(s, 10)
	defer executor.Stop()
	// SET name redis
	got, err := executor.Execute(Command{
		OP:    "SET",
		Key:   "name",
		Value: "redis",
	})
	if err != nil {
		t.Fatalf("SET failed: %v", err)
	}
	if got != "OK" {
		t.Fatalf("SET returned %q, want %q", got, "OK")
	}

	// GET name
	got, err = executor.Execute(Command{
		OP:  "GET",
		Key: "name",
	})
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	if got != "redis" {
		t.Fatalf("GET returned %q, want %q", got, "redis")
	}
}

func TestExecutor_StopIsIdempotent(t *testing.T) {
	s := kv.NewStore(10, 0)
	executor := NewExecutor(s, 10)

	executor.Stop()
	executor.Stop()
}

func TestExecutor_ExecuteAfterStop(t *testing.T) {
	s := kv.NewStore(10, 0)
	executor := NewExecutor(s, 10)
	executor.Stop()

	_, err := executor.Execute(Command{
		OP:  "GET",
		Key: "name",
	})
	if !errors.Is(err, ErrExecutorStopped) {
		t.Fatalf("Execute error = %v, want %v", err, ErrExecutorStopped)
	}
}

func TestExecutor_ConcurrentExecution(t *testing.T) {
	s := kv.NewStore(100, 0)
	executor := NewExecutor(s, 10)
	defer executor.Stop()

	const workers = 50

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			_, err := executor.Execute(Command{
				OP:  "INCR",
				Key: "counter",
			})
			if err != nil {
				t.Errorf("INCR failed: %v", err)
			}
		}()
	}

	wg.Wait()

	got, err := executor.Execute(Command{
		OP:  "GET",
		Key: "counter",
	})
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}

	if got != "50" {
		t.Fatalf("counter = %q, want %q", got, "50")
	}
}
