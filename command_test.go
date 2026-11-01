package main

import (
	"testing"

	"github.com/Ayushmangit/redisGO/store/kv"
)

func TestDispatch(t *testing.T) {
	s := kv.NewStore(10, 0)

	tests := []struct {
		name     string
		command  Command
		wantType ResponseType
		wantVal  string
	}{
		{"SET", Command{OP: "SET", Key: "name", Value: "redis"}, SimpleString, "OK"},
		{"GET", Command{OP: "GET", Key: "name"}, BulkString, "redis"},
		{"INCR", Command{OP: "INCR", Key: "count"}, Integer, ""},
		{"INCR again", Command{OP: "INCR", Key: "count"}, Integer, ""},
		{"case insensitive", Command{OP: "get", Key: "name"}, BulkString, "redis"},
		{"missing key", Command{OP: "GET", Key: "missing"}, NullBulkString, ""},
		{"unknown command", Command{OP: "WAT"}, ErrorResponse, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dispatch(s, tt.command)
			if err != nil {
				t.Fatalf("dispatch() error = %v", err)
			}

			if got.responseType != tt.wantType {
				t.Fatalf("response type = %v, want %v", got.responseType, tt.wantType)
			}

			if tt.wantVal != "" && got.val != tt.wantVal {
				t.Errorf("response value = %q, want %q", got.val, tt.wantVal)
			}
		})
	}
}
