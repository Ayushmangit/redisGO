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

func TestKeys_IgnoresExpiredKeys(t *testing.T) {
	s := NewStore(10, 0)

	if err := s.Set("active", "1", time.Minute); err != nil {
		t.Fatal(err)
	}
	if err := s.Set("expired", "2", 50*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if err := s.Set("another", "3", time.Minute); err != nil {
		t.Fatal(err)
	}

	time.Sleep(100 * time.Millisecond)

	got := s.Keys()
	want := []string{"active", "another"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Keys() = %v, want %v", got, want)
	}
}

func TestLen_IgnoresExpiredKeys(t *testing.T) {
	s := NewStore(10, 0)

	if err := s.Set("active", "1", time.Minute); err != nil {
		t.Fatal(err)
	}

	if err := s.Set("expired", "2", 50*time.Millisecond); err != nil {
		t.Fatal(err)
	}

	time.Sleep(100 * time.Millisecond)

	got := s.Len()
	want := 1

	if got != want {
		t.Fatalf("Len() = %d, want %d", got, want)
	}
}

func TestDelete_ExistingKey(t *testing.T) {
	s := NewStore(10, 0)

	if err := s.Set("name", "redis", 0); err != nil {
		t.Fatal(err)
	}

	if err := s.Delete("name"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err := s.Get("name")
	if !errors.Is(err, store.ErrKeyNotFound) {
		t.Fatalf("Get() error = %v, want %v", err, store.ErrKeyNotFound)
	}
}

func TestDelete_MissingKey(t *testing.T) {
	s := NewStore(10, 0)

	if err := s.Delete("missing"); err != nil {
		t.Fatalf("Delete() error = %v, want nil", err)
	}
}

func TestPop_ExistingKey(t *testing.T) {
	s := NewStore(10, 0)

	if err := s.Set("name", "redis", 0); err != nil {
		t.Fatal(err)
	}

	got, err := s.Pop("name")
	if err != nil {
		t.Fatalf("Pop() error = %v", err)
	}
	if got != "redis" {
		t.Fatalf("Pop() = %q, want %q", got, "redis")
	}

	_, err = s.Get("name")
	if !errors.Is(err, store.ErrKeyNotFound) {
		t.Fatalf("Get() error = %v, want %v", err, store.ErrKeyNotFound)
	}
}

func TestPop_MissingKey(t *testing.T) {
	s := NewStore(10, 0)

	_, err := s.Pop("missing")
	if !errors.Is(err, store.ErrKeyNotFound) {
		t.Fatalf("Pop() error = %v, want %v", err, store.ErrKeyNotFound)
	}
}

func TestPop_ExpiredKey(t *testing.T) {
	s := NewStore(10, 0)

	if err := s.Set("temporary", "value", 50*time.Millisecond); err != nil {
		t.Fatal(err)
	}

	time.Sleep(100 * time.Millisecond)

	_, err := s.Pop("temporary")
	if !errors.Is(err, store.ErrKeyNotFound) {
		t.Fatalf("Pop() error = %v, want %v", err, store.ErrKeyNotFound)
	}
}

func TestRename_ExistingKey(t *testing.T) {
	s := NewStore(10, 0)

	if err := s.Set("old", "redis", 0); err != nil {
		t.Fatal(err)
	}

	if err := s.Rename("old", "new"); err != nil {
		t.Fatalf("Rename() error = %v", err)
	}

	got, err := s.Get("new")
	if err != nil {
		t.Fatalf("Get(new) error = %v", err)
	}
	if got != "redis" {
		t.Fatalf("Get(new) = %q, want %q", got, "redis")
	}

	_, err = s.Get("old")
	if !errors.Is(err, store.ErrKeyNotFound) {
		t.Fatalf("Get(old) error = %v, want %v", err, store.ErrKeyNotFound)
	}
}

func TestRename_DestinationKeyExists(t *testing.T) {
	s := NewStore(10, 0)

	if err := s.Set("old", "redis", 0); err != nil {
		t.Fatal(err)
	}
	if err := s.Set("new", "existing", 0); err != nil {
		t.Fatal(err)
	}

	if err := s.Rename("old", "new"); err != nil {
		t.Fatalf("Rename() error = %v", err)
	}

	got, err := s.Get("new")
	if err != nil {
		t.Fatalf("Get(new) error = %v", err)
	}
	if got != "redis" {
		t.Fatalf("Get(new) = %q, want %q", got, "redis")
	}

	_, err = s.Get("old")
	if !errors.Is(err, store.ErrKeyNotFound) {
		t.Fatalf("Get(old) error = %v, want %v", err, store.ErrKeyNotFound)
	}
}

func TestRename_PreservesTTL(t *testing.T) {
	s := NewStore(10, 0)

	if err := s.Set("old", "redis", 300*time.Millisecond); err != nil {
		t.Fatal(err)
	}

	time.Sleep(150 * time.Millisecond)

	if err := s.Rename("old", "new"); err != nil {
		t.Fatalf("Rename() error = %v", err)
	}

	// The renamed key should still exist shortly after the rename.
	got, err := s.Get("new")
	if err != nil {
		t.Fatalf("Get(new) error = %v", err)
	}
	if got != "redis" {
		t.Fatalf("Get(new) = %q, want %q", got, "redis")
	}

	// Wait beyond the original expiration time.
	time.Sleep(200 * time.Millisecond)

	_, err = s.Get("new")
	if !errors.Is(err, store.ErrKeyNotFound) {
		t.Fatalf("Get(new) error = %v, want %v", err, store.ErrKeyNotFound)
	}
}

func TestRename_MissingSourceKey(t *testing.T) {
	s := NewStore(10, 0)

	err := s.Rename("missing", "new")
	if !errors.Is(err, store.ErrKeyNotFound) {
		t.Fatalf("Rename() error = %v, want %v", err, store.ErrKeyNotFound)
	}

	_, err = s.Get("new")
	if !errors.Is(err, store.ErrKeyNotFound) {
		t.Fatalf("Get(new) error = %v, want %v", err, store.ErrKeyNotFound)
	}
}

func TestRename_ExpiredSourceKey(t *testing.T) {
	s := NewStore(10, 0)

	if err := s.Set("old", "redis", 50*time.Millisecond); err != nil {
		t.Fatal(err)
	}

	time.Sleep(100 * time.Millisecond)

	err := s.Rename("old", "new")
	if !errors.Is(err, store.ErrKeyNotFound) {
		t.Fatalf("Rename() error = %v, want %v", err, store.ErrKeyNotFound)
	}

	_, err = s.Get("new")
	if !errors.Is(err, store.ErrKeyNotFound) {
		t.Fatalf("Get(new) error = %v, want %v", err, store.ErrKeyNotFound)
	}
}

func TestRename_SameKey(t *testing.T) {
	s := NewStore(10, 0)

	if err := s.Set("key", "redis", 0); err != nil {
		t.Fatal(err)
	}

	if err := s.Rename("key", "key"); err != nil {
		t.Fatalf("Rename() error = %v", err)
	}

	got, err := s.Get("key")
	if err != nil {
		t.Fatalf("Get(key) error = %v", err)
	}

	if got != "redis" {
		t.Fatalf("Get(key) = %q, want %q", got, "redis")
	}
}

func TestClone_IndependentCopy(t *testing.T) {
	s := NewStore(10, 0)

	if err := s.Set("name", "redis", 0); err != nil {
		t.Fatal(err)
	}

	cloned := s.Clone()

	// The clone should contain the original value.
	got, err := cloned.Get("name")
	if err != nil {
		t.Fatalf("Clone.Get(name) error = %v", err)
	}
	if got != "redis" {
		t.Fatalf("Clone.Get(name) = %q, want %q", got, "redis")
	}

	// Changing the original must not change the clone.
	if err := s.Set("name", "changed", 0); err != nil {
		t.Fatal(err)
	}

	got, err = cloned.Get("name")
	if err != nil {
		t.Fatalf("Clone.Get(name) after original update: %v", err)
	}
	if got != "redis" {
		t.Fatalf("Clone.Get(name) = %q, want %q", got, "redis")
	}
}

func TestClone_PreservesTTL(t *testing.T) {
	s := NewStore(10, 0)

	if err := s.Set("temporary", "redis", 300*time.Millisecond); err != nil {
		t.Fatal(err)
	}

	time.Sleep(150 * time.Millisecond)

	cloned := s.Clone()

	// The cloned key should still exist.
	got, err := cloned.Get("temporary")
	if err != nil {
		t.Fatalf("Clone.Get() error = %v", err)
	}
	if got != "redis" {
		t.Fatalf("Clone.Get() = %q, want %q", got, "redis")
	}

	// Both stores should respect the original expiration time.
	time.Sleep(200 * time.Millisecond)

	_, err = cloned.Get("temporary")
	if !errors.Is(err, store.ErrKeyNotFound) {
		t.Fatalf("Clone.Get() error = %v, want %v", err, store.ErrKeyNotFound)
	}
}

func TestClone_PreservesConfiguration(t *testing.T) {
	defaultTTL := 5 * time.Second
	s := NewStore(20, defaultTTL)

	cloned := s.Clone()

	if cloned.MaxSize != s.MaxSize {
		t.Errorf("Clone.MaxSize = %d, want %d", cloned.MaxSize, s.MaxSize)
	}

	if cloned.ttl != s.ttl {
		t.Errorf("Clone.ttl = %v, want %v", cloned.ttl, s.ttl)
	}
}

func TestSet_CapacityFull(t *testing.T) {
	s := NewStore(1, 0)

	if err := s.Set("first", "one", 0); err != nil {
		t.Fatal(err)
	}

	err := s.Set("second", "two", 0)
	if !errors.Is(err, store.ErrCapacityFull) {
		t.Fatalf("Set() error = %v, want %v", err, store.ErrCapacityFull)
	}

	got, err := s.Get("first")
	if err != nil {
		t.Fatalf("Get(first) error = %v", err)
	}
	if got != "one" {
		t.Fatalf("Get(first) = %q, want %q", got, "one")
	}
}

func TestSet_ExpiredKeyDoesNotBlockCapacity(t *testing.T) {
	s := NewStore(1, 0)

	if err := s.Set("temporary", "old", 50*time.Millisecond); err != nil {
		t.Fatal(err)
	}

	time.Sleep(100 * time.Millisecond)

	if err := s.Set("new", "value", 0); err != nil {
		t.Fatalf("Set() error = %v, want nil", err)
	}

	got, err := s.Get("new")
	if err != nil {
		t.Fatalf("Get(new) error = %v", err)
	}
	if got != "value" {
		t.Fatalf("Get(new) = %q, want %q", got, "value")
	}
}

func TestCleanupWorker_RemovesExpiredEntries(t *testing.T) {
	s := NewStore(10, 0)

	if err := s.Set("temporary", "value", 50*time.Millisecond); err != nil {
		t.Fatal(err)
	}

	s.StartCleanup(20 * time.Millisecond)
	defer s.StopCleanup()

	deadline := time.Now().Add(2 * time.Second)

	for time.Now().Before(deadline) {
		s.mu.RLock()
		_, exists := s.data["temporary"]
		s.mu.RUnlock()

		if !exists {
			return
		}

		time.Sleep(10 * time.Millisecond)
	}

	t.Fatal("cleanup worker did not remove expired entry before deadline")
}

func TestCleanupWorker_Stop(t *testing.T) {
	s := NewStore(10, 0)

	s.StartCleanup(10 * time.Millisecond)

	// StopCleanup should wait for the worker to exit.
	s.StopCleanup()

	// Repeated calls should be safe.
	s.StopCleanup()

	// Starting again after stopping should not restart the worker.
	s.StartCleanup(10 * time.Millisecond)

	if !s.cleanupStopped {
		t.Fatal("expected cleanup worker to be marked stopped")
	}
}

func TestClone_IndependentCleanupWorker(t *testing.T) {
	s := NewStore(10, 0)

	if err := s.Set("temporary", "value", 100*time.Millisecond); err != nil {
		t.Fatal(err)
	}

	s.StartCleanup(10 * time.Millisecond)
	defer s.StopCleanup()

	cloned := s.Clone()

	// The clone should have its own cleanup lifecycle.
	cloned.StartCleanup(10 * time.Millisecond)
	defer cloned.StopCleanup()

	time.Sleep(200 * time.Millisecond)

	s.mu.RLock()
	_, originalExists := s.data["temporary"]
	s.mu.RUnlock()

	cloned.mu.RLock()
	_, cloneExists := cloned.data["temporary"]
	cloned.mu.RUnlock()

	if originalExists {
		t.Error("original store should have cleaned up the expired key")
	}
	if cloneExists {
		t.Error("cloned store should have cleaned up the expired key")
	}
}

func TestCleanupWorker_ConcurrentOperations(t *testing.T) {
	s := NewStore(1000, 0)

	s.StartCleanup(5 * time.Millisecond)
	defer s.StopCleanup()

	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			for j := 0; j < 100; j++ {
				key := fmt.Sprintf("key-%d-%d", id, j)

				if err := s.Set(key, "value", 10*time.Millisecond); err != nil {
					t.Errorf("Set() error = %v", err)
					return
				}

				_, _ = s.Get(key)
				_ = s.Delete(key)
			}
		}(i)
	}

	wg.Wait()
}

func TestCleanupWorker_StopAfterStart(t *testing.T) {
	s := NewStore(10, 0)

	s.StartCleanup(5 * time.Millisecond)
	s.StopCleanup()

	select {
	case <-s.cleanupDone:
		// Worker has exited.
	default:
		t.Fatal("cleanup worker has not exited")
	}
}

func TestCleanupWorker_StopBeforeStart(t *testing.T) {
	s := NewStore(10, 0)

	s.StopCleanup()

	select {
	case <-s.cleanupDone:
		// Shutdown completed successfully.
	default:
		t.Fatal("cleanupDone should be closed after stopping before start")
	}

	// Starting after stopping must not launch a worker.
	s.StartCleanup(5 * time.Millisecond)
}

func TestCleanupWorker_StopMultipleTimes(t *testing.T) {
	s := NewStore(10, 0)

	s.StartCleanup(5 * time.Millisecond)

	s.StopCleanup()
	s.StopCleanup()
	s.StopCleanup()
}

func TestCleanupWorker_RemovesExpiredKeys(t *testing.T) {
	s := NewStore(10, 0)

	if err := s.Set("temp", "value", 20*time.Millisecond); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	s.StartCleanup(5 * time.Millisecond)
	defer s.StopCleanup()

	time.Sleep(50 * time.Millisecond)

	s.mu.RLock()
	_, exists := s.data["temp"]
	s.mu.RUnlock()

	if exists {
		t.Fatal("cleanup worker did not physically remove expired key")
	}
}

func TestClone_CleanupLifecycleIndependent(t *testing.T) {
	original := NewStore(10, 0)
	clone := original.Clone()

	original.StartCleanup(5 * time.Millisecond)
	clone.StartCleanup(5 * time.Millisecond)

	original.StopCleanup()

	// The clone's worker must still be running.
	select {
	case <-clone.cleanupDone:
		t.Fatal("clone's cleanup worker stopped when original stopped")
	default:
	}

	clone.StopCleanup()
}

func TestStore_WorksWithoutCleanupWorker(t *testing.T) {
	s := NewStore(10, 0)

	if err := s.Set("name", "redisGO", 0); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	got, err := s.Get("name")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if got != "redisGO" {
		t.Fatalf("Get() = %q, want %q", got, "redisGO")
	}
}
