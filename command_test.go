package main

import (
	"fmt"
	"testing"

	"github.com/Ayushmangit/redisGO/protocol"
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

func TestValueToCommand(t *testing.T) {
	tests := []struct {
		name    string
		input   protocol.Value
		want    Command
		wantErr error
	}{
		{
			name: "valid GET",
			input: protocol.Value{
				Type: protocol.Array,
				Array: []protocol.Value{
					{Type: protocol.BulkString, Str: "GET"},
					{Type: protocol.BulkString, Str: "username"},
				},
			},
			want: Command{OP: "GET", Key: "username"},
		},
		{
			name: "valid SET",
			input: protocol.Value{
				Type: protocol.Array,
				Array: []protocol.Value{
					{Type: protocol.BulkString, Str: "SET"},
					{Type: protocol.BulkString, Str: "username"},
					{Type: protocol.BulkString, Str: "ayushman"},
				},
			},
			want: Command{OP: "SET", Key: "username", Value: "ayushman"},
		},
		{
			name: "valid INCR",
			input: protocol.Value{
				Type: protocol.Array,
				Array: []protocol.Value{
					{Type: protocol.BulkString, Str: "INCR"},
					{Type: protocol.BulkString, Str: "counter"},
				},
			},
			want: Command{OP: "INCR", Key: "counter"},
		},
		{
			name: "unknown command",
			input: protocol.Value{
				Type: protocol.Array,
				Array: []protocol.Value{
					{Type: protocol.BulkString, Str: "DEL"},
					{Type: protocol.BulkString, Str: "username"},
				},
			},
			wantErr: ErrUnknownCommand,
		},
		{
			name: "incorrect GET argument count",
			input: protocol.Value{
				Type: protocol.Array,
				Array: []protocol.Value{
					{Type: protocol.BulkString, Str: "GET"},
				},
			},
			wantErr: ErrInvalidCommandLength,
		},
		{
			name: "non-array value",
			input: protocol.Value{
				Type: protocol.BulkString,
				Str:  "GET",
			},
			wantErr: fmt.Errorf("invalid value needs an array"),
		},
		{
			name: "empty array",
			input: protocol.Value{
				Type:  protocol.Array,
				Array: []protocol.Value{},
			},
			wantErr: fmt.Errorf("invalid length of the array"),
		},
		{
			name: "null bulk string argument",
			input: protocol.Value{
				Type: protocol.Array,
				Array: []protocol.Value{
					{Type: protocol.BulkString, Str: "GET"},
					{Type: protocol.BulkString, Num: -1},
				},
			},
			wantErr: fmt.Errorf("invalid command entry"),
		},
		{
			name: "empty key",
			input: protocol.Value{
				Type: protocol.Array,
				Array: []protocol.Value{
					{Type: protocol.BulkString, Str: "GET"},
					{Type: protocol.BulkString, Str: ""},
				},
			},
			wantErr: fmt.Errorf("invalid key length"),
		},
		{
			name: "lowercase command",
			input: protocol.Value{
				Type: protocol.Array,
				Array: []protocol.Value{
					{Type: protocol.BulkString, Str: "get"},
					{Type: protocol.BulkString, Str: "username"},
				},
			},
			want: Command{OP: "GET", Key: "username"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValueToCommand(tt.input)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}

				if err.Error() != tt.wantErr.Error() {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.want {
				t.Errorf("expected %+v, got %+v", tt.want, got)
			}
		})
	}
}
