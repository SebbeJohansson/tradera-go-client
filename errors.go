package tradera

import (
	"errors"
	"fmt"
)

// Sentinel errors for common error conditions.
var (
	// ErrInvalidAppID is returned when the AppID is not set or invalid.
	ErrInvalidAppID = errors.New("tradera: invalid or missing AppID")

	// ErrInvalidAppKey is returned when the AppKey is not set or invalid.
	ErrInvalidAppKey = errors.New("tradera: invalid or missing AppKey")

	// ErrAuthRequired is returned when user authentication is required but not provided.
	ErrAuthRequired = errors.New("tradera: user authentication required (UserID and Token)")

	// ErrRateLimited is returned when the rate limit has been exceeded.
	ErrRateLimited = errors.New("tradera: rate limit exceeded")

	// ErrTimeout is returned when a request times out.
	ErrTimeout = errors.New("tradera: request timeout")

	// ErrNotFound is returned when the requested resource is not found.
	ErrNotFound = errors.New("tradera: resource not found")
)

// APIError represents an error returned by the Tradera API.
type APIError struct {
	// Code is the error code from the API.
	Code string

	// Message is the error message from the API.
	Message string

	// Details contains additional error details if available.
	Details string
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("tradera API error [%s]: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("tradera API error [%s]: %s", e.Code, e.Message)
}

// Is implements errors.Is for APIError.
func (e *APIError) Is(target error) bool {
	t, ok := target.(*APIError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// NewAPIError creates a new APIError.
func NewAPIError(code, message string) *APIError {
	return &APIError{Code: code, Message: message}
}

// NewAPIErrorWithDetails creates a new APIError with additional details.
func NewAPIErrorWithDetails(code, message, details string) *APIError {
	return &APIError{Code: code, Message: message, Details: details}
}

// SOAPFault represents a SOAP fault returned by the API.
type SOAPFault struct {
	FaultCode   string
	FaultString string
	Detail      string
}

// Error implements the error interface.
func (f *SOAPFault) Error() string {
	if f.Detail != "" {
		return fmt.Sprintf("SOAP fault [%s]: %s (%s)", f.FaultCode, f.FaultString, f.Detail)
	}
	return fmt.Sprintf("SOAP fault [%s]: %s", f.FaultCode, f.FaultString)
}

// NetworkError wraps network-related errors.
type NetworkError struct {
	Op  string // Operation that failed
	Err error  // Underlying error
}

// Error implements the error interface.
func (e *NetworkError) Error() string {
	return fmt.Sprintf("tradera network error during %s: %v", e.Op, e.Err)
}

// Unwrap implements errors.Unwrap.
func (e *NetworkError) Unwrap() error {
	return e.Err
}

// IsRetryable returns true if the error is potentially retryable.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Network errors are generally retryable
	var netErr *NetworkError
	if errors.As(err, &netErr) {
		return true
	}

	// Rate limit errors should be retried after waiting
	if errors.Is(err, ErrRateLimited) {
		return true
	}

	// Timeout errors might be retryable
	if errors.Is(err, ErrTimeout) {
		return true
	}

	// SOAP faults are generally not retryable (indicates a problem with the request)
	var soapFault *SOAPFault
	if errors.As(err, &soapFault) {
		return false
	}

	// API errors are generally not retryable
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return false
	}

	return false
}
