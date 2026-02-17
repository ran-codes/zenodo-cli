package api

import (
	"net/http"
	"testing"
	"time"
)

func TestNewRateLimiter(t *testing.T) {
	rl := NewRateLimiter()
	if rl.general.maxTokens != 100 {
		t.Errorf("general max = %f, want 100", rl.general.maxTokens)
	}
	if rl.search.maxTokens != 30 {
		t.Errorf("search max = %f, want 30", rl.search.maxTokens)
	}
}

func TestIsSearchPath(t *testing.T) {
	tests := []struct {
		path   string
		search bool
	}{
		{"/records", true},
		{"/records/123", true},
		{"/communities", true},
		{"/licenses", true},
		{"/deposit/depositions", false},
		{"/api/user/records", false},
	}
	for _, tt := range tests {
		if got := isSearchPath(tt.path); got != tt.search {
			t.Errorf("isSearchPath(%q) = %v, want %v", tt.path, got, tt.search)
		}
	}
}

func TestRateLimiter_GeneralConsumes(t *testing.T) {
	rl := NewRateLimiter()

	// Consume a general token.
	startTokens := rl.general.tokens
	rl.Wait("/deposit/depositions/123")

	rl.mu.Lock()
	after := rl.general.tokens
	rl.mu.Unlock()

	// Should have consumed ~1 token (might refill a tiny amount).
	if after > startTokens-0.5 {
		t.Errorf("expected general tokens to decrease, before=%f after=%f", startTokens, after)
	}
}

func TestRateLimiter_SearchConsumesBoth(t *testing.T) {
	rl := NewRateLimiter()

	generalBefore := rl.general.tokens
	searchBefore := rl.search.tokens

	rl.Wait("/records")

	rl.mu.Lock()
	generalAfter := rl.general.tokens
	searchAfter := rl.search.tokens
	rl.mu.Unlock()

	if generalAfter > generalBefore-0.5 {
		t.Errorf("expected general tokens to decrease, before=%f after=%f", generalBefore, generalAfter)
	}
	if searchAfter > searchBefore-0.5 {
		t.Errorf("expected search tokens to decrease, before=%f after=%f", searchBefore, searchAfter)
	}
}

func TestBucket_Refill(t *testing.T) {
	b := newBucket(10, 60) // 1 token per second
	b.tokens = 0
	b.lastRefill = time.Now().Add(-2 * time.Second) // 2 seconds ago

	b.refill()

	// Should have ~2 tokens after refill.
	if b.tokens < 1.5 || b.tokens > 2.5 {
		t.Errorf("expected ~2 tokens after 2s refill, got %f", b.tokens)
	}
}

func TestBucket_RefillCapped(t *testing.T) {
	b := newBucket(10, 60)
	b.tokens = 10
	b.lastRefill = time.Now().Add(-10 * time.Second)

	b.refill()

	if b.tokens != 10 {
		t.Errorf("tokens should be capped at max, got %f", b.tokens)
	}
}

func TestBucket_WaitDuration(t *testing.T) {
	b := newBucket(100, 60) // 1 token per second
	b.tokens = 0.5

	wait := b.waitDuration()
	// Need 0.5 more tokens at 1/sec = ~500ms
	if wait < 400*time.Millisecond || wait > 600*time.Millisecond {
		t.Errorf("expected ~500ms wait, got %v", wait)
	}

	b.tokens = 5
	wait = b.waitDuration()
	if wait != 0 {
		t.Errorf("expected 0 wait with tokens available, got %v", wait)
	}
}

func TestUpdateFromHeaders(t *testing.T) {
	rl := NewRateLimiter()

	resp := &http.Response{
		Header: http.Header{
			"X-Ratelimit-Remaining": []string{"5"},
		},
	}

	rl.UpdateFromHeaders(resp, "/records")

	rl.mu.Lock()
	defer rl.mu.Unlock()

	if rl.general.tokens > 5.5 {
		t.Errorf("expected general tokens <=5 after header update, got %f", rl.general.tokens)
	}
	if rl.search.tokens > 5.5 {
		t.Errorf("expected search tokens <=5 after header update, got %f", rl.search.tokens)
	}
}

func TestUpdateFromHeaders_Nil(t *testing.T) {
	rl := NewRateLimiter()
	// Should not panic.
	rl.UpdateFromHeaders(nil, "/test")
}
