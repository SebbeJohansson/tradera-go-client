package middleware

import (
	"context"
	"sync"
	"time"
)

// RateLimiter provides rate limiting for API requests using a token bucket algorithm.
type RateLimiter struct {
	rate       float64    // tokens per second
	bucketSize float64    // max tokens (burst capacity)
	tokens     float64    // current tokens
	lastUpdate time.Time  // last token update time
	mu         sync.Mutex // protects tokens and lastUpdate
}

// NewRateLimiter creates a new rate limiter with the specified rate (requests per second).
// The bucket size (burst capacity) is set to the rate by default.
func NewRateLimiter(requestsPerSecond float64) *RateLimiter {
	return NewRateLimiterWithBurst(requestsPerSecond, requestsPerSecond)
}

// NewRateLimiterWithBurst creates a new rate limiter with custom burst capacity.
func NewRateLimiterWithBurst(requestsPerSecond, burstCapacity float64) *RateLimiter {
	return &RateLimiter{
		rate:       requestsPerSecond,
		bucketSize: burstCapacity,
		tokens:     burstCapacity, // start with full bucket
		lastUpdate: time.Now(),
	}
}

// Wait blocks until a token is available or the context is cancelled.
// Returns nil if a token was acquired, or the context error if cancelled.
func (r *RateLimiter) Wait(ctx context.Context) error {
	for {
		if r.TryAcquire() {
			return nil
		}

		// Calculate wait time until next token
		waitTime := r.timeUntilNextToken()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			// Try again
		}
	}
}

// TryAcquire attempts to acquire a token without blocking.
// Returns true if a token was acquired, false otherwise.
func (r *RateLimiter) TryAcquire() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.refillTokens()

	if r.tokens >= 1.0 {
		r.tokens -= 1.0
		return true
	}

	return false
}

// refillTokens adds tokens based on elapsed time since last update.
// Must be called with mutex held.
func (r *RateLimiter) refillTokens() {
	now := time.Now()
	elapsed := now.Sub(r.lastUpdate).Seconds()
	r.lastUpdate = now

	// Add tokens based on elapsed time
	r.tokens += elapsed * r.rate

	// Cap at bucket size
	if r.tokens > r.bucketSize {
		r.tokens = r.bucketSize
	}
}

// timeUntilNextToken returns the duration until the next token is available.
func (r *RateLimiter) timeUntilNextToken() time.Duration {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.tokens >= 1.0 {
		return 0
	}

	// Calculate time needed to accumulate 1 token
	tokensNeeded := 1.0 - r.tokens
	secondsNeeded := tokensNeeded / r.rate

	return time.Duration(secondsNeeded * float64(time.Second))
}

// Available returns the current number of available tokens.
func (r *RateLimiter) Available() float64 {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.refillTokens()
	return r.tokens
}

// Rate returns the rate limit in requests per second.
func (r *RateLimiter) Rate() float64 {
	return r.rate
}
