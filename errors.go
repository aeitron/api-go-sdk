package aeitron

import (
	"fmt"
	"time"
)

// APIError represents a structured error returned by the Aeitron API.
type APIError struct {
	StatusCode int    `json:"status_code"`
	ErrorCode  string `json:"error_code"`
	Message    string `json:"message"`
	RequestID  string `json:"request_id,omitempty"`
}

func (e *APIError) Error() string {
	if e.ErrorCode != "" {
		return fmt.Sprintf("aeitron: %s (status %d, code %s, request %s)", e.Message, e.StatusCode, e.ErrorCode, e.RequestID)
	}
	return fmt.Sprintf("aeitron: %s (status %d, request %s)", e.Message, e.StatusCode, e.RequestID)
}

// AuthenticationError occurs when an API key is missing, invalid, or expired (HTTP 401).
type AuthenticationError struct {
	APIError
}

// PermissionDeniedError occurs when the caller lacks required permissions (HTTP 403).
type PermissionDeniedError struct {
	APIError
}

// NotFoundError occurs when a requested resource does not exist (HTTP 404).
type NotFoundError struct {
	APIError
}

// ConflictError occurs on state conflict or duplicate entries (HTTP 409).
type ConflictError struct {
	APIError
}

// RateLimitError occurs when the rate limit is exceeded (HTTP 429).
type RateLimitError struct {
	APIError
	RetryAfter time.Duration
}

// ServerError occurs when the server encounters an internal error (HTTP 5xx).
type ServerError struct {
	APIError
}

// NetworkError occurs on connection failure, timeout, or DNS resolution failure.
type NetworkError struct {
	Err error
}

func (e *NetworkError) Error() string {
	return fmt.Sprintf("aeitron network error: %v", e.Err)
}

func (e *NetworkError) Unwrap() error {
	return e.Err
}
