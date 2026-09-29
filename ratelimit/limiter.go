// Package ratelimit provides in-memory, per-key token bucket limiting.
//
// State lives in the process, so limits are per instance: behind N replicas
// the effective limit is N times higher. Use a shared store (e.g. Redis) or a
// gateway-level limit if that matters.
package ratelimit

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type entry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// KeyedLimiter keeps one token bucket per key and evicts idle keys, so memory
// stays bounded even when keys are attacker-controlled (IPs, emails).
type KeyedLimiter struct {
	rate  rate.Limit
	burst int
	ttl   time.Duration
	now   func() time.Time

	mu        sync.Mutex
	entries   map[string]*entry
	lastSweep time.Time
}

// NewKeyedLimiter allows `burst` events at once per key, refilling at `r` events per second.
// Keys idle for longer than ttl are forgotten.
func NewKeyedLimiter(r rate.Limit, burst int, ttl time.Duration) *KeyedLimiter {
	return &KeyedLimiter{
		rate:    r,
		burst:   burst,
		ttl:     ttl,
		now:     time.Now,
		entries: make(map[string]*entry),
	}
}

// PerMinute converts "n events per minute" to a rate.
func PerMinute(n float64) rate.Limit { return rate.Limit(n / 60) }

// get returns the bucket for key, creating it (full) if needed. Caller holds mu.
func (l *KeyedLimiter) get(key string, now time.Time) *entry {
	l.sweep(now)
	e, ok := l.entries[key]
	if !ok {
		e = &entry{limiter: rate.NewLimiter(l.rate, l.burst)}
		l.entries[key] = e
	}
	e.lastSeen = now
	return e
}

func (l *KeyedLimiter) sweep(now time.Time) {
	if now.Sub(l.lastSweep) < l.ttl {
		return
	}
	l.lastSweep = now
	for k, e := range l.entries {
		if now.Sub(e.lastSeen) > l.ttl {
			delete(l.entries, k)
		}
	}
}

// retryAfter is how long until one token is available given the current token count.
func (l *KeyedLimiter) retryAfter(tokens float64) time.Duration {
	if tokens >= 1 {
		return 0
	}
	return time.Duration((1 - tokens) / float64(l.rate) * float64(time.Second))
}

// Allow consumes a token for key. If none is available it returns false and how long to wait.
func (l *KeyedLimiter) Allow(key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	e := l.get(key, now)
	if e.limiter.AllowN(now, 1) {
		return true, 0
	}
	return false, l.retryAfter(e.limiter.TokensAt(now))
}

// Allowed reports whether a token is available for key without consuming it.
func (l *KeyedLimiter) Allowed(key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	tokens := l.get(key, now).limiter.TokensAt(now)
	if tokens >= 1 {
		return true, 0
	}
	return false, l.retryAfter(tokens)
}

// Fail consumes a token for key (used to count failed attempts).
func (l *KeyedLimiter) Fail(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	l.get(key, now).limiter.AllowN(now, 1)
}

// Reset forgets key, refilling its bucket.
func (l *KeyedLimiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.entries, key)
}

// Len returns the number of tracked keys (for tests and metrics).
func (l *KeyedLimiter) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.entries)
}
