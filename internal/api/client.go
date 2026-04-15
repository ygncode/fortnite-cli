package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Version is injected at build time via -ldflags. Defaults to "dev".
var Version = "dev"

// Client is a thin HTTP client for the Fortnite Ecosystem v1 Data API.
// It knows nothing about CLI flags, cobra, or stdout.
type Client struct {
	BaseURL    string        // e.g. "https://api.fortnite.com/ecosystem/v1"
	HTTP       *http.Client  // uses its Timeout; callers should set sensibly
	MaxRetries int           // default 3
	RetryBase  time.Duration // default 1s (initial backoff)
	NoRetry    bool          // if true, never retries
}

// DefaultClient builds a Client with production defaults.
func DefaultClient(timeout time.Duration) *Client {
	return &Client{
		BaseURL:    "https://api.fortnite.com/ecosystem/v1",
		HTTP:       &http.Client{Timeout: timeout},
		MaxRetries: 3,
		RetryBase:  time.Second,
	}
}

// APIError is returned for non-2xx responses. It is also returned when
// all retries are exhausted.
type APIError struct {
	Status     int
	Code       string // populated from ErrorResponse.errorCode when decoded
	Message    string // populated from ErrorResponse.errorMessage or raw body
	UUID       string // populated from ErrorResponse.uuid when present
	Body       string // raw body for non-JSON responses (429 is text/plain)
	RetryAfter string // Retry-After response header (seconds, or empty)
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("api error %d %s: %s", e.Status, e.Code, e.Message)
	}
	return fmt.Sprintf("api error %d: %s", e.Status, e.Message)
}

// Get performs a GET request against BaseURL+path with the given query params
// and decodes the JSON body into out. It handles retries for 429/5xx.
func (c *Client) Get(ctx context.Context, path string, query url.Values, out any) error {
	u, err := url.Parse(c.BaseURL + path)
	if err != nil {
		return fmt.Errorf("bad url: %w", err)
	}
	if query != nil {
		u.RawQuery = query.Encode()
	}

	maxAttempts := c.MaxRetries + 1
	if c.NoRetry {
		maxAttempts = 1
	}

	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			wait := c.RetryBase << (attempt - 1)
			// Honor server-provided Retry-After if larger.
			if apiErr, ok := lastErr.(*APIError); ok && apiErr.Status == http.StatusTooManyRequests {
				if ra := parseRetryAfter(apiErr.RetryAfter); ra > wait {
					wait = ra
				}
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wait):
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return fmt.Errorf("build request: %w", err)
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "fortnite-cli/"+Version)

		resp, err := c.HTTP.Do(req)
		if err != nil {
			// Network error: retry unless context expired.
			lastErr = fmt.Errorf("http: %w", err)
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
				return err
			}
			continue
		}

		body, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			lastErr = fmt.Errorf("read body: %w", readErr)
			continue
		}

		// Success.
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			if out == nil || len(body) == 0 {
				return nil
			}
			if err := json.Unmarshal(body, out); err != nil {
				return fmt.Errorf("decode response: %w", err)
			}
			return nil
		}

		// Build APIError.
		apiErr := &APIError{Status: resp.StatusCode, Body: string(body), RetryAfter: resp.Header.Get("Retry-After")}
		if ct := resp.Header.Get("Content-Type"); bytes.Contains([]byte(ct), []byte("json")) {
			var decoded ErrorResponse
			if err := json.Unmarshal(body, &decoded); err == nil {
				apiErr.Code = decoded.ErrorCode
				apiErr.Message = decoded.ErrorMessage
				apiErr.UUID = decoded.UUID
			}
		}
		if apiErr.Message == "" {
			apiErr.Message = string(body)
		}
		lastErr = apiErr

		// Retry decision.
		switch {
		case c.NoRetry:
			return apiErr
		case resp.StatusCode == http.StatusTooManyRequests:
			continue
		case resp.StatusCode >= 500 && resp.StatusCode < 600:
			continue
		default:
			return apiErr
		}
	}
	return lastErr
}

// parseRetryAfter parses a Retry-After header value (seconds only — we ignore the HTTP-date form).
func parseRetryAfter(v string) time.Duration {
	if v == "" {
		return 0
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		return 0
	}
	return time.Duration(n) * time.Second
}
