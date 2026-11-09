package aeitron

import (
	"net/http"
	"time"
)

const (
	DefaultBaseURL    = "https://dashboard.aeitron.com"
	DefaultAPIVersion = "2026-11-01"
	DefaultTimeout    = 30 * time.Second
	DefaultMaxRetries = 3
)

// Option configures an Aeitron Client.
type Option func(*clientOptions)

type clientOptions struct {
	baseURL    string
	apiVersion string
	timeout    time.Duration
	maxRetries int
	httpClient *http.Client
	projectID  string
}

func defaultOptions() *clientOptions {
	return &clientOptions{
		baseURL:    DefaultBaseURL,
		apiVersion: DefaultAPIVersion,
		timeout:    DefaultTimeout,
		maxRetries: DefaultMaxRetries,
		httpClient: nil,
	}
}

// WithBaseURL overrides the default API endpoint.
func WithBaseURL(url string) Option {
	return func(o *clientOptions) {
		o.baseURL = url
	}
}

// WithAPIVersion specifies an explicit API contract version.
func WithAPIVersion(version string) Option {
	return func(o *clientOptions) {
		o.apiVersion = version
	}
}

// WithTimeout sets the maximum duration for a single HTTP request.
func WithTimeout(d time.Duration) Option {
	return func(o *clientOptions) {
		o.timeout = d
	}
}

// WithMaxRetries configures the retry limit for transient failures.
func WithMaxRetries(retries int) Option {
	return func(o *clientOptions) {
		o.maxRetries = retries
	}
}

// WithHTTPClient provides a custom *http.Client.
func WithHTTPClient(client *http.Client) Option {
	return func(o *clientOptions) {
		o.httpClient = client
	}
}

// WithProjectID attaches a default tenant project scope.
func WithProjectID(projectID string) Option {
	return func(o *clientOptions) {
		o.projectID = projectID
	}
}
