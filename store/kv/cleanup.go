package kv

import "time"

func (s *Store) cleanupExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for key, entry := range s.data {
		if !entry.expiry.IsZero() && !now.Before(entry.expiry) {
			delete(s.data, key)
		}
	}
}

func (s *Store) StartCleanup(interval time.Duration) {
	if interval <= 0 {
		panic("cleanup interval must be positive")
	}

	s.cleanupMu.Lock()
	defer s.cleanupMu.Unlock()

	if s.cleanupStarted || s.cleanupStopped {
		return
	}
	s.cleanupStarted = true

	go func() {
		defer s.cleanupDoneOnce.Do(func() {
			close(s.cleanupDone)
		})

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.cleanupExpired()
			case <-s.cleanupStop:
				return
			}
		}
	}()
}

func (s *Store) StopCleanup() {
	s.cleanupMu.Lock()

	if !s.cleanupStopped {
		s.cleanupStopped = true
		close(s.cleanupStop)
	}

	if !s.cleanupStarted {
		s.cleanupDoneOnce.Do(func() {
			close(s.cleanupDone)
		})
	}

	s.cleanupMu.Unlock()

	<-s.cleanupDone
}
