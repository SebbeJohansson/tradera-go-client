package tradera

import (
	"net/http"
	"time"
)

const DefaultBaseURL = "https://api.tradera.com"

// Config configures a Tradera REST API v4 client.
type Config struct {
	AppID  int
	AppKey string

	UserID int
	Token  string

	BaseURL    string
	HTTPClient *http.Client
	Headers    http.Header

	RateLimit      float64
	RetryEnabled   bool
	MaxRetries     int
	RetryBaseDelay time.Duration
	Timeout        time.Duration
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig(appID int, appKey string) Config {
	return Config{
		AppID:          appID,
		AppKey:         appKey,
		BaseURL:        DefaultBaseURL,
		MaxRetries:     3,
		RetryBaseDelay: time.Second,
		Timeout:        30 * time.Second,
	}
}

// WithUserAuth returns a copy of the config with user authentication set.
func (config Config) WithUserAuth(userID int, token string) Config {
	config.UserID = userID
	config.Token = token
	return config
}

// WithBaseURL returns a copy of the config with a custom API base URL.
func (config Config) WithBaseURL(baseURL string) Config {
	config.BaseURL = baseURL
	return config
}

// WithHTTPClient returns a copy of the config with a custom HTTP client.
func (config Config) WithHTTPClient(client *http.Client) Config {
	config.HTTPClient = client
	return config
}

// WithHeaders returns a copy of the config with additional request headers.
func (config Config) WithHeaders(headers http.Header) Config {
	config.Headers = headers.Clone()
	return config
}

// WithRateLimit returns a copy of the config with rate limiting enabled.
func (config Config) WithRateLimit(requestsPerSecond float64) Config {
	config.RateLimit = requestsPerSecond
	return config
}

// WithRetry returns a copy of the config with retry enabled.
func (config Config) WithRetry(maxRetries int, baseDelay time.Duration) Config {
	config.RetryEnabled = true
	config.MaxRetries = maxRetries
	config.RetryBaseDelay = baseDelay
	return config
}

// WithTimeout returns a copy of the config with the specified timeout.
func (config Config) WithTimeout(timeout time.Duration) Config {
	config.Timeout = timeout
	return config
}

// HasUserAuth reports whether user authentication is configured.
func (config Config) HasUserAuth() bool {
	return config.UserID > 0 && config.Token != ""
}

// Validate checks whether the configuration is valid.
func (config Config) Validate() error {
	if config.AppID <= 0 {
		return ErrInvalidAppID
	}
	if config.AppKey == "" {
		return ErrInvalidAppKey
	}
	if (config.UserID > 0) != (config.Token != "") {
		return ErrIncompleteUserAuth
	}
	return nil
}
