package main

import (
	"testing"

	"github.com/Ayushmangit/redisGO/store/kv"
)

func TestHandleInput(t *testing.T) {
	s := kv.NewStore(10, 0)
	executor := NewExecutor(s, 10)
	defer executor.Stop()

	tests := []struct {
		input string
		want  string
	}{
		{"SET name redis", "OK"},
		{"GET name", "redis"},
		{"INCR counter", "1"},
		{"GET counter", "1"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := HandleInput(executor, tt.input)
			if err != nil {
				t.Fatalf("HandleInput() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("HandleInput() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHandleInput_InvalidCommand(t *testing.T) {
	s := kv.NewStore(10, 0)
	executor := NewExecutor(s, 10)
	defer executor.Stop()

	_, err := HandleInput(executor, "FLUSHALL")
	if err == nil {
		t.Fatal("expected an error for an invalid command")
	}
}
