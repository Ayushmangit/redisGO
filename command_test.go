package main

import (
	"errors"
	"testing"
	"time"

	"github.com/Ayushmangit/redisGO/store"
	"github.com/Ayushmangit/redisGO/store/kv"
)

func TestDispatch(t *testing.T) {
	s := kv.NewStore(10, 0)

	tests := []struct {
		name    string
		command Command
		want    string
		wantErr bool
	}{
		{
			name:    "SET",
			command: Command{OP: "SET", Key: "name", Value: "redis"},
			want:    "OK",
		},
		{
			name:    "GET",
			command: Command{OP: "GET", Key: "name"},
			want:    "redis",
		},
		{
			name:    "INCR",
			command: Command{OP: "INCR", Key: "count"},
			want:    "1",
		},
		{
			name:    "INCR again",
			command: Command{OP: "INCR", Key: "count"},
			want:    "2",
		},
		{
			name:    "case insensitive",
			command: Command{OP: "get", Key: "name"},
			want:    "redis",
		},
		{
			name:    "missing key",
			command: Command{OP: "GET", Key: "missing"},
			wantErr: true,
		},
		{
			name:    "unknown command",
			command: Command{OP: "WAT", Key: "name"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dispatch(s, tt.command)

			if tt.wantErr {
				if err == nil {
					t.Fatal("dispatch() error = nil, want an error")
				}
				return
			}

			if err != nil {
				t.Fatalf("dispatch() error = %v", err)
			}

			if got != tt.want {
				t.Errorf("dispatch() = %q, want %q", got, tt.want)
			}
		})
	}

	_, err := dispatch(s, Command{OP: "GET", Key: "missing"})
	if !errors.Is(err, store.ErrKeyNotFound) {
		t.Errorf("error = %v, want ErrKeyNotFound", err)
	}

	// Keep the time import useful for a later TTL command test.
	_ = time.Second
}
