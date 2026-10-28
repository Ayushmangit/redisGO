package kv

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/Ayushmangit/redisGO/store"
)

func TestKeys_ReturnAllKeysSorted(t *testing.T) {
	store := NewStore(5, 5*time.Second)
	store.Set("charlie", "3", store.ttl)
	store.Set("alpha", "1", store.ttl)
	store.Set("bravo", "2", store.ttl)

	got := store.Keys()
	want := []string{"alpha", "bravo", "charlie"}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Keys() = %v, want %v", got, want)
	}
}

func TestKeys_EmptyStore(t *testing.T) {
	store := NewStore(10, 5*time.Second)
	got := store.Keys()
	want := []string{}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Keys() = %v, want %v", got, want)
	}
}

func TestKeys_EmptyKey(t *testing.T) {
	s := NewStore(10, 5*time.Second)
	if _, err := s.Get(""); err == nil || !errors.Is(err, store.ErrInvalidKey) {
		t.Errorf("GET() failed")
	}
}

func TestSetGet_RoundTrip(t *testing.T) {
	ttl := time.Second * 5
	s := NewStore(10, ttl)

	s.Set("hello", "world", ttl)
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
	s := NewStore(10, 5*time.Second)

	s.Set("a", "1", s.ttl)
	s.Set("b", "2", s.ttl)
	s.Set("c", "3", s.ttl)
	s.Set("d", "4", s.ttl)

	length := s.Len()

	if length != len(s.data) {
		t.Errorf("Get() returned %v , wanted %v", length, len(s.data))
	}
}

func TestCouncurrentIncr(t *testing.T) {
	s := NewStore(10, 5*time.Second)
	err := s.Set("pageviews", "0", s.ttl)
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
	s := NewStore(100, 5*time.Second)
	for i := range 100 {
		key := fmt.Sprintf("pageviews:%d", i)
		if err := s.Set(key, "0", s.ttl); err != nil {
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
	s := NewStore(1, 5*time.Second)
	if err := s.Set("pageviews", "0", s.ttl); err != nil {
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

func TestSetWithTTLExpires(t *testing.T) {
	s := NewStore(10, 0)

	err := s.Set("session", "abc", 50*time.Millisecond)
	if err != nil {
		t.Fatalf("Set() failed: %v", err)
	}

	got, err := s.Get("session")
	if err != nil {
		t.Fatalf("Get() before expiry failed: %v", err)
	}
	if got != "abc" {
		t.Fatalf("Get() = %q, want %q", got, "abc")
	}

	time.Sleep(100 * time.Millisecond)

	_, err = s.Get("session")
	if !errors.Is(err, store.ErrKeyNotFound) {
		t.Fatalf("Get() after expiry error = %v, want %v",
			err, store.ErrKeyNotFound)
	}
}

func TestSetWithZeroTTLDoesNotExpire(t *testing.T) {
	s := NewStore(10, 0)

	err := s.Set("permanent", "value", 0)
	if err != nil {
		t.Fatalf("Set() failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	got, err := s.Get("permanent")
	if err != nil {
		t.Fatalf("Get() failed for non-expiring key: %v", err)
	}

	if got != "value" {
		t.Fatalf("Get() = %q, want %q", got, "value")
	}
}

func TestSetWithNegativeTTL(t *testing.T) {
	s := NewStore(10, 0)

	err := s.Set("invalid", "value", -time.Second)
	if err == nil {
		t.Fatal("Set() with negative TTL succeeded, want an error")
	}

	_, err = s.Get("invalid")
	if !errors.Is(err, store.ErrKeyNotFound) {
		t.Fatalf("Get() error = %v, want %v", err, store.ErrKeyNotFound)
	}
}

func TestSetResetsTTL(t *testing.T) {
	s := NewStore(10, 0)

	if err := s.Set("session", "old", 150*time.Millisecond); err != nil {
		t.Fatalf("Initial Set() failed: %v", err)
	}

	time.Sleep(80 * time.Millisecond)

	if err := s.Set("session", "new", 500*time.Millisecond); err != nil {
		t.Fatalf("Updating Set() failed: %v", err)
	}

	// The original TTL would have expired by now.
	time.Sleep(100 * time.Millisecond)

	got, err := s.Get("session")
	if err != nil {
		t.Fatalf("Get() failed after TTL reset: %v", err)
	}

	if got != "new" {
		t.Fatalf("Get() = %q, want %q", got, "new")
	}
}

//
// func TestIncrExpiredKey(t *testing.T) {
// 	s := NewStore(10, 0)
//
// 	if err := s.Set("counter", "10", 50*time.Millisecond); err != nil {
// 		t.Fatalf("Set() failed: %v", err)
// 	}
//
// 	time.Sleep(100 * time.Millisecond)
//
// 	_, err := s.Incr("counter")
// 	if !errors.Is(err, store.ErrKeyNotFound) {
// 		t.Fatalf("Incr() error = %v, want %v", err, store.ErrKeyNotFound)
// 	}
// }

func TestIncrMissingKey(t *testing.T) {
	s := NewStore(10, 0)

	got, err := s.Incr("counter")
	if err != nil {
		t.Fatalf("Incr() failed: %v", err)
	}

	if got != 1 {
		t.Fatalf("Incr() = %d, want 1", got)
	}

	value, err := s.Get("counter")
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	if value != "1" {
		t.Fatalf("Get() = %q, want %q", value, "1")
	}
}

func TestIncrMissingKeyUsesDefaultTTL(t *testing.T) {
	s := NewStore(10, 100*time.Millisecond)

	got, err := s.Incr("counter")
	if err != nil {
		t.Fatalf("Incr() failed: %v", err)
	}
	if got != 1 {
		t.Fatalf("Incr() = %d, want 1", got)
	}

	time.Sleep(150 * time.Millisecond)

	// INCR on the expired key should recreate it at 1.
	got, err = s.Incr("counter")
	if err != nil {
		t.Fatalf("Incr() after expiry failed: %v", err)
	}
	if got != 1 {
		t.Fatalf("Incr() after expiry = %d, want 1", got)
	}
}

func TestIncrPreservesTTL(t *testing.T) {
	s := NewStore(10, 0)

	if err := s.Set("counter", "10", 150*time.Millisecond); err != nil {
		t.Fatalf("Set() failed: %v", err)
	}

	time.Sleep(80 * time.Millisecond)

	got, err := s.Incr("counter")
	if err != nil {
		t.Fatalf("Incr() failed: %v", err)
	}
	if got != 11 {
		t.Fatalf("Incr() = %d, want 11", got)
	}

	// Wait beyond the original expiration time.
	time.Sleep(100 * time.Millisecond)

	_, err = s.Get("counter")
	if !errors.Is(err, store.ErrKeyNotFound) {
		t.Fatalf("Get() error = %v, want %v", err, store.ErrKeyNotFound)
	}
}
