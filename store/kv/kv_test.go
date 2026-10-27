package kv

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"sync"
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

func TestCouncurrentIncr(t *testing.T) {
	s := NewStore(10)
	err := s.Set("pageviews", "0")
	if err != nil {
		fmt.Errorf("SET() %v", err)
	}
	routines := 100
	var wg sync.WaitGroup
	errs := make(chan error)

	for range routines {
		wg.Go(func() {
			for range 1000 {
				_, err := s.Incr("pageviews")
				if err != nil {
					errs <- err
					return
				}
			}
		})
	}
	go func() {
		wg.Wait()
		close(errs)
	}()
	for err := range errs {
		t.Errorf("INCR() failed: %v", err)
	}
	got, err := s.Get("pageviews")
	if err != nil {
		t.Errorf("GET() failed")
	}
	gotInt, err := strconv.Atoi(got)
	if err != nil {
		fmt.Errorf("ATOI() %v", err)
	}
	expected := 100000

	if expected != gotInt {
		t.Errorf("INCR() returned %v , wanted %v", gotInt, expected)
	}
}

func TestConcurrentIncrDiffKeys(t *testing.T) {
	s := NewStore(100)
	for i := range 100 {
		key := fmt.Sprintf("pageviews:%d", i)
		if err := s.Set(key, "0"); err != nil {
			t.Errorf("SET() failed: %v\n", key)
			return
		}
	}
	var wg sync.WaitGroup
	errs := make(chan error)

	for i := range 100 {
		wg.Go(func() {
			key := fmt.Sprintf("pageviews:%d", i)
			for range 1000 {
				_, err := s.Incr(key)
				if err != nil {
					errs <- err
					return
				}
			}
		})
	}

	go func() {
		wg.Wait()
		close(errs)
	}()

	for err := range errs {
		t.Errorf("INCR() failed: %v", err)
	}
	expected := 1000

	for i := range 100 {
		key := fmt.Sprintf("pageviews:%d", i)

		got, err := s.Get(key)
		if err != nil {
			t.Errorf("GET(%s) failed: %v", key, err)
			continue
		}

		gotInt, err := strconv.Atoi(got)
		if err != nil {
			t.Errorf("Atoi(%s) failed: %v", key, err)
			continue
		}

		if gotInt != expected {
			t.Errorf("key %s = %d, want %d", key, gotInt, expected)
		}
	}
}

func TestConcurrentIncrWithGets(t *testing.T) {
	s := NewStore(1)
	if err := s.Set("pageviews", "0"); err != nil {
		t.Fatalf("SET() failed: %v", err)
	}

	routines := 100
	var wg sync.WaitGroup
	errs := make(chan error, routines*2)
	start := make(chan struct{})

	for range routines {
		wg.Go(func() {
			// routine is wating for the starting gun
			<-start

			for range 1000 {
				if _, err := s.Incr("pageviews"); err != nil {
					errs <- err
					return
				}
			}
		})
	}

	for range routines {
		wg.Go(func() {
			// routine is waiting for the starting gun
			<-start

			for range 1000 {
				if _, err := s.Get("pageviews"); err != nil {
					errs <- err
					return
				}
			}
		})
	}

	// starting gun boom
	close(start)

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("Concurrent operation failed: %v", err)
	}

	got, err := s.Get("pageviews")
	if err != nil {
		t.Fatalf("GET() failed: %v", err)
	}

	gotInt, err := strconv.Atoi(got)
	if err != nil {
		t.Fatalf("ATOI() failed: %v", err)
	}

	expected := routines * 1000
	if gotInt != expected {
		t.Errorf("INCR() returned %v, wanted %v", gotInt, expected)
	}
}
