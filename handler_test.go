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
		input    string
		wantType ResponseType
		wantVal  string
	}{
		{"SET name redis", SimpleString, "OK"},
		{"GET name", BulkString, "redis"},
		{"INCR counter", Integer, ""},
		{"GET counter", BulkString, "1"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := HandleInput(executor, tt.input)
			if err != nil {
				t.Fatalf("HandleInput() error = %v", err)
			}
			if got.responseType != tt.wantType {
				t.Fatalf("response type = %v, want %v", got.responseType, tt.wantType)
			}

			if tt.wantVal != "" && got.val != tt.wantVal {
				t.Fatalf("response value = %q, want %q", got.val, tt.wantVal)
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
