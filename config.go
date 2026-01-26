package tradera

import "time"

// Config holds the configuration for the Tradera API client.
type Config struct {
	// AppID is your Tradera application ID (required)
	AppID int

	// AppKey is your Tradera application key/GUID (required)
	AppKey string

	// UserID is the user ID for authenticated operations (optional)
	// Required for Restricted, Order, and Buyer services
	UserID int

	// Token is the authorization token for authenticated operations (optional)
	// Required for Restricted, Order, and Buyer services
	// Obtain via PublicClient.FetchToken()
	Token string

	// RateLimit is the maximum number of requests per second (0 = disabled)
	RateLimit float64

	// RetryEnabled enables automatic retry with exponential backoff
	RetryEnabled bool

	// MaxRetries is the maximum number of retry attempts (default: 3)
	MaxRetries int

	// RetryBaseDelay is the base delay for exponential backoff (default: 1s)
	RetryBaseDelay time.Duration

	// CacheTTL enables caching with the specified TTL (0 = disabled)
	// Useful for caching relatively static data like categories
	CacheTTL time.Duration

	// Timeout is the default timeout for API requests (default: 30s)
	Timeout time.Duration
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig(appID int, appKey string) Config {
	return Config{
		AppID:          appID,
		AppKey:         appKey,
		RateLimit:      0, // disabled by default
		RetryEnabled:   false,
		MaxRetries:     3,
		RetryBaseDelay: time.Second,
		CacheTTL:       0, // disabled by default
		Timeout:        30 * time.Second,
	}
}

// WithUserAuth returns a copy of the config with user authentication set.
func (c Config) WithUserAuth(userID int, token string) Config {
	c.UserID = userID
	c.Token = token
	return c
}

// WithRateLimit returns a copy of the config with rate limiting enabled.
func (c Config) WithRateLimit(requestsPerSecond float64) Config {
	c.RateLimit = requestsPerSecond
	return c
}

// WithRetry returns a copy of the config with retry enabled.
func (c Config) WithRetry(maxRetries int, baseDelay time.Duration) Config {
	c.RetryEnabled = true
	c.MaxRetries = maxRetries
	c.RetryBaseDelay = baseDelay
	return c
}

// WithCache returns a copy of the config with caching enabled.
func (c Config) WithCache(ttl time.Duration) Config {
	c.CacheTTL = ttl
	return c
}

// WithTimeout returns a copy of the config with the specified timeout.
func (c Config) WithTimeout(timeout time.Duration) Config {
	c.Timeout = timeout
	return c
}

// HasUserAuth returns true if user authentication is configured.
func (c Config) HasUserAuth() bool {
	return c.UserID > 0 && c.Token != ""
}

// Validate checks if the configuration is valid.
func (c Config) Validate() error {
	if c.AppID <= 0 {
		return ErrInvalidAppID
	}
	if c.AppKey == "" {
		return ErrInvalidAppKey
	}
	return nil
}
