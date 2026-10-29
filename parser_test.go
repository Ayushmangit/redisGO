package main

import (
	"testing"
)

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Command
		wantErr bool
	}{
		{
			name:  "SET",
			input: "SET name redis",
			want:  Command{OP: "SET", Key: "name", Value: "redis"},
		},
		{
			name:  "GET",
			input: "GET name",
			want:  Command{OP: "GET", Key: "name"},
		},
		{
			name:  "INCR",
			input: "INCR counter",
			want:  Command{OP: "INCR", Key: "counter"},
		},
		{
			name:  "case insensitive",
			input: "set name redis",
			want:  Command{OP: "SET", Key: "name", Value: "redis"},
		},
		{
			name:  "extra whitespace",
			input: "  SET   name   redis  ",
			want:  Command{OP: "SET", Key: "name", Value: "redis"},
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: true,
		},
		{
			name:    "unknown command",
			input:   "FLUSHALL",
			wantErr: true,
		},
		{
			name:    "SET missing value",
			input:   "SET name",
			wantErr: true,
		},
		{
			name:    "SET too many arguments",
			input:   "SET name redis extra",
			wantErr: true,
		},
		{
			name:    "GET missing key",
			input:   "GET",
			wantErr: true,
		},
		{
			name:    "GET too many arguments",
			input:   "GET name extra",
			wantErr: true,
		},
		{
			name:    "INCR missing key",
			input:   "INCR",
			wantErr: true,
		},
		{
			name:    "invalid command",
			input:   "FLUSHALL",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCommand(tt.input)

			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseCommand() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil {
				return
			}

			if got != tt.want {
				t.Errorf("ParseCommand() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
