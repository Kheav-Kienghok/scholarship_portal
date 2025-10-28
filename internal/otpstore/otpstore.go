package otpstore

import (
	"sync"
	"time"
)

type attempt struct {
	count     int
	firstSeen time.Time
}

type OTPStore struct {
	mu       sync.Mutex
	attempts map[string]*attempt
	max      int
	window   time.Duration
}

func NewOTPStore(max int, window time.Duration) *OTPStore {
	s := &OTPStore{
		attempts: make(map[string]*attempt),
		max:      max,
		window:   window,
	}
	go s.cleanupLoop()
	return s
}

// Remaining returns remaining attempts and whether the key is locked
func (s *OTPStore) Remaining(key string) (int, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	a, ok := s.attempts[key]
	if !ok {
		return s.max, false
	}
	if time.Since(a.firstSeen) > s.window {
		delete(s.attempts, key)
		return s.max, false
	}
	remaining := s.max - a.count
	return remaining, remaining <= 0
}

// Increment failed attempt. Returns remaining attempts and locked status
func (s *OTPStore) Increment(key string) (int, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	a, ok := s.attempts[key]
	if !ok || time.Since(a.firstSeen) > s.window {
		s.attempts[key] = &attempt{count: 1, firstSeen: now}
		return s.max - 1, s.max-1 <= 0
	}
	a.count++
	remaining := s.max - a.count
	return remaining, remaining <= 0
}

// Reset clears attempts for a key
func (s *OTPStore) Reset(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.attempts, key)
}

// cleanupLoop removes expired entries periodically
func (s *OTPStore) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		cutoff := time.Now().Add(-s.window)
		s.mu.Lock()
		for k, v := range s.attempts {
			if v.firstSeen.Before(cutoff) {
				delete(s.attempts, k)
			}
		}
		s.mu.Unlock()
	}
}

// Window returns the configured lock window duration
func (s *OTPStore) Window() time.Duration {
	return s.window
}
