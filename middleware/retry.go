package middleware

import (
	"context"
	"math"
	"math/rand"
	"time"
)

// RetryConfig holds configuration for retry behavior.
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts.
	MaxRetries int

	// BaseDelay is the initial delay before the first retry.
	BaseDelay time.Duration

	// MaxDelay is the maximum delay between retries.
	MaxDelay time.Duration

	// Multiplier is the factor by which the delay increases after each retry.
	Multiplier float64

	// Jitter adds randomness to the delay to prevent thundering herd.
	// Value between 0 (no jitter) and 1 (full jitter).
	Jitter float64

	// ShouldRetry is a function that determines if an error is retryable.
	// If nil, all errors are considered retryable.
	ShouldRetry func(error) bool
}

// DefaultRetryConfig returns a RetryConfig with sensible defaults.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: 3,
		BaseDelay:  time.Second,
		MaxDelay:   30 * time.Second,
		Multiplier: 2.0,
		Jitter:     0.2,
		ShouldRetry: func(err error) bool {
			return true // Override with IsRetryable from errors.go in actual use
		},
	}
}

// Retryer provides retry functionality with exponential backoff.
type Retryer struct {
	config RetryConfig
}

// NewRetryer creates a new Retryer with the given configuration.
func NewRetryer(config RetryConfig) *Retryer {
	if config.MaxRetries <= 0 {
		config.MaxRetries = 3
	}
	if config.BaseDelay <= 0 {
		config.BaseDelay = time.Second
	}
	if config.MaxDelay <= 0 {
		config.MaxDelay = 30 * time.Second
	}
	if config.Multiplier <= 0 {
		config.Multiplier = 2.0
	}
	if config.Jitter < 0 {
		config.Jitter = 0
	}
	if config.Jitter > 1 {
		config.Jitter = 1
	}

	return &Retryer{config: config}
}

// Do executes the given function with retry logic.
// Returns the result of the function or the last error if all retries fail.
func (r *Retryer) Do(ctx context.Context, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt <= r.config.MaxRetries; attempt++ {
		// Execute the function
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if we should retry
		if r.config.ShouldRetry != nil && !r.config.ShouldRetry(err) {
			return err
		}

		// Don't wait after the last attempt
		if attempt == r.config.MaxRetries {
			break
		}

		// Calculate delay with exponential backoff
		delay := r.calculateDelay(attempt)

		// Wait or return if context is cancelled
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	return lastErr
}

// DoWithResult executes a function that returns a value with retry logic.
func DoWithResult[T any](ctx context.Context, r *Retryer, fn func() (T, error)) (T, error) {
	var result T
	var lastErr error

	for attempt := 0; attempt <= r.config.MaxRetries; attempt++ {
		var err error
		result, err = fn()
		if err == nil {
			return result, nil
		}

		lastErr = err

		if r.config.ShouldRetry != nil && !r.config.ShouldRetry(err) {
			return result, err
		}

		if attempt == r.config.MaxRetries {
			break
		}

		delay := r.calculateDelay(attempt)

		select {
		case <-ctx.Done():
			return result, ctx.Err()
		case <-time.After(delay):
		}
	}

	return result, lastErr
}

// calculateDelay calculates the delay for a given attempt number.
func (r *Retryer) calculateDelay(attempt int) time.Duration {
	// Calculate exponential backoff
	delay := float64(r.config.BaseDelay) * math.Pow(r.config.Multiplier, float64(attempt))

	// Apply jitter
	if r.config.Jitter > 0 {
		jitterRange := delay * r.config.Jitter
		delay = delay - jitterRange + (rand.Float64() * 2 * jitterRange)
	}

	// Cap at max delay
	if delay > float64(r.config.MaxDelay) {
		delay = float64(r.config.MaxDelay)
	}

	return time.Duration(delay)
}

// Config returns a copy of the retry configuration.
func (r *Retryer) Config() RetryConfig {
	return r.config
}
