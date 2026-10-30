package main

import (
	"errors"
	"testing"
)

func TestResponseEncodeSimpleString(t *testing.T) {
	response := NewSimpleString("OK")

	got, err := response.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	want := "+OK\r\n"

	if string(got) != want {
		t.Fatalf("Encode() = %q, want %q", string(got), want)
	}
}

func TestResponseEncodeError(t *testing.T) {
	response := NewError("ERR unknown command")

	got, err := response.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	want := "-ERR unknown command\r\n"

	if string(got) != want {
		t.Fatalf("Encode() = %q, want %q", string(got), want)
	}
}

func TestResponseEncodeInteger(t *testing.T) {
	response := NewInteger(42)

	got, err := response.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	want := ":42\r\n"

	if string(got) != want {
		t.Fatalf("Encode() = %q, want %q", string(got), want)
	}
}

func TestResponseEncodeBulkString(t *testing.T) {
	response := NewBulkString("你好")

	got, err := response.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	want := "$6\r\n你好\r\n"

	if string(got) != want {
		t.Fatalf("Encode() = %q, want %q", string(got), want)
	}
}

func TestResponseEncodeNullBulkString(t *testing.T) {
	response := NewNullBulkString()

	got, err := response.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	want := "$-1\r\n"

	if string(got) != want {
		t.Fatalf("Encode() = %q, want %q", string(got), want)
	}
}

func TestResponseEncodeNestedArray(t *testing.T) {
	response := NewArray([]Response{
		*NewInteger(10),
		*NewArray([]Response{
			*NewBulkString("hello"),
			*NewSimpleString("OK"),
		}),
	})

	got, err := response.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	want := "*2\r\n" +
		":10\r\n" +
		"*2\r\n" +
		"$5\r\nhello\r\n" +
		"+OK\r\n"

	if string(got) != want {
		t.Fatalf("Encode() = %q, want %q", string(got), want)
	}
}

func TestResponseEncodeInvalidType(t *testing.T) {
	response := &Response{
		responseType: ResponseType(100),
	}

	got, err := response.Encode()

	if !errors.Is(err, ErrInvalidResponseType) {
		t.Fatalf("Encode() error = %v, want %v", err, ErrInvalidResponseType)
	}

	if got != nil {
		t.Fatalf("Encode() = %v, want nil", got)
	}
}

func TestResponseEncodeNestedInvalidType(t *testing.T) {
	response := NewArray([]Response{
		*NewInteger(10),
		{
			responseType: ResponseType(100),
		},
	})

	got, err := response.Encode()

	if !errors.Is(err, ErrInvalidResponseType) {
		t.Fatalf("Encode() error = %v, want %v", err, ErrInvalidResponseType)
	}

	if got != nil {
		t.Fatalf("Encode() = %v, want nil", got)
	}
}
