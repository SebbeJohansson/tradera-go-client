package tradera

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrInvalidAppID       = errors.New("tradera: invalid or missing AppID")
	ErrInvalidAppKey      = errors.New("tradera: invalid or missing AppKey")
	ErrIncompleteUserAuth = errors.New("tradera: both UserID and Token are required")
)

// APIError represents a non-successful response from the Tradera REST API.
type APIError struct {
	StatusCode int
	Body       []byte
	Header     http.Header
	RequestID  string
}

func (apiError *APIError) Error() string {
	if apiError.RequestID != "" {
		return fmt.Sprintf("tradera API request failed with status %d (request ID %s)", apiError.StatusCode, apiError.RequestID)
	}
	return fmt.Sprintf("tradera API request failed with status %d", apiError.StatusCode)
}

// IsRetryable reports whether an error is suitable for an automatic retry.
func IsRetryable(err error) bool {
	var apiError *APIError
	if errors.As(err, &apiError) {
		return apiError.StatusCode == http.StatusTooManyRequests || apiError.StatusCode >= http.StatusInternalServerError
	}
	return true
}
