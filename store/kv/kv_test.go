package kv

import (
	"errors"
	"reflect"
	"testing"

	"github.com/Ayushmangit/redisGO/store"
)

func TestKeys_ReturnAllKeysSorted(t *testing.T) {
	store := NewStore(5)
	store.Set("charlie", "3")
	store.Set("alpha", "1")
	store.Set("bravo", "2")

	got := store.Keys()
	want := []string{"alpha", "bravo", "charlie"}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Keys() = %v, want %v", got, want)
	}
}

func TestKeys_EmptyStore(t *testing.T) {
	store := NewStore(10)
	got := store.Keys()
	want := []string{}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Keys() = %v, want %v", got, want)
	}
}

func TestKeys_EmptyKey(t *testing.T) {
	s := NewStore(10)
	if _, err := s.Get(""); err == nil || !errors.Is(err, store.ErrInvalidKey) {
		t.Errorf("GET() failed")
	}
}

func TestSetGet_RoundTrip(t *testing.T) {
	s := NewStore(10)

	s.Set("hello", "world")
	if val, err := s.Get("hello"); err != nil || val != "world" {
		t.Errorf("Get() returned %v , wanted %v", val, "world")
	}

	val, err := s.Get("missing")
	if err != nil {
		if !errors.Is(err, store.ErrKeyNotFound) {
			t.Errorf("Got(ERROR: %v) wanted (ERROR:%v)", err, store.ErrKeyNotFound)
		}
	}
	if val != "" {
		t.Errorf("Got(%v) wanted (%v)", val, "")
	}
}

func Test_Len(t *testing.T) {
	s := NewStore(10)

	s.Set("a", "1")
	s.Set("b", "2")
	s.Set("c", "3")
	s.Set("d", "4")

	length := s.Len()

	if length != len(s.data) {
		t.Errorf("Get() returned %v , wanted %v", length, len(s.data))
	}
}
