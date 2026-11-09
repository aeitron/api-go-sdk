package aeitron

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const SDKVersion = "1.0.0"

type transport struct {
	apiKey     string
	baseURL    string
	apiVersion string
	timeout    time.Duration
	maxRetries int
	httpClient *http.Client
	projectID  string
}

func newTransport(apiKey string, opts *clientOptions) *transport {
	httpClient := opts.httpClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: opts.timeout,
		}
	}

	return &transport{
		apiKey:     apiKey,
		baseURL:    strings.TrimRight(opts.baseURL, "/"),
		apiVersion: opts.apiVersion,
		timeout:    opts.timeout,
		maxRetries: opts.maxRetries,
		httpClient: httpClient,
		projectID:  opts.projectID,
	}
}

func (t *transport) do(ctx context.Context, method, path string, body any, out any) error {
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("aeitron: failed to serialize request body: %w", err)
		}
	}

	url := fmt.Sprintf("%s%s", t.baseURL, path)
	var resp *http.Response
	var lastErr error

	maxAttempts := 1 + t.maxRetries
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			backoff := t.calculateBackoff(attempt, resp)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}

		req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(bodyBytes))
		if err != nil {
			return &NetworkError{Err: err}
		}

		// Security and contract headers
		req.Header.Set("Authorization", "Bearer "+t.apiKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", fmt.Sprintf("aeitron-go-sdk/%s (Go 1.24)", SDKVersion))
		if t.apiVersion != "" {
			req.Header.Set("X-API-Version", t.apiVersion)
		}
		if t.projectID != "" {
			req.Header.Set("X-Project-ID", t.projectID)
		}

		resp, err = t.httpClient.Do(req)
		if err != nil {
			lastErr = &NetworkError{Err: err}
			continue
		}

		// Check response code
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			defer resp.Body.Close()
			if out != nil && resp.StatusCode != http.StatusNoContent {
				dec := json.NewDecoder(resp.Body)
				if err := dec.Decode(out); err != nil {
					return fmt.Errorf("aeitron: failed to decode response JSON: %w", err)
				}
			}
			return nil
		}

		// Handle HTTP errors
		err = t.parseError(resp)
		resp.Body.Close()

		if t.isRetryable(resp.StatusCode) && attempt < maxAttempts-1 {
			lastErr = err
			continue
		}

		return err
	}

	if lastErr != nil {
		return lastErr
	}
	return &ServerError{APIError: APIError{StatusCode: 500, Message: "request failed after retries"}}
}

func (t *transport) isRetryable(status int) bool {
	return status == http.StatusTooManyRequests ||
		status == http.StatusInternalServerError ||
		status == http.StatusBadGateway ||
		status == http.StatusServiceUnavailable ||
		status == http.StatusGatewayTimeout
}

func (t *transport) calculateBackoff(attempt int, resp *http.Response) time.Duration {
	if resp != nil && resp.StatusCode == http.StatusTooManyRequests {
		if val := resp.Header.Get("Retry-After"); val != "" {
			if secs, err := strconv.Atoi(val); err == nil && secs > 0 {
				return time.Duration(secs) * time.Second
			}
		}
	}

	base := 100 * time.Millisecond
	multiplier := 1 << (attempt - 1)
	backoff := base * time.Duration(multiplier)
	if backoff > 5*time.Second {
		backoff = 5 * time.Second
	}
	jitter := time.Duration(rand.Int63n(int64(backoff / 2)))
	return backoff + jitter
}

func (t *transport) parseError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	apiErr := APIError{
		StatusCode: resp.StatusCode,
		RequestID:  resp.Header.Get("X-Request-ID"),
		Message:    http.StatusText(resp.StatusCode),
	}

	var jsonErr struct {
		Error   string `json:"error"`
		Message string `json:"message"`
		Code    string `json:"code"`
	}
	if err := json.Unmarshal(body, &jsonErr); err == nil {
		if jsonErr.Message != "" {
			apiErr.Message = jsonErr.Message
		} else if jsonErr.Error != "" {
			apiErr.Message = jsonErr.Error
		}
		apiErr.ErrorCode = jsonErr.Code
	} else if len(body) > 0 {
		apiErr.Message = string(body)
	}

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return &AuthenticationError{APIError: apiErr}
	case http.StatusForbidden:
		return &PermissionDeniedError{APIError: apiErr}
	case http.StatusNotFound:
		return &NotFoundError{APIError: apiErr}
	case http.StatusConflict:
		return &ConflictError{APIError: apiErr}
	case http.StatusTooManyRequests:
		var retryAfter time.Duration
		if val := resp.Header.Get("Retry-After"); val != "" {
			if secs, err := strconv.Atoi(val); err == nil {
				retryAfter = time.Duration(secs) * time.Second
			}
		}
		return &RateLimitError{APIError: apiErr, RetryAfter: retryAfter}
	default:
		if resp.StatusCode >= 500 {
			return &ServerError{APIError: apiErr}
		}
		return &apiErr
	}
}
