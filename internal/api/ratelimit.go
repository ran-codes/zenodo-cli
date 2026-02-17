package api

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// RateLimiter implements dual token bucket rate limiting for the Zenodo API.
// General bucket: 100 req/min, Search bucket: 30 req/min.
type RateLimiter struct {
	mu      sync.Mutex
	general *bucket
	search  *bucket
}

type bucket struct {
	tokens    float64
	maxTokens float64
	refillRate float64 // tokens per second
	lastRefill time.Time
}

func newBucket(maxTokens float64, perMinute float64) *bucket {
	return &bucket{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: perMinute / 60.0,
		lastRefill: time.Now(),
	}
}

func (b *bucket) refill() {
	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens += elapsed * b.refillRate
	if b.tokens > b.maxTokens {
		b.tokens = b.maxTokens
	}
	b.lastRefill = now
}

// waitDuration returns how long to wait before a token is available.
func (b *bucket) waitDuration() time.Duration {
	if b.tokens >= 1 {
		return 0
	}
	deficit := 1.0 - b.tokens
	return time.Duration(deficit/b.refillRate*1000) * time.Millisecond
}

func (b *bucket) consume() {
	b.tokens--
}

// NewRateLimiter creates a rate limiter with Zenodo's limits.
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		general: newBucket(100, 100),
		search:  newBucket(30, 30),
	}
}

// isSearchPath returns true if the path hits a search-limited endpoint.
func isSearchPath(path string) bool {
	searchPrefixes := []string{"/records", "/communities", "/licenses"}
	for _, prefix := range searchPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

// Wait blocks until the request is allowed under rate limits.
func (rl *RateLimiter) Wait(path string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Refill both buckets.
	rl.general.refill()
	rl.search.refill()

	// Check general bucket.
	if wait := rl.general.waitDuration(); wait > 0 {
		slog.Info("rate limiting: waiting for general bucket", "wait", wait)
		rl.mu.Unlock()
		time.Sleep(wait)
		rl.mu.Lock()
		rl.general.refill()
	}

	// Check search bucket for search endpoints.
	if isSearchPath(path) {
		if wait := rl.search.waitDuration(); wait > 0 {
			slog.Info("rate limiting: waiting for search bucket", "wait", wait)
			rl.mu.Unlock()
			time.Sleep(wait)
			rl.mu.Lock()
			rl.search.refill()
		}
		rl.search.consume()
	}

	rl.general.consume()
}

// UpdateFromHeaders reads X-RateLimit-Remaining and X-RateLimit-Reset headers
// to adjust the rate limiter state based on server feedback.
func (rl *RateLimiter) UpdateFromHeaders(resp *http.Response, path string) {
	if resp == nil {
		return
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	remaining := resp.Header.Get("X-RateLimit-Remaining")
	reset := resp.Header.Get("X-RateLimit-Reset")

	if remaining == "" {
		return
	}

	rem, err := strconv.ParseFloat(remaining, 64)
	if err != nil {
		return
	}

	// Update the appropriate bucket with server-reported remaining tokens.
	if isSearchPath(path) && rem < rl.search.tokens {
		slog.Debug("rate limit: server reports lower search remaining", "remaining", rem)
		rl.search.tokens = rem
	}

	if rem < rl.general.tokens {
		slog.Debug("rate limit: server reports lower general remaining", "remaining", rem)
		rl.general.tokens = rem
	}

	// If reset header is present, use it to schedule refill.
	if reset != "" {
		resetTime, err := strconv.ParseInt(reset, 10, 64)
		if err == nil {
			resetAt := time.Unix(resetTime, 0)
			untilReset := time.Until(resetAt)
			if untilReset > 0 && rem <= 5 {
				slog.Warn("rate limit: approaching limit, server resets in", "seconds", untilReset.Seconds(), "remaining", rem)
			}
		}
	}
}
