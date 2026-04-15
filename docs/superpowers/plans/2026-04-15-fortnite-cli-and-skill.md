# Fortnite CLI + Agent Skill Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Go CLI (`fortnite`) that wraps the Fortnite Ecosystem v1 Data API and ship it together with an agentskills.io-compatible skill (`fortnite-islands`) from `github.com/ygncode/fortnite-cli`.

**Architecture:** Clean layer separation. `internal/api` is a pure HTTP client (no cobra, no stdout). `internal/slim`, `internal/timerange`, `internal/output` are pure functions. `internal/commands` is the cobra glue that wires them together. Commands are thin — one file per subcommand, target <150 lines. The skill lives in `skill/fortnite-islands/` and is versioned in lockstep with the binary.

**Tech Stack:** Go 1.23+, cobra/pflag for CLI, stretchr/testify for tests, net/http stdlib for HTTP, GoReleaser for distribution. No OpenAPI generator, no viper, no logger.

**Design reference:** `docs/superpowers/specs/2026-04-15-fortnite-cli-and-skill-design.md`

---

## Task 1: Project bootstrap

**Files:**
- Create: `go.mod`
- Create: `.gitignore`
- Create: `LICENSE` (MIT)
- Create: `README.md` (minimal stub)
- Create: `cmd/fortnite/main.go` (hello-world placeholder)

- [ ] **Step 1: Initialize Go module**

Run from repo root:
```bash
go mod init github.com/ygncode/fortnite-cli
```

Expected: creates `go.mod` with `module github.com/ygncode/fortnite-cli` and `go 1.23` (or whatever the local Go version is; 1.23+ required).

- [ ] **Step 2: Create `.gitignore`**

```
# Binaries
/fortnite
/dist/
*.test
*.out

# IDE
.idea/
.vscode/
*.swp

# OS
.DS_Store
```

- [ ] **Step 3: Create `LICENSE` (MIT)**

Standard MIT license text. Year `2026`, copyright holder `ygncode`. Use the SPDX MIT template verbatim.

- [ ] **Step 4: Create `README.md` stub**

```markdown
# fortnite-cli

A Go CLI for the official Fortnite Ecosystem v1 Data API, plus a companion agent skill (`fortnite-islands`).

## Status
Under active development. See `docs/superpowers/specs/` for the design spec and `docs/superpowers/plans/` for the implementation plan.
```

- [ ] **Step 5: Create `cmd/fortnite/main.go` (placeholder)**

```go
package main

import "fmt"

func main() {
	fmt.Println("fortnite: not yet implemented")
}
```

- [ ] **Step 6: Verify build**

Run:
```bash
go build ./...
```

Expected: exits 0, produces no output. A `fortnite` binary appears in repo root.

- [ ] **Step 7: Commit**

```bash
git add go.mod .gitignore LICENSE README.md cmd/fortnite/main.go
git commit -m "chore: bootstrap go module and repo layout"
```

---

## Task 2: API response types

**Files:**
- Create: `internal/api/types.go`
- Create: `internal/api/types_test.go`
- Create: `testdata/raw/islands_list.json`
- Create: `testdata/raw/island_get.json`
- Create: `testdata/raw/metrics_bundled.json`
- Create: `testdata/raw/metric_peak_ccu.json`
- Create: `testdata/raw/retention.json`

- [ ] **Step 1: Create testdata — `testdata/raw/islands_list.json`**

This is a captured `GET /islands?size=2` response, pruned to two islands:

```json
{
  "links": {
    "next": "/ecosystem/v1/islands?after=ODQ5NS0zMTc3LTY5NDU%3D&size=2",
    "prev": null
  },
  "meta": {
    "count": 2,
    "page": {
      "prevCursor": null,
      "nextCursor": "ODQ5NS0zMTc3LTY5NDU="
    }
  },
  "data": [
    {
      "code": "7022-1566-6527",
      "creatorCode": "daryadidi",
      "title": "BRAINROT RACE",
      "createdIn": "UEFN",
      "tags": ["race", "tycoon", "just for fun", "action"],
      "meta": {"page": {"cursor": "NzAyMi0xNTY2LTY1Mjc="}}
    },
    {
      "code": "8495-3177-6945",
      "creatorCode": "maphits",
      "title": "ULTIMATE FIRE ZONE",
      "createdIn": "UEFN",
      "tags": ["one shot", "gun game", "free for all", "casual"],
      "meta": {"page": {"cursor": "ODQ5NS0zMTc3LTY5NDU="}}
    }
  ]
}
```

- [ ] **Step 2: Create testdata — `testdata/raw/island_get.json`**

```json
{
  "code": "7022-1566-6527",
  "creatorCode": "daryadidi",
  "title": "BRAINROT RACE",
  "createdIn": "UEFN",
  "tags": ["race", "tycoon", "just for fun", "action"]
}
```

- [ ] **Step 3: Create testdata — `testdata/raw/metrics_bundled.json`**

A bundled `/metrics` response with one non-null metric and the rest null:

```json
{
  "averageMinutesPerPlayer": [
    {"value": null, "timestamp": "2026-04-14T00:00:00.000Z"},
    {"value": null, "timestamp": "2026-04-15T00:00:00.000Z"}
  ],
  "peakCCU": [
    {"value": 120, "timestamp": "2026-04-14T00:00:00.000Z"},
    {"value": 95, "timestamp": "2026-04-15T00:00:00.000Z"}
  ],
  "favorites": [
    {"value": null, "timestamp": "2026-04-14T00:00:00.000Z"},
    {"value": null, "timestamp": "2026-04-15T00:00:00.000Z"}
  ],
  "minutesPlayed": [
    {"value": null, "timestamp": "2026-04-14T00:00:00.000Z"},
    {"value": null, "timestamp": "2026-04-15T00:00:00.000Z"}
  ],
  "recommendations": [
    {"value": null, "timestamp": "2026-04-14T00:00:00.000Z"},
    {"value": null, "timestamp": "2026-04-15T00:00:00.000Z"}
  ],
  "plays": [
    {"value": null, "timestamp": "2026-04-14T00:00:00.000Z"},
    {"value": null, "timestamp": "2026-04-15T00:00:00.000Z"}
  ],
  "uniquePlayers": [
    {"value": null, "timestamp": "2026-04-14T00:00:00.000Z"},
    {"value": null, "timestamp": "2026-04-15T00:00:00.000Z"}
  ],
  "retention": [
    {"d1": null, "d7": null, "timestamp": "2026-04-14T00:00:00.000Z"},
    {"d1": null, "d7": null, "timestamp": "2026-04-15T00:00:00.000Z"}
  ]
}
```

- [ ] **Step 4: Create testdata — `testdata/raw/metric_peak_ccu.json`**

Single-metric response shape (endpoint `.../metrics/{interval}/peak-ccu`):

```json
{
  "intervals": [
    {"value": 120, "timestamp": "2026-04-14T07:00:00.000Z"},
    {"value": 95, "timestamp": "2026-04-14T08:00:00.000Z"},
    {"value": null, "timestamp": "2026-04-14T09:00:00.000Z"}
  ]
}
```

- [ ] **Step 5: Create testdata — `testdata/raw/retention.json`**

```json
{
  "intervals": [
    {"d1": 80, "d7": 45, "timestamp": "2026-04-14T00:00:00.000Z"},
    {"d1": null, "d7": null, "timestamp": "2026-04-15T00:00:00.000Z"}
  ]
}
```

- [ ] **Step 6: Write failing test `internal/api/types_test.go`**

```go
package api

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func readTestdata(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "raw", name))
	require.NoError(t, err)
	return data
}

func TestIslandListResponseRoundtrip(t *testing.T) {
	var resp IslandListResponse
	require.NoError(t, json.Unmarshal(readTestdata(t, "islands_list.json"), &resp))
	require.Len(t, resp.Data, 2)
	require.Equal(t, "7022-1566-6527", resp.Data[0].Code)
	require.Equal(t, "daryadidi", resp.Data[0].CreatorCode)
	require.Equal(t, "ODQ5NS0zMTc3LTY5NDU=", resp.Meta.Page.NextCursor)
	require.Contains(t, resp.Data[0].Tags, "race")
}

func TestIslandMetadataRoundtrip(t *testing.T) {
	var meta IslandMetadata
	require.NoError(t, json.Unmarshal(readTestdata(t, "island_get.json"), &meta))
	require.Equal(t, "BRAINROT RACE", meta.Title)
	require.Equal(t, "UEFN", meta.CreatedIn)
}

func TestBundledMetricsRoundtrip(t *testing.T) {
	var m BundledMetrics
	require.NoError(t, json.Unmarshal(readTestdata(t, "metrics_bundled.json"), &m))
	require.Len(t, m.PeakCCU, 2)
	require.NotNil(t, m.PeakCCU[0].Value)
	require.Equal(t, float64(120), *m.PeakCCU[0].Value)
	require.Nil(t, m.PeakCCU[0].Value == nil) // nil-check sanity
}

func TestMetricResponseRoundtrip(t *testing.T) {
	var m MetricResponse
	require.NoError(t, json.Unmarshal(readTestdata(t, "metric_peak_ccu.json"), &m))
	require.Len(t, m.Intervals, 3)
	require.NotNil(t, m.Intervals[0].Value)
	require.Nil(t, m.Intervals[2].Value)
}

func TestRetentionResponseRoundtrip(t *testing.T) {
	var r RetentionResponse
	require.NoError(t, json.Unmarshal(readTestdata(t, "retention.json"), &r))
	require.Len(t, r.Intervals, 2)
	require.NotNil(t, r.Intervals[0].D1)
	require.Equal(t, float64(80), *r.Intervals[0].D1)
	require.Nil(t, r.Intervals[1].D1)
}
```

- [ ] **Step 7: Run test to verify it fails**

```bash
go test ./internal/api/...
```

Expected: FAIL with "undefined: IslandListResponse" (and other undefined types).

- [ ] **Step 8: Write `internal/api/types.go`**

```go
// Package api defines request and response types for the Fortnite Ecosystem v1 Data API.
package api

// IslandMetadata represents an island's metadata (summary form).
type IslandMetadata struct {
	Code        string   `json:"code"`
	CreatorCode string   `json:"creatorCode,omitempty"`
	DisplayName string   `json:"displayName,omitempty"`
	Title       string   `json:"title"`
	Category    string   `json:"category,omitempty"`
	CreatedIn   string   `json:"createdIn,omitempty"`
	Tags        []string `json:"tags"`
}

// IslandListItem is an island in a list response, including per-item pagination cursor.
type IslandListItem struct {
	IslandMetadata
	Meta struct {
		Page struct {
			Cursor string `json:"cursor"`
		} `json:"page"`
	} `json:"meta"`
}

// PaginationLinks carries path-style next/prev links.
type PaginationLinks struct {
	Prev *string `json:"prev"`
	Next *string `json:"next"`
}

// PaginationMeta carries count + cursor metadata.
type PaginationMeta struct {
	Count int `json:"count"`
	Page  struct {
		PrevCursor string `json:"prevCursor"`
		NextCursor string `json:"nextCursor"`
	} `json:"page"`
}

// IslandListResponse is the envelope returned by GET /islands.
type IslandListResponse struct {
	Links PaginationLinks  `json:"links"`
	Meta  PaginationMeta   `json:"meta"`
	Data  []IslandListItem `json:"data"`
}

// MetricValue is a single (timestamp, value) bucket where value may be null.
type MetricValue struct {
	Value     *float64 `json:"value"`
	Timestamp string   `json:"timestamp"`
}

// BundledMetrics is the response from GET /islands/{code}/metrics (and the filterable variant).
type BundledMetrics struct {
	AverageMinutesPerPlayer []MetricValue    `json:"averageMinutesPerPlayer,omitempty"`
	PeakCCU                 []MetricValue    `json:"peakCCU,omitempty"`
	Favorites               []MetricValue    `json:"favorites,omitempty"`
	MinutesPlayed           []MetricValue    `json:"minutesPlayed,omitempty"`
	Recommendations         []MetricValue    `json:"recommendations,omitempty"`
	Plays                   []MetricValue    `json:"plays,omitempty"`
	UniquePlayers           []MetricValue    `json:"uniquePlayers,omitempty"`
	Retention               []RetentionValue `json:"retention,omitempty"`
}

// MetricResponse is the response from a single-metric endpoint.
type MetricResponse struct {
	Intervals []MetricValue `json:"intervals"`
}

// RetentionValue is a d1/d7 retention bucket.
type RetentionValue struct {
	D1        *float64 `json:"d1"`
	D7        *float64 `json:"d7"`
	Timestamp string   `json:"timestamp"`
}

// RetentionResponse is the response from .../metrics/day/retention.
type RetentionResponse struct {
	Intervals []RetentionValue `json:"intervals"`
}

// ErrorResponse is the API's error body (4xx).
type ErrorResponse struct {
	ErrorCode    string `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
	UUID         string `json:"uuid"`
}
```

- [ ] **Step 9: Run test to verify pass**

```bash
go test ./internal/api/...
```

Expected: PASS (5 tests).

- [ ] **Step 10: Commit**

```bash
git add internal/api/types.go internal/api/types_test.go testdata/raw
git commit -m "feat(api): add response types and testdata roundtrip tests"
```

---

## Task 3: API client core (Do, base URL, retry, error decoding)

**Files:**
- Create: `internal/api/client.go`
- Create: `internal/api/client_test.go`

- [ ] **Step 1: Write failing test `internal/api/client_test.go`**

```go
package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func newTestClient(baseURL string) *Client {
	return &Client{
		BaseURL:    baseURL,
		HTTP:       &http.Client{Timeout: 5 * time.Second},
		MaxRetries: 3,
		RetryBase:  10 * time.Millisecond, // fast for tests
	}
}

func TestClientGetSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/islands", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[],"links":{"prev":null,"next":null},"meta":{"count":0,"page":{"prevCursor":"","nextCursor":""}}}`))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	var resp IslandListResponse
	require.NoError(t, c.Get(context.Background(), "/islands", nil, &resp))
	require.Equal(t, 0, resp.Meta.Count)
}

func TestClientRetries429(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n < 3 {
			w.Header().Set("Retry-After", "0")
			http.Error(w, "slow down", http.StatusTooManyRequests)
			return
		}
		_, _ = w.Write([]byte(`{"data":[],"links":{"prev":null,"next":null},"meta":{"count":0,"page":{"prevCursor":"","nextCursor":""}}}`))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	var resp IslandListResponse
	require.NoError(t, c.Get(context.Background(), "/islands", nil, &resp))
	require.Equal(t, int32(3), atomic.LoadInt32(&calls))
}

func TestClientNoRetryFlag(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		http.Error(w, "slow down", http.StatusTooManyRequests)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	c.NoRetry = true
	var resp IslandListResponse
	err := c.Get(context.Background(), "/islands", nil, &resp)
	require.Error(t, err)
	require.Equal(t, int32(1), atomic.LoadInt32(&calls))

	var apiErr *APIError
	require.True(t, errors.As(err, &apiErr))
	require.Equal(t, http.StatusTooManyRequests, apiErr.Status)
}

func TestClientDecodes4xxError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errorCode":"errors.com.epicgames.not_found","errorMessage":"Island not found","uuid":"abc-123"}`))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	var resp IslandMetadata
	err := c.Get(context.Background(), "/islands/bad", nil, &resp)
	require.Error(t, err)

	var apiErr *APIError
	require.True(t, errors.As(err, &apiErr))
	require.Equal(t, http.StatusNotFound, apiErr.Status)
	require.Equal(t, "errors.com.epicgames.not_found", apiErr.Code)
	require.Equal(t, "Island not found", apiErr.Message)
	require.Equal(t, "abc-123", apiErr.UUID)
}

func TestClientQueryParams(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "peakCCU", r.URL.Query().Get("metrics"))
		require.Equal(t, []string{"peakCCU", "plays"}, r.URL.Query()["metrics"])
		require.True(t, strings.HasPrefix(r.Header.Get("User-Agent"), "fortnite-cli/"))
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	var resp BundledMetrics
	require.NoError(t, c.Get(context.Background(), "/test", map[string][]string{
		"metrics": {"peakCCU", "plays"},
	}, &resp))
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/api/ -run 'TestClient'
```

Expected: FAIL with "undefined: Client" / "undefined: APIError".

- [ ] **Step 3: Write `internal/api/client.go`**

```go
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
	Status  int
	Code    string // populated from ErrorResponse.errorCode when decoded
	Message string // populated from ErrorResponse.errorMessage or raw body
	UUID    string // populated from ErrorResponse.uuid when present
	Body    string // raw body for non-JSON responses (429 is text/plain)
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
				if ra := parseRetryAfter(apiErr.Body); ra > wait {
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
		apiErr := &APIError{Status: resp.StatusCode, Body: string(body)}
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
```

- [ ] **Step 4: Run test to verify pass**

```bash
go test ./internal/api/ -run 'TestClient'
```

Expected: PASS (5 tests).

- [ ] **Step 5: Commit**

```bash
git add internal/api/client.go internal/api/client_test.go
git commit -m "feat(api): add http client with retry and error decoding"
```

---

## Task 4: API client island methods

**Files:**
- Create: `internal/api/islands.go`
- Create: `internal/api/islands_test.go`

- [ ] **Step 1: Write failing test `internal/api/islands_test.go`**

```go
package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func newServerFromTestdata(t *testing.T, routes map[string]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Method + " " + r.URL.Path
		body, ok := routes[key]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
}

func TestListIslands(t *testing.T) {
	body := string(readTestdata(t, "islands_list.json"))
	srv := newServerFromTestdata(t, map[string]string{
		"GET /islands": body,
	})
	defer srv.Close()

	c := newTestClient(srv.URL)
	resp, err := c.ListIslands(context.Background(), ListIslandsParams{Size: 2})
	require.NoError(t, err)
	require.Len(t, resp.Data, 2)
	require.Equal(t, "7022-1566-6527", resp.Data[0].Code)
}

func TestGetIsland(t *testing.T) {
	body := string(readTestdata(t, "island_get.json"))
	srv := newServerFromTestdata(t, map[string]string{
		"GET /islands/7022-1566-6527": body,
	})
	defer srv.Close()

	c := newTestClient(srv.URL)
	meta, err := c.GetIsland(context.Background(), "7022-1566-6527")
	require.NoError(t, err)
	require.Equal(t, "BRAINROT RACE", meta.Title)
}

func TestGetBundledMetrics(t *testing.T) {
	body := string(readTestdata(t, "metrics_bundled.json"))
	srv := newServerFromTestdata(t, map[string]string{
		"GET /islands/7022-1566-6527/metrics": body,
	})
	defer srv.Close()

	c := newTestClient(srv.URL)
	m, err := c.GetBundledMetrics(context.Background(), "7022-1566-6527", MetricsParams{})
	require.NoError(t, err)
	require.Len(t, m.PeakCCU, 2)
}

func TestGetFilterableMetrics(t *testing.T) {
	body := string(readTestdata(t, "metrics_bundled.json"))
	srv := newServerFromTestdata(t, map[string]string{
		"GET /islands/7022-1566-6527/metrics/day": body,
	})
	defer srv.Close()

	c := newTestClient(srv.URL)
	m, err := c.GetFilterableMetrics(context.Background(), "7022-1566-6527", "day", MetricsParams{Metrics: []string{"peakCCU"}})
	require.NoError(t, err)
	require.Len(t, m.PeakCCU, 2)
}

func TestGetIslandMetric(t *testing.T) {
	body := string(readTestdata(t, "metric_peak_ccu.json"))
	srv := newServerFromTestdata(t, map[string]string{
		"GET /islands/7022-1566-6527/metrics/hour/peak-ccu": body,
	})
	defer srv.Close()

	c := newTestClient(srv.URL)
	m, err := c.GetIslandMetric(context.Background(), "7022-1566-6527", "hour", "peak-ccu", MetricsParams{})
	require.NoError(t, err)
	require.Len(t, m.Intervals, 3)
}

func TestGetRetention(t *testing.T) {
	body := string(readTestdata(t, "retention.json"))
	srv := newServerFromTestdata(t, map[string]string{
		"GET /islands/7022-1566-6527/metrics/day/retention": body,
	})
	defer srv.Close()

	c := newTestClient(srv.URL)
	r, err := c.GetRetention(context.Background(), "7022-1566-6527", MetricsParams{})
	require.NoError(t, err)
	require.Len(t, r.Intervals, 2)
	require.NotNil(t, r.Intervals[0].D1)
}

func TestListIslandsEncodesCursorAndSize(t *testing.T) {
	var captured url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = r.URL.Query()
		_, _ = w.Write([]byte(`{"data":[],"links":{"prev":null,"next":null},"meta":{"count":0,"page":{"prevCursor":"","nextCursor":""}}}`))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	_, err := c.ListIslands(context.Background(), ListIslandsParams{Size: 50, After: "abc"})
	require.NoError(t, err)
	require.Equal(t, "50", captured.Get("size"))
	require.Equal(t, "abc", captured.Get("after"))
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/api/ -run 'TestList|TestGet'
```

Expected: FAIL with "undefined: ListIslandsParams" etc.

- [ ] **Step 3: Write `internal/api/islands.go`**

```go
package api

import (
	"context"
	"net/url"
	"strconv"
	"time"
)

// ListIslandsParams is the query for GET /islands.
type ListIslandsParams struct {
	Size   int
	After  string
	Before string
}

// MetricsParams is the shared query struct for metric endpoints.
// Not every field applies to every endpoint; unused fields are zero and ignored.
type MetricsParams struct {
	From    time.Time
	To      time.Time
	Metrics []string // only used by GetFilterableMetrics
}

func (p MetricsParams) encode() url.Values {
	v := url.Values{}
	if !p.From.IsZero() {
		v.Set("from", p.From.UTC().Format(time.RFC3339Nano))
	}
	if !p.To.IsZero() {
		v.Set("to", p.To.UTC().Format(time.RFC3339Nano))
	}
	for _, m := range p.Metrics {
		v.Add("metrics", m)
	}
	return v
}

// ListIslands calls GET /islands.
func (c *Client) ListIslands(ctx context.Context, p ListIslandsParams) (*IslandListResponse, error) {
	q := url.Values{}
	if p.Size > 0 {
		q.Set("size", strconv.Itoa(p.Size))
	}
	if p.After != "" {
		q.Set("after", p.After)
	}
	if p.Before != "" {
		q.Set("before", p.Before)
	}
	var out IslandListResponse
	if err := c.Get(ctx, "/islands", q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetIsland calls GET /islands/{code}.
func (c *Client) GetIsland(ctx context.Context, code string) (*IslandMetadata, error) {
	var out IslandMetadata
	if err := c.Get(ctx, "/islands/"+code, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetBundledMetrics calls GET /islands/{code}/metrics (day interval, all metrics).
func (c *Client) GetBundledMetrics(ctx context.Context, code string, p MetricsParams) (*BundledMetrics, error) {
	var out BundledMetrics
	if err := c.Get(ctx, "/islands/"+code+"/metrics", p.encode(), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetFilterableMetrics calls GET /islands/{code}/metrics/{interval}.
func (c *Client) GetFilterableMetrics(ctx context.Context, code, interval string, p MetricsParams) (*BundledMetrics, error) {
	var out BundledMetrics
	if err := c.Get(ctx, "/islands/"+code+"/metrics/"+interval, p.encode(), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetIslandMetric calls one of the eight single-metric endpoints. metric must be one of:
// "peak-ccu", "favorites", "minutes-played", "average-minutes-per-player",
// "recommendations", "unique-players", "plays" — NOT "retention" (use GetRetention).
func (c *Client) GetIslandMetric(ctx context.Context, code, interval, metric string, p MetricsParams) (*MetricResponse, error) {
	var out MetricResponse
	path := "/islands/" + code + "/metrics/" + interval + "/" + metric
	if err := c.Get(ctx, path, p.encode(), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetRetention calls GET /islands/{code}/metrics/day/retention. The API only
// supports day interval for retention; callers should not pass any other.
func (c *Client) GetRetention(ctx context.Context, code string, p MetricsParams) (*RetentionResponse, error) {
	var out RetentionResponse
	if err := c.Get(ctx, "/islands/"+code+"/metrics/day/retention", p.encode(), &out); err != nil {
		return nil, err
	}
	return &out, nil
}
```

- [ ] **Step 4: Run test to verify pass**

```bash
go test ./internal/api/...
```

Expected: PASS (all api tests).

- [ ] **Step 5: Commit**

```bash
git add internal/api/islands.go internal/api/islands_test.go
git commit -m "feat(api): add typed island methods for all 12 endpoints"
```

---

## Task 5: Time-range parsing

**Files:**
- Create: `internal/timerange/parse.go`
- Create: `internal/timerange/parse_test.go`

- [ ] **Step 1: Write failing test `internal/timerange/parse_test.go`**

```go
package timerange

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseLast(t *testing.T) {
	now := time.Date(2026, 4, 15, 12, 0, 0, 0, time.UTC)
	r, warn, err := Parse(Input{Last: "24h", Interval: "hour"}, now)
	require.NoError(t, err)
	require.Empty(t, warn)
	require.Equal(t, now.Add(-24*time.Hour), r.From)
	require.Equal(t, now, r.To)
}

func TestParseFromTo(t *testing.T) {
	now := time.Date(2026, 4, 15, 12, 0, 0, 0, time.UTC)
	r, _, err := Parse(Input{From: "2026-04-14T00:00:00Z", To: "2026-04-15T00:00:00Z", Interval: "day"}, now)
	require.NoError(t, err)
	require.Equal(t, "2026-04-14T00:00:00Z", r.From.Format(time.RFC3339))
	require.Equal(t, "2026-04-15T00:00:00Z", r.To.Format(time.RFC3339))
}

func TestParseLastWinsOverFromTo(t *testing.T) {
	now := time.Date(2026, 4, 15, 12, 0, 0, 0, time.UTC)
	r, warn, err := Parse(Input{Last: "1h", From: "2026-04-14T00:00:00Z", To: "2026-04-15T00:00:00Z", Interval: "hour"}, now)
	require.NoError(t, err)
	require.NotEmpty(t, warn)
	require.Equal(t, now.Add(-1*time.Hour), r.From)
}

func TestParseRejectsFromAfterTo(t *testing.T) {
	now := time.Date(2026, 4, 15, 12, 0, 0, 0, time.UTC)
	_, _, err := Parse(Input{From: "2026-04-15T00:00:00Z", To: "2026-04-14T00:00:00Z", Interval: "day"}, now)
	require.Error(t, err)
}

func TestParseRejectsOver7Days(t *testing.T) {
	now := time.Date(2026, 4, 15, 12, 0, 0, 0, time.UTC)
	_, _, err := Parse(Input{Last: "8d", Interval: "day"}, now)
	require.Error(t, err)
	require.Contains(t, err.Error(), "7-day")
}

func TestParseRejectsOver7DaysFromExplicit(t *testing.T) {
	now := time.Date(2026, 4, 15, 12, 0, 0, 0, time.UTC)
	_, _, err := Parse(Input{From: "2026-04-01T00:00:00Z", To: "2026-04-15T00:00:00Z", Interval: "day"}, now)
	require.Error(t, err)
	require.Contains(t, err.Error(), "7-day")
}

func TestParseClampsToFuture(t *testing.T) {
	now := time.Date(2026, 4, 15, 12, 0, 0, 0, time.UTC)
	r, warn, err := Parse(Input{From: "2026-04-15T00:00:00Z", To: "2026-04-16T00:00:00Z", Interval: "day"}, now)
	require.NoError(t, err)
	require.NotEmpty(t, warn)
	require.Equal(t, now, r.To)
}

func TestParseDefaults(t *testing.T) {
	now := time.Date(2026, 4, 15, 12, 0, 0, 0, time.UTC)

	rDay, _, err := Parse(Input{Interval: "day"}, now)
	require.NoError(t, err)
	require.Equal(t, now.Add(-24*time.Hour), rDay.From)

	rHour, _, err := Parse(Input{Interval: "hour"}, now)
	require.NoError(t, err)
	require.Equal(t, now.Add(-24*time.Hour), rHour.From)

	rMin, _, err := Parse(Input{Interval: "minute"}, now)
	require.NoError(t, err)
	require.Equal(t, now.Add(-60*time.Minute), rMin.From)
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/timerange/
```

Expected: FAIL with "undefined: Input" / "undefined: Parse".

- [ ] **Step 3: Write `internal/timerange/parse.go`**

```go
// Package timerange parses --last / --from / --to flags into a validated time range.
package timerange

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// maxHistory is the Fortnite API's historical window cap.
const maxHistory = 7 * 24 * time.Hour

// Input captures the raw flag values from a command.
type Input struct {
	Last     string // e.g. "24h", "7d", "60m"
	From     string // RFC3339
	To       string // RFC3339
	Interval string // "day" | "hour" | "minute"
}

// Range is the parsed, validated time range.
type Range struct {
	From time.Time
	To   time.Time
}

// Parse converts Input into a Range. It returns (range, warnings, error).
// Warnings are non-empty strings intended for stderr (e.g. "both --last and --from provided").
// The caller supplies `now` so tests can pin time.
func Parse(in Input, now time.Time) (Range, string, error) {
	var warn string

	// --last wins over --from/--to.
	if in.Last != "" {
		dur, err := parseDuration(in.Last)
		if err != nil {
			return Range{}, "", fmt.Errorf("invalid --last: %w", err)
		}
		if dur > maxHistory {
			return Range{}, "", errors.New("invalid --last: exceeds API 7-day historical limit")
		}
		if in.From != "" || in.To != "" {
			warn = "--last takes precedence; --from/--to ignored"
		}
		return Range{From: now.Add(-dur), To: now}, warn, nil
	}

	// Explicit --from/--to.
	if in.From != "" || in.To != "" {
		if in.From == "" || in.To == "" {
			return Range{}, "", errors.New("--from and --to must be provided together (or use --last)")
		}
		from, err := time.Parse(time.RFC3339, in.From)
		if err != nil {
			return Range{}, "", fmt.Errorf("invalid --from: %w", err)
		}
		to, err := time.Parse(time.RFC3339, in.To)
		if err != nil {
			return Range{}, "", fmt.Errorf("invalid --to: %w", err)
		}
		if !from.Before(to) {
			return Range{}, "", errors.New("--from must be before --to")
		}
		if to.After(now) {
			warn = "--to is in the future; clamping to now"
			to = now
		}
		if now.Sub(from) > maxHistory {
			return Range{}, "", errors.New("--from exceeds API 7-day historical limit")
		}
		return Range{From: from, To: to}, warn, nil
	}

	// Defaults per interval.
	switch in.Interval {
	case "day", "":
		return Range{From: now.Add(-24 * time.Hour), To: now}, "", nil
	case "hour":
		return Range{From: now.Add(-24 * time.Hour), To: now}, "", nil
	case "minute":
		return Range{From: now.Add(-60 * time.Minute), To: now}, "", nil
	default:
		return Range{}, "", fmt.Errorf("unknown interval %q", in.Interval)
	}
}

// parseDuration accepts "24h", "60m", "7d", "30s" — i.e. Go duration syntax plus "d" for days.
func parseDuration(s string) (time.Duration, error) {
	if strings.HasSuffix(s, "d") {
		var days int
		if _, err := fmt.Sscanf(s, "%dd", &days); err != nil || days < 0 {
			return 0, fmt.Errorf("bad duration %q", s)
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, err
	}
	if d < 0 {
		return 0, errors.New("duration must be positive")
	}
	return d, nil
}
```

- [ ] **Step 4: Run test to verify pass**

```bash
go test ./internal/timerange/
```

Expected: PASS (8 tests).

- [ ] **Step 5: Commit**

```bash
git add internal/timerange/
git commit -m "feat(timerange): parse --last/--from/--to with 7-day validation"
```

---

## Task 6: Slim JSON — list envelope unwrap

**Files:**
- Create: `internal/slim/slim.go`
- Create: `internal/slim/slim_test.go`
- Create: `testdata/slim/islands_list.json`

- [ ] **Step 1: Create expected golden `testdata/slim/islands_list.json`**

```json
{
  "_pagination": {
    "next": "/ecosystem/v1/islands?after=ODQ5NS0zMTc3LTY5NDU%3D&size=2"
  },
  "data": [
    {
      "code": "7022-1566-6527",
      "creatorCode": "daryadidi",
      "title": "BRAINROT RACE",
      "createdIn": "UEFN",
      "tags": ["race", "tycoon", "just for fun", "action"],
      "cursor": "NzAyMi0xNTY2LTY1Mjc="
    },
    {
      "code": "8495-3177-6945",
      "creatorCode": "maphits",
      "title": "ULTIMATE FIRE ZONE",
      "createdIn": "UEFN",
      "tags": ["one shot", "gun game", "free for all", "casual"],
      "cursor": "ODQ5NS0zMTc3LTY5NDU="
    }
  ]
}
```

- [ ] **Step 2: Write failing test `internal/slim/slim_test.go`**

```go
package slim

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func readFile(t *testing.T, rel string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", rel))
	require.NoError(t, err)
	return data
}

func assertJSONEqual(t *testing.T, expected, got []byte) {
	t.Helper()
	var e, g any
	require.NoError(t, json.Unmarshal(expected, &e))
	require.NoError(t, json.Unmarshal(got, &g))
	eb, _ := json.MarshalIndent(e, "", "  ")
	gb, _ := json.MarshalIndent(g, "", "  ")
	require.Equal(t, string(eb), string(gb))
}

func TestSlimIslandList(t *testing.T) {
	raw := readFile(t, "testdata/raw/islands_list.json")
	expected := readFile(t, "testdata/slim/islands_list.json")

	got, err := IslandList(raw)
	require.NoError(t, err)
	assertJSONEqual(t, expected, got)
}
```

- [ ] **Step 3: Run test to verify it fails**

```bash
go test ./internal/slim/
```

Expected: FAIL with "undefined: IslandList".

- [ ] **Step 4: Write `internal/slim/slim.go`**

```go
// Package slim transforms raw Fortnite Ecosystem API responses into agent-friendly
// compact JSON. All functions are pure — no I/O, no HTTP.
package slim

import (
	"encoding/json"
	"fmt"
)

// IslandList transforms a GET /islands response into slim form:
// - Envelope {data, links, meta} is flattened into {_pagination?, data}.
// - Each island's nested meta.page.cursor is lifted to a top-level "cursor" field.
// - Null optional fields (displayName, creatorCode, category) are omitted by the source marshaler.
// - `_pagination` is omitted entirely when both links are null.
func IslandList(raw []byte) ([]byte, error) {
	var in struct {
		Links struct {
			Prev *string `json:"prev"`
			Next *string `json:"next"`
		} `json:"links"`
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return nil, fmt.Errorf("slim island list: decode: %w", err)
	}

	out := map[string]any{}

	if in.Links.Prev != nil || in.Links.Next != nil {
		pag := map[string]any{}
		if in.Links.Prev != nil {
			pag["prev"] = *in.Links.Prev
		}
		if in.Links.Next != nil {
			pag["next"] = *in.Links.Next
		}
		out["_pagination"] = pag
	}

	slimmed := make([]map[string]any, 0, len(in.Data))
	for _, item := range in.Data {
		o := map[string]any{}
		for k, v := range item {
			if k == "meta" {
				continue
			}
			if s, ok := v.(string); ok && s == "" {
				continue // omit empty optional strings
			}
			o[k] = v
		}
		if meta, ok := item["meta"].(map[string]any); ok {
			if page, ok := meta["page"].(map[string]any); ok {
				if c, ok := page["cursor"].(string); ok && c != "" {
					o["cursor"] = c
				}
			}
		}
		slimmed = append(slimmed, o)
	}
	out["data"] = slimmed

	return json.Marshal(out)
}
```

- [ ] **Step 5: Run test to verify pass**

```bash
go test ./internal/slim/
```

Expected: PASS (1 test).

- [ ] **Step 6: Commit**

```bash
git add internal/slim/ testdata/slim/
git commit -m "feat(slim): unwrap island list envelope"
```

---

## Task 7: Slim JSON — bundled metrics (timestamp rounding + null collapse)

**Files:**
- Modify: `internal/slim/slim.go`
- Modify: `internal/slim/slim_test.go`
- Create: `testdata/slim/metrics_bundled.json`

- [ ] **Step 1: Create expected golden `testdata/slim/metrics_bundled.json`**

```json
{
  "averageMinutesPerPlayer": null,
  "peakCCU": [
    {"value": 120, "timestamp": "2026-04-14"},
    {"value": 95, "timestamp": "2026-04-15"}
  ],
  "favorites": null,
  "minutesPlayed": null,
  "recommendations": null,
  "plays": null,
  "uniquePlayers": null,
  "retention": null
}
```

- [ ] **Step 2: Add failing test to `internal/slim/slim_test.go`**

```go
func TestSlimBundledMetricsDayInterval(t *testing.T) {
	raw := readFile(t, "testdata/raw/metrics_bundled.json")
	expected := readFile(t, "testdata/slim/metrics_bundled.json")

	got, err := BundledMetrics(raw, "day")
	require.NoError(t, err)
	assertJSONEqual(t, expected, got)
}

func TestRoundTimestamp(t *testing.T) {
	require.Equal(t, "2026-04-14", roundTimestamp("2026-04-14T00:00:00.000Z", "day"))
	require.Equal(t, "2026-04-14T07:00", roundTimestamp("2026-04-14T07:00:00.000Z", "hour"))
	require.Equal(t, "2026-04-14T07:20", roundTimestamp("2026-04-14T07:20:00.000Z", "minute"))
	require.Equal(t, "garbage", roundTimestamp("garbage", "day"))
}
```

- [ ] **Step 3: Run test to verify it fails**

```bash
go test ./internal/slim/
```

Expected: FAIL with "undefined: BundledMetrics" / "undefined: roundTimestamp".

- [ ] **Step 4: Append to `internal/slim/slim.go`**

```go
import (
	// add to existing imports:
	"time"
)

// BundledMetrics slims a BundledMetrics response (from /metrics or /metrics/{interval}).
// interval must be "day", "hour", or "minute" and controls timestamp rounding.
func BundledMetrics(raw []byte, interval string) ([]byte, error) {
	var in map[string]any
	if err := json.Unmarshal(raw, &in); err != nil {
		return nil, fmt.Errorf("slim bundled metrics: decode: %w", err)
	}

	// Known metric arrays. Retention is special (d1/d7 instead of value).
	metricKeys := []string{
		"averageMinutesPerPlayer",
		"peakCCU",
		"favorites",
		"minutesPlayed",
		"recommendations",
		"plays",
		"uniquePlayers",
	}
	out := map[string]any{}
	for _, k := range metricKeys {
		arr, ok := in[k].([]any)
		if !ok {
			out[k] = nil
			continue
		}
		out[k] = slimValueArray(arr, interval)
	}
	// Retention.
	if retArr, ok := in["retention"].([]any); ok {
		out["retention"] = slimRetentionArray(retArr, interval)
	} else {
		out["retention"] = nil
	}

	return json.Marshal(out)
}

// slimValueArray rounds timestamps and collapses all-null arrays to nil.
func slimValueArray(arr []any, interval string) any {
	allNull := true
	slim := make([]map[string]any, 0, len(arr))
	for _, item := range arr {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		value := m["value"]
		ts, _ := m["timestamp"].(string)
		if value != nil {
			allNull = false
		}
		slim = append(slim, map[string]any{
			"value":     value,
			"timestamp": roundTimestamp(ts, interval),
		})
	}
	if allNull {
		return nil
	}
	return slim
}

// slimRetentionArray collapses fully-null retention arrays (where both d1 and d7 are null) to nil.
func slimRetentionArray(arr []any, interval string) any {
	allNull := true
	slim := make([]map[string]any, 0, len(arr))
	for _, item := range arr {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		d1 := m["d1"]
		d7 := m["d7"]
		ts, _ := m["timestamp"].(string)
		if d1 != nil || d7 != nil {
			allNull = false
		}
		slim = append(slim, map[string]any{
			"d1":        d1,
			"d7":        d7,
			"timestamp": roundTimestamp(ts, interval),
		})
	}
	if allNull {
		return nil
	}
	return slim
}

// roundTimestamp truncates an ISO8601 timestamp to the given interval granularity.
// day → "2026-04-14", hour → "2026-04-14T07:00", minute → "2026-04-14T07:20".
// Returns the input unchanged if it can't be parsed.
func roundTimestamp(ts, interval string) string {
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return ts
	}
	switch interval {
	case "day":
		return t.UTC().Format("2006-01-02")
	case "hour":
		return t.UTC().Format("2006-01-02T15:04")
	case "minute":
		return t.UTC().Format("2006-01-02T15:04")
	default:
		return ts
	}
}
```

- [ ] **Step 5: Run test to verify pass**

```bash
go test ./internal/slim/
```

Expected: PASS (3 tests total now).

- [ ] **Step 6: Commit**

```bash
git add internal/slim/slim.go internal/slim/slim_test.go testdata/slim/metrics_bundled.json
git commit -m "feat(slim): add bundled metrics slimming with timestamp rounding"
```

---

## Task 8: Slim JSON — single-metric + retention responses

**Files:**
- Modify: `internal/slim/slim.go`
- Modify: `internal/slim/slim_test.go`
- Create: `testdata/slim/metric_peak_ccu.json`
- Create: `testdata/slim/retention.json`

- [ ] **Step 1: Create expected goldens**

`testdata/slim/metric_peak_ccu.json`:
```json
{
  "intervals": [
    {"value": 120, "timestamp": "2026-04-14T07:00"},
    {"value": 95, "timestamp": "2026-04-14T08:00"},
    {"value": null, "timestamp": "2026-04-14T09:00"}
  ]
}
```

`testdata/slim/retention.json`:
```json
{
  "intervals": [
    {"d1": 80, "d7": 45, "timestamp": "2026-04-14"},
    {"d1": null, "d7": null, "timestamp": "2026-04-15"}
  ]
}
```

- [ ] **Step 2: Add failing tests**

```go
func TestSlimMetricHourInterval(t *testing.T) {
	raw := readFile(t, "testdata/raw/metric_peak_ccu.json")
	expected := readFile(t, "testdata/slim/metric_peak_ccu.json")

	got, err := MetricResponse(raw, "hour")
	require.NoError(t, err)
	assertJSONEqual(t, expected, got)
}

func TestSlimRetention(t *testing.T) {
	raw := readFile(t, "testdata/raw/retention.json")
	expected := readFile(t, "testdata/slim/retention.json")

	got, err := Retention(raw)
	require.NoError(t, err)
	assertJSONEqual(t, expected, got)
}
```

- [ ] **Step 3: Run to verify fail**

```bash
go test ./internal/slim/
```

Expected: FAIL with "undefined: MetricResponse" / "undefined: Retention".

- [ ] **Step 4: Append to `internal/slim/slim.go`**

```go
// MetricResponse slims a single-metric response (peak-ccu, plays, etc.).
// When all interval values are non-null, the array is preserved. When all are
// null, we keep the array (unlike bundled metrics) since the single-metric
// endpoint is queried precisely to see the series; collapsing to null would
// discard what the caller asked for.
func MetricResponse(raw []byte, interval string) ([]byte, error) {
	var in struct {
		Intervals []map[string]any `json:"intervals"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return nil, fmt.Errorf("slim metric: decode: %w", err)
	}
	out := struct {
		Intervals []map[string]any `json:"intervals"`
	}{Intervals: make([]map[string]any, 0, len(in.Intervals))}
	for _, item := range in.Intervals {
		ts, _ := item["timestamp"].(string)
		out.Intervals = append(out.Intervals, map[string]any{
			"value":     item["value"],
			"timestamp": roundTimestamp(ts, interval),
		})
	}
	return json.Marshal(out)
}

// Retention slims a retention response. Always uses day granularity since
// retention is day-only per the API spec.
func Retention(raw []byte) ([]byte, error) {
	var in struct {
		Intervals []map[string]any `json:"intervals"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return nil, fmt.Errorf("slim retention: decode: %w", err)
	}
	out := struct {
		Intervals []map[string]any `json:"intervals"`
	}{Intervals: make([]map[string]any, 0, len(in.Intervals))}
	for _, item := range in.Intervals {
		ts, _ := item["timestamp"].(string)
		out.Intervals = append(out.Intervals, map[string]any{
			"d1":        item["d1"],
			"d7":        item["d7"],
			"timestamp": roundTimestamp(ts, "day"),
		})
	}
	return json.Marshal(out)
}
```

- [ ] **Step 5: Run to verify pass**

```bash
go test ./internal/slim/
```

Expected: PASS (5 tests total now).

- [ ] **Step 6: Commit**

```bash
git add internal/slim/slim.go internal/slim/slim_test.go testdata/slim/metric_peak_ccu.json testdata/slim/retention.json
git commit -m "feat(slim): add single-metric and retention slimming"
```

---

## Task 9: Output writer (JSON / NDJSON / raw passthrough)

**Files:**
- Create: `internal/output/writer.go`
- Create: `internal/output/writer_test.go`

- [ ] **Step 1: Write failing test `internal/output/writer_test.go`**

```go
package output

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteJSON(t *testing.T) {
	var buf bytes.Buffer
	w := &Writer{Out: &buf, Format: FormatJSON}
	require.NoError(t, w.Write(map[string]any{"a": 1, "b": "hi"}))
	require.JSONEq(t, `{"a":1,"b":"hi"}`, buf.String())
}

func TestWriteNDJSONIterates(t *testing.T) {
	var buf bytes.Buffer
	w := &Writer{Out: &buf, Format: FormatNDJSON}
	require.NoError(t, w.Write(map[string]any{"n": 1}))
	require.NoError(t, w.Write(map[string]any{"n": 2}))
	require.Equal(t, "{\"n\":1}\n{\"n\":2}\n", buf.String())
}

func TestWriteRawBytes(t *testing.T) {
	var buf bytes.Buffer
	w := &Writer{Out: &buf, Format: FormatJSON}
	require.NoError(t, w.WriteRaw([]byte(`{"already":"json"}`)))
	require.Equal(t, `{"already":"json"}`+"\n", buf.String())
}
```

- [ ] **Step 2: Run to verify fail**

```bash
go test ./internal/output/
```

Expected: FAIL with "undefined: Writer".

- [ ] **Step 3: Write `internal/output/writer.go`**

```go
// Package output writes JSON and NDJSON to an io.Writer with a --raw bypass.
package output

import (
	"encoding/json"
	"fmt"
	"io"
)

// Format controls how values are written.
type Format int

const (
	FormatJSON Format = iota
	FormatNDJSON
)

// Writer writes values to Out in the chosen Format.
type Writer struct {
	Out    io.Writer
	Format Format
}

// Write encodes v and writes it.
//   - FormatJSON: single line JSON, no trailing newline between calls — callers should
//     only call Write once unless they know what they're doing.
//   - FormatNDJSON: one JSON object per line, each call appends a newline.
func (w *Writer) Write(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("encode: %w", err)
	}
	if _, err := w.Out.Write(data); err != nil {
		return err
	}
	if w.Format == FormatNDJSON {
		_, err = fmt.Fprintln(w.Out)
	}
	return err
}

// WriteRaw writes pre-encoded JSON bytes followed by a newline.
// Used by the --raw flag to pass the API response through untouched (modulo
// a trailing newline).
func (w *Writer) WriteRaw(data []byte) error {
	if _, err := w.Out.Write(data); err != nil {
		return err
	}
	_, err := fmt.Fprintln(w.Out)
	return err
}
```

- [ ] **Step 4: Run to verify pass**

```bash
go test ./internal/output/
```

Expected: PASS (3 tests).

- [ ] **Step 5: Commit**

```bash
git add internal/output/
git commit -m "feat(output): add JSON/NDJSON writer with raw bypass"
```

---

## Task 10: Command root + shared flags + version

**Files:**
- Create: `internal/commands/root.go`
- Modify: `cmd/fortnite/main.go`

- [ ] **Step 1: Add cobra dependency**

```bash
go get github.com/spf13/cobra@latest
go mod tidy
```

- [ ] **Step 2: Write `internal/commands/root.go`**

```go
// Package commands wires cobra commands to the pure packages.
package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/spf13/cobra"

	"github.com/ygncode/fortnite-cli/internal/api"
	"github.com/ygncode/fortnite-cli/internal/output"
)

// GlobalFlags are cross-cutting flags available on every command.
type GlobalFlags struct {
	Raw         bool
	Format      string // "json" | "ndjson"
	NoRetry     bool
	Timeout     time.Duration
	Concurrency int
}

// NewRoot builds the root `fortnite` command wired with all subcommands.
func NewRoot(version string) *cobra.Command {
	gf := &GlobalFlags{}
	root := &cobra.Command{
		Use:          "fortnite",
		Short:        "Query the Fortnite Ecosystem v1 Data API (island metadata and metrics).",
		SilenceUsage: true, // don't print usage on runtime errors
		Version:      version,
	}

	root.PersistentFlags().BoolVar(&gf.Raw, "raw", false, "emit the API response as-is (no slimming)")
	root.PersistentFlags().StringVar(&gf.Format, "format", "json", "output format: json or ndjson")
	root.PersistentFlags().BoolVar(&gf.NoRetry, "no-retry", false, "disable 429/5xx retry")
	root.PersistentFlags().DurationVar(&gf.Timeout, "timeout", 30*time.Second, "HTTP timeout")
	root.PersistentFlags().IntVar(&gf.Concurrency, "concurrency", 8, "worker pool size for composite verbs")

	// islands subtree
	islands := &cobra.Command{Use: "islands", Short: "Island metadata and metrics endpoints."}
	islands.AddCommand(newListCmd(gf))
	islands.AddCommand(newGetCmd(gf))
	islands.AddCommand(newMetricsCmd(gf))
	for _, m := range singleMetricSpecs {
		islands.AddCommand(newSingleMetricCmd(gf, m))
	}
	islands.AddCommand(newRetentionCmd(gf))
	root.AddCommand(islands)

	// composite verbs
	root.AddCommand(newTopCmd(gf))
	root.AddCommand(newSummaryCmd(gf))
	root.AddCommand(newCompareCmd(gf))
	root.AddCommand(newSearchCmd(gf))

	return root
}

// newClient builds a configured api.Client from the global flags.
func newClient(gf *GlobalFlags) *api.Client {
	c := api.DefaultClient(gf.Timeout)
	c.NoRetry = gf.NoRetry
	return c
}

// newWriter builds an output.Writer from the global flags.
func newWriter(gf *GlobalFlags) *output.Writer {
	f := output.FormatJSON
	if gf.Format == "ndjson" {
		f = output.FormatNDJSON
	}
	return &output.Writer{Out: os.Stdout, Format: f}
}

// withSignalContext returns a context cancelled on Ctrl-C.
func withSignalContext() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt)
}

// printError writes a structured error to stderr.
//
//	{"error": "<code>", "message": "<msg>", "status": <code-or-null>, "uuid": "<uuid-or-omitted>"}
func printError(err error) {
	out := map[string]any{"error": "generic", "message": err.Error()}
	var apiErr *api.APIError
	if asAPIErr(err, &apiErr) {
		out["error"] = classify(apiErr)
		out["message"] = apiErr.Message
		out["status"] = apiErr.Status
		if apiErr.UUID != "" {
			out["uuid"] = apiErr.UUID
		}
	}
	data, _ := json.Marshal(out)
	fmt.Fprintln(os.Stderr, string(data))
}

func classify(apiErr *api.APIError) string {
	switch {
	case apiErr.Status == 429:
		return "rate_limited"
	case apiErr.Status == 404:
		return "not_found"
	case apiErr.Status == 400:
		return "bad_request"
	case apiErr.Status >= 500:
		return "upstream_error"
	default:
		return "api_error"
	}
}

// asAPIErr is a tiny shim so root.go does not need to import "errors" just to alias errors.As.
func asAPIErr(err error, target **api.APIError) bool {
	for err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			*target = apiErr
			return true
		}
		u, ok := err.(interface{ Unwrap() error })
		if !ok {
			return false
		}
		err = u.Unwrap()
	}
	return false
}

// exitCodeFor classifies an error into the documented exit codes.
// 0=success, 1=usage/validation, 2=API error, 3=network/IO.
func exitCodeFor(err error) int {
	if err == nil {
		return 0
	}
	var apiErr *api.APIError
	if asAPIErr(err, &apiErr) {
		return 2
	}
	return 3
}
```

- [ ] **Step 3: Update `cmd/fortnite/main.go`**

```go
package main

import (
	"os"

	"github.com/ygncode/fortnite-cli/internal/api"
	"github.com/ygncode/fortnite-cli/internal/commands"
)

// version is injected by ldflags at release time.
var version = "dev"

func main() {
	api.Version = version
	root := commands.NewRoot(version)
	if err := root.Execute(); err != nil {
		// Cobra prints its own usage errors before returning; we only need to
		// translate to the documented exit code.
		os.Exit(commands.ExitCodeFor(err))
	}
}
```

Note: `ExitCodeFor` needs to be exported. Rename `exitCodeFor` → `ExitCodeFor` in `internal/commands/root.go`. (Update the package-private uses too.)

Step 3 continues: rename in `internal/commands/root.go`:

```go
// ExitCodeFor classifies an error into the documented exit codes.
// 0=success, 1=usage/validation, 2=API error, 3=network/IO.
func ExitCodeFor(err error) int {
	if err == nil {
		return 0
	}
	var apiErr *api.APIError
	if asAPIErr(err, &apiErr) {
		return 2
	}
	return 3
}
```

- [ ] **Step 4: Build (no tests yet; subcommands land in following tasks)**

Run:
```bash
go build ./...
```

Expected: FAIL — `newListCmd`, `newGetCmd`, etc. are not yet defined. That's expected. In the next task we implement the first subcommand file; once every subcommand exists, the build succeeds.

For this task's commit boundary, stub the subcommand constructors in `internal/commands/root.go` temporarily:

```go
// Temporary stubs — replaced in subsequent tasks.
var singleMetricSpecs = []singleMetricSpec{}

type singleMetricSpec struct {
	Use, Short, Metric string
	DayOnly            bool
}

func newListCmd(_ *GlobalFlags) *cobra.Command        { return &cobra.Command{Use: "list", Hidden: true} }
func newGetCmd(_ *GlobalFlags) *cobra.Command         { return &cobra.Command{Use: "get", Hidden: true} }
func newMetricsCmd(_ *GlobalFlags) *cobra.Command     { return &cobra.Command{Use: "metrics", Hidden: true} }
func newSingleMetricCmd(_ *GlobalFlags, _ singleMetricSpec) *cobra.Command {
	return &cobra.Command{Hidden: true}
}
func newRetentionCmd(_ *GlobalFlags) *cobra.Command { return &cobra.Command{Use: "retention", Hidden: true} }
func newTopCmd(_ *GlobalFlags) *cobra.Command       { return &cobra.Command{Use: "top", Hidden: true} }
func newSummaryCmd(_ *GlobalFlags) *cobra.Command   { return &cobra.Command{Use: "summary", Hidden: true} }
func newCompareCmd(_ *GlobalFlags) *cobra.Command   { return &cobra.Command{Use: "compare", Hidden: true} }
func newSearchCmd(_ *GlobalFlags) *cobra.Command    { return &cobra.Command{Use: "search", Hidden: true} }
```

Put the stubs in a new file `internal/commands/stubs.go` so they can be deleted individually as real implementations land.

- [ ] **Step 5: Build succeeds**

```bash
go build ./...
./fortnite --version
```

Expected: prints `fortnite version dev`.

- [ ] **Step 6: Commit**

```bash
git add internal/commands/ cmd/fortnite/main.go go.mod go.sum
git commit -m "feat(commands): add cobra root and global flags"
```

---

## Task 11: Command — `islands list`

**Files:**
- Create: `internal/commands/islands_list.go`
- Modify: `internal/commands/stubs.go` (remove the `newListCmd` stub)
- Create: `internal/commands/islands_list_test.go`

- [ ] **Step 1: Write failing test**

```go
package commands

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ygncode/fortnite-cli/internal/api"
)

// captureStdout replaces os.Stdout for the duration of fn and returns whatever was written.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	defer func() { os.Stdout = old }()
	fn()
	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	return buf.String()
}

// fortniteServer spins up an httptest.Server backed by testdata files keyed by method+path.
func fortniteServer(t *testing.T, routes map[string]string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Method + " " + r.URL.Path
		body, ok := routes[key]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
}

func readRepoFile(t *testing.T, rel string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", rel))
	require.NoError(t, err)
	return data
}

func TestIslandsListSlimmed(t *testing.T) {
	body := string(readRepoFile(t, "testdata/raw/islands_list.json"))
	srv := fortniteServer(t, map[string]string{"GET /islands": body})
	defer srv.Close()

	gf := &GlobalFlags{Format: "json"}
	client := api.DefaultClient(0)
	client.BaseURL = srv.URL

	var stdout bytes.Buffer
	err := runIslandsList(context.Background(), client, &stdout, gf, listFlags{Size: 2})
	require.NoError(t, err)
	require.Contains(t, stdout.String(), `"7022-1566-6527"`)
	require.Contains(t, stdout.String(), `"_pagination"`)
	require.NotContains(t, stdout.String(), `"meta":{"page"`) // no raw nested meta
}

func TestIslandsListRawPassesThrough(t *testing.T) {
	body := string(readRepoFile(t, "testdata/raw/islands_list.json"))
	srv := fortniteServer(t, map[string]string{"GET /islands": body})
	defer srv.Close()

	gf := &GlobalFlags{Raw: true, Format: "json"}
	client := api.DefaultClient(0)
	client.BaseURL = srv.URL

	var stdout bytes.Buffer
	err := runIslandsList(context.Background(), client, &stdout, gf, listFlags{Size: 2})
	require.NoError(t, err)
	// Raw passthrough contains the original wrapper.
	require.Contains(t, stdout.String(), `"links"`)
	require.Contains(t, stdout.String(), `"page":{"prevCursor"`)
}
```

- [ ] **Step 2: Run to verify fail**

```bash
go test ./internal/commands/ -run TestIslandsList
```

Expected: FAIL with "undefined: runIslandsList" / "undefined: listFlags".

- [ ] **Step 3: Write `internal/commands/islands_list.go`**

```go
package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/ygncode/fortnite-cli/internal/api"
	"github.com/ygncode/fortnite-cli/internal/output"
	"github.com/ygncode/fortnite-cli/internal/slim"
)

type listFlags struct {
	Size   int
	After  string
	Before string
}

func newListCmd(gf *GlobalFlags) *cobra.Command {
	var f listFlags
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List islands (cursor-paginated, newest-released first).",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := withSignalContext()
			defer cancel()
			client := newClient(gf)
			if err := runIslandsList(ctx, client, os.Stdout, gf, f); err != nil {
				printError(err)
				return err
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&f.Size, "size", 100, "max results per page (1-1000)")
	cmd.Flags().StringVar(&f.After, "after", "", "cursor: return results after this position")
	cmd.Flags().StringVar(&f.Before, "before", "", "cursor: return results before this position")
	return cmd
}

// runIslandsList is the testable core of the command.
func runIslandsList(ctx context.Context, client *api.Client, stdout io.Writer, gf *GlobalFlags, f listFlags) error {
	if f.Size < 1 || f.Size > 1000 {
		return fmt.Errorf("--size must be in [1,1000], got %d", f.Size)
	}

	// Raw mode: bypass typed decoding and stream the original JSON.
	// We still go through Get because it handles retries, headers, error decoding.
	if gf.Raw {
		raw, err := getRaw(ctx, client, "/islands", listQueryValues(f))
		if err != nil {
			return err
		}
		w := &output.Writer{Out: stdout, Format: formatFrom(gf)}
		return w.WriteRaw(raw)
	}

	resp, err := client.ListIslands(ctx, api.ListIslandsParams{Size: f.Size, After: f.After, Before: f.Before})
	if err != nil {
		return err
	}
	rawRespBytes, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	slimmed, err := slim.IslandList(rawRespBytes)
	if err != nil {
		return err
	}
	w := &output.Writer{Out: stdout, Format: formatFrom(gf)}
	return w.WriteRaw(slimmed)
}

func formatFrom(gf *GlobalFlags) output.Format {
	if gf.Format == "ndjson" {
		return output.FormatNDJSON
	}
	return output.FormatJSON
}
```

- [ ] **Step 4: Add `getRaw` and `listQueryValues` helpers**

Append to `internal/commands/root.go`:

```go
import (
	// add to imports:
	"net/url"
	"strconv"
)

// getRaw performs a GET and returns the raw response body. Uses the client's
// retry/error machinery but skips JSON decoding.
func getRaw(ctx context.Context, client *api.Client, path string, q url.Values) ([]byte, error) {
	// We roundtrip through a *json.RawMessage so the client's decoder can still
	// catch non-JSON error bodies as APIError.
	var raw json.RawMessage
	if err := client.Get(ctx, path, q, &raw); err != nil {
		return nil, err
	}
	return []byte(raw), nil
}

func listQueryValues(f listFlags) url.Values {
	q := url.Values{}
	if f.Size > 0 {
		q.Set("size", strconv.Itoa(f.Size))
	}
	if f.After != "" {
		q.Set("after", f.After)
	}
	if f.Before != "" {
		q.Set("before", f.Before)
	}
	return q
}
```

- [ ] **Step 5: Remove the `newListCmd` stub**

Delete the `newListCmd` stub line from `internal/commands/stubs.go`. The real `newListCmd` in `islands_list.go` takes over.

- [ ] **Step 6: Run tests**

```bash
go test ./internal/commands/ -run TestIslandsList
```

Expected: PASS (2 tests).

- [ ] **Step 7: Manual smoke test**

```bash
go build ./...
./fortnite islands list --size 3
```

Expected: hits the real API, returns 3 slimmed islands to stdout, exit 0.

- [ ] **Step 8: Commit**

```bash
git add internal/commands/ cmd/fortnite/main.go
git commit -m "feat(cmd): implement islands list (slim + --raw passthrough)"
```

---

## Task 12: Command — `islands get`

**Files:**
- Create: `internal/commands/islands_get.go`
- Modify: `internal/commands/stubs.go`
- Create: `internal/commands/islands_get_test.go`

- [ ] **Step 1: Write failing test**

```go
func TestIslandsGet(t *testing.T) {
	body := string(readRepoFile(t, "testdata/raw/island_get.json"))
	srv := fortniteServer(t, map[string]string{
		"GET /islands/7022-1566-6527": body,
	})
	defer srv.Close()

	gf := &GlobalFlags{Format: "json"}
	client := api.DefaultClient(0)
	client.BaseURL = srv.URL

	var stdout bytes.Buffer
	err := runIslandsGet(context.Background(), client, &stdout, gf, "7022-1566-6527")
	require.NoError(t, err)
	require.Contains(t, stdout.String(), `"title":"BRAINROT RACE"`)
}
```

Place in `internal/commands/islands_get_test.go`.

- [ ] **Step 2: Run to verify fail**

```bash
go test ./internal/commands/ -run TestIslandsGet
```

- [ ] **Step 3: Write `internal/commands/islands_get.go`**

```go
package commands

import (
	"context"
	"encoding/json"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/ygncode/fortnite-cli/internal/api"
	"github.com/ygncode/fortnite-cli/internal/output"
)

func newGetCmd(gf *GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <code>",
		Short: "Get metadata for a single island.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := withSignalContext()
			defer cancel()
			client := newClient(gf)
			if err := runIslandsGet(ctx, client, os.Stdout, gf, args[0]); err != nil {
				printError(err)
				return err
			}
			return nil
		},
	}
	return cmd
}

func runIslandsGet(ctx context.Context, client *api.Client, stdout io.Writer, gf *GlobalFlags, code string) error {
	if gf.Raw {
		raw, err := getRaw(ctx, client, "/islands/"+code, nil)
		if err != nil {
			return err
		}
		w := &output.Writer{Out: stdout, Format: formatFrom(gf)}
		return w.WriteRaw(raw)
	}
	meta, err := client.GetIsland(ctx, code)
	if err != nil {
		return err
	}
	// GetIsland is already a flat metadata struct; omit empty fields via the json tags.
	data, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	w := &output.Writer{Out: stdout, Format: formatFrom(gf)}
	return w.WriteRaw(data)
}
```

- [ ] **Step 4: Remove `newGetCmd` stub**

- [ ] **Step 5: Run tests**

```bash
go test ./internal/commands/
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/commands/
git commit -m "feat(cmd): implement islands get"
```

---

## Task 13: Command — `islands metrics` (dispatch between bundled and filterable)

**Files:**
- Create: `internal/commands/islands_metrics.go`
- Modify: `internal/commands/stubs.go`
- Create: `internal/commands/islands_metrics_test.go`

- [ ] **Step 1: Write failing test**

```go
func TestIslandsMetricsBundled(t *testing.T) {
	body := string(readRepoFile(t, "testdata/raw/metrics_bundled.json"))
	srv := fortniteServer(t, map[string]string{
		"GET /islands/7022-1566-6527/metrics": body,
	})
	defer srv.Close()

	gf := &GlobalFlags{Format: "json"}
	client := api.DefaultClient(0)
	client.BaseURL = srv.URL

	var stdout bytes.Buffer
	err := runIslandsMetrics(context.Background(), client, &stdout, gf, metricsFlags{Code: "7022-1566-6527"})
	require.NoError(t, err)
	require.Contains(t, stdout.String(), `"peakCCU"`)
	require.Contains(t, stdout.String(), `"favorites":null`)
}

func TestIslandsMetricsFilterable(t *testing.T) {
	body := string(readRepoFile(t, "testdata/raw/metrics_bundled.json"))
	var gotPath, gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	gf := &GlobalFlags{Format: "json"}
	client := api.DefaultClient(0)
	client.BaseURL = srv.URL

	var stdout bytes.Buffer
	err := runIslandsMetrics(context.Background(), client, &stdout, gf, metricsFlags{
		Code:     "7022-1566-6527",
		Interval: "day",
		Metrics:  []string{"peakCCU", "plays"},
	})
	require.NoError(t, err)
	require.Equal(t, "/islands/7022-1566-6527/metrics/day", gotPath)
	require.Contains(t, gotQuery, "metrics=peakCCU")
	require.Contains(t, gotQuery, "metrics=plays")
}
```

Place in `internal/commands/islands_metrics_test.go`.

- [ ] **Step 2: Run to verify fail**

```bash
go test ./internal/commands/ -run TestIslandsMetrics
```

- [ ] **Step 3: Write `internal/commands/islands_metrics.go`**

```go
package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/ygncode/fortnite-cli/internal/api"
	"github.com/ygncode/fortnite-cli/internal/output"
	"github.com/ygncode/fortnite-cli/internal/slim"
	"github.com/ygncode/fortnite-cli/internal/timerange"
)

type metricsFlags struct {
	Code     string
	Interval string
	Metrics  []string
	Last     string
	From     string
	To       string
}

func newMetricsCmd(gf *GlobalFlags) *cobra.Command {
	var f metricsFlags
	cmd := &cobra.Command{
		Use:   "metrics <code>",
		Short: "Island metrics. Without --interval: bundled day metrics. With --interval: filterable.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f.Code = args[0]
			ctx, cancel := withSignalContext()
			defer cancel()
			client := newClient(gf)
			if err := runIslandsMetrics(ctx, client, os.Stdout, gf, f); err != nil {
				printError(err)
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&f.Interval, "interval", "", "day | hour | minute (omit for bundled day response)")
	cmd.Flags().StringSliceVar(&f.Metrics, "metric", nil, "filter to specific metrics (repeatable; requires --interval)")
	cmd.Flags().StringVar(&f.Last, "last", "", "shorthand time range, e.g. 24h, 7d")
	cmd.Flags().StringVar(&f.From, "from", "", "start of range (RFC3339)")
	cmd.Flags().StringVar(&f.To, "to", "", "end of range (RFC3339)")
	return cmd
}

func runIslandsMetrics(ctx context.Context, client *api.Client, stdout io.Writer, gf *GlobalFlags, f metricsFlags) error {
	interval := f.Interval
	effective := interval
	if effective == "" {
		effective = "day"
	}
	r, warn, err := timerange.Parse(timerange.Input{
		Last: f.Last, From: f.From, To: f.To, Interval: effective,
	}, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("time range: %w", err)
	}
	if warn != "" {
		fmt.Fprintln(os.Stderr, "warning:", warn)
	}
	if len(f.Metrics) > 0 && interval == "" {
		return fmt.Errorf("--metric requires --interval")
	}

	p := api.MetricsParams{From: r.From, To: r.To, Metrics: f.Metrics}

	if gf.Raw {
		path := "/islands/" + f.Code + "/metrics"
		if interval != "" {
			path += "/" + interval
		}
		raw, err := getRaw(ctx, client, path, p.EncodePublic())
		if err != nil {
			return err
		}
		w := &output.Writer{Out: stdout, Format: formatFrom(gf)}
		return w.WriteRaw(raw)
	}

	var resp *api.BundledMetrics
	if interval == "" {
		resp, err = client.GetBundledMetrics(ctx, f.Code, p)
	} else {
		resp, err = client.GetFilterableMetrics(ctx, f.Code, interval, p)
	}
	if err != nil {
		return err
	}
	raw, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	slimmed, err := slim.BundledMetrics(raw, effective)
	if err != nil {
		return err
	}
	w := &output.Writer{Out: stdout, Format: formatFrom(gf)}
	return w.WriteRaw(slimmed)
}
```

- [ ] **Step 4: Export `MetricsParams.encode`**

`MetricsParams.encode()` in `internal/api/islands.go` is lowercase. The `runIslandsMetrics` function (in `internal/commands`) can't call it directly. Either (a) make an exported wrapper or (b) expose a method. Pick (a) by adding to `internal/api/islands.go`:

```go
// EncodePublic returns the URL values the client would send for this MetricsParams.
// Used by the commands package when it needs to build a raw request path.
func (p MetricsParams) EncodePublic() url.Values { return p.encode() }
```

- [ ] **Step 5: Remove `newMetricsCmd` stub**

- [ ] **Step 6: Run tests**

```bash
go test ./internal/commands/ -run TestIslandsMetrics
go test ./...
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/api/islands.go internal/commands/
git commit -m "feat(cmd): implement islands metrics with bundled/filterable dispatch"
```

---

## Task 14: Commands — 7 per-metric verbs (peak-ccu through plays)

**Files:**
- Create: `internal/commands/islands_single_metric.go`
- Modify: `internal/commands/stubs.go`
- Create: `internal/commands/islands_single_metric_test.go`

This task creates all 7 non-retention per-metric commands via a single `newSingleMetricCmd` factory driven by a spec table. Retention gets its own file (Task 15) because its response shape is different.

- [ ] **Step 1: Write failing test**

```go
func TestSingleMetricPeakCCU(t *testing.T) {
	body := string(readRepoFile(t, "testdata/raw/metric_peak_ccu.json"))
	srv := fortniteServer(t, map[string]string{
		"GET /islands/7022-1566-6527/metrics/hour/peak-ccu": body,
	})
	defer srv.Close()

	gf := &GlobalFlags{Format: "json"}
	client := api.DefaultClient(0)
	client.BaseURL = srv.URL

	spec := singleMetricSpec{Use: "peak-ccu", Short: "", Metric: "peak-ccu"}
	var stdout bytes.Buffer
	err := runSingleMetric(context.Background(), client, &stdout, gf, spec, singleMetricFlags{
		Code: "7022-1566-6527", Interval: "hour",
	})
	require.NoError(t, err)
	require.Contains(t, stdout.String(), `"intervals"`)
	require.Contains(t, stdout.String(), `"2026-04-14T07:00"`) // hour-rounded
}

func TestSingleMetricAvgMinutesRejectsHour(t *testing.T) {
	gf := &GlobalFlags{Format: "json"}
	spec := singleMetricSpec{Use: "avg-minutes", Metric: "average-minutes-per-player", DayOnly: true}
	var stdout bytes.Buffer
	err := runSingleMetric(context.Background(), nil, &stdout, gf, spec, singleMetricFlags{
		Code: "7022-1566-6527", Interval: "hour",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "day")
}
```

- [ ] **Step 2: Run to verify fail**

```bash
go test ./internal/commands/ -run TestSingleMetric
```

- [ ] **Step 3: Write `internal/commands/islands_single_metric.go`**

```go
package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/ygncode/fortnite-cli/internal/api"
	"github.com/ygncode/fortnite-cli/internal/output"
	"github.com/ygncode/fortnite-cli/internal/slim"
	"github.com/ygncode/fortnite-cli/internal/timerange"
)

// singleMetricSpecs is the catalogue of non-retention per-metric commands.
// Populated here so Task 14 can delete the stub in stubs.go without breaking the root.
var singleMetricSpecs = []singleMetricSpec{
	{Use: "peak-ccu", Short: "Peak concurrent players.", Metric: "peak-ccu"},
	{Use: "favorites", Short: "Times the island was favorited.", Metric: "favorites"},
	{Use: "minutes-played", Short: "Total player minutes.", Metric: "minutes-played"},
	{Use: "avg-minutes", Short: "Average minutes per player (day only).", Metric: "average-minutes-per-player", DayOnly: true},
	{Use: "recommendations", Short: "Player recommendations.", Metric: "recommendations"},
	{Use: "unique-players", Short: "Unique players.", Metric: "unique-players"},
	{Use: "plays", Short: "Play count.", Metric: "plays"},
}

type singleMetricFlags struct {
	Code     string
	Interval string
	Last     string
	From     string
	To       string
}

func newSingleMetricCmd(gf *GlobalFlags, spec singleMetricSpec) *cobra.Command {
	var f singleMetricFlags
	cmd := &cobra.Command{
		Use:   spec.Use + " <code>",
		Short: spec.Short,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f.Code = args[0]
			ctx, cancel := withSignalContext()
			defer cancel()
			client := newClient(gf)
			if err := runSingleMetric(ctx, client, os.Stdout, gf, spec, f); err != nil {
				printError(err)
				return err
			}
			return nil
		},
	}
	defaultInterval := "day"
	if !spec.DayOnly {
		cmd.Flags().StringVar(&f.Interval, "interval", defaultInterval, "day | hour | minute")
	} else {
		cmd.Flags().StringVar(&f.Interval, "interval", "day", "day only (this metric is day-only)")
	}
	cmd.Flags().StringVar(&f.Last, "last", "", "shorthand time range, e.g. 24h, 7d")
	cmd.Flags().StringVar(&f.From, "from", "", "start of range (RFC3339)")
	cmd.Flags().StringVar(&f.To, "to", "", "end of range (RFC3339)")
	return cmd
}

func runSingleMetric(ctx context.Context, client *api.Client, stdout io.Writer, gf *GlobalFlags, spec singleMetricSpec, f singleMetricFlags) error {
	interval := f.Interval
	if interval == "" {
		interval = "day"
	}
	if spec.DayOnly && interval != "day" {
		return fmt.Errorf("%s is only available at day interval", spec.Metric)
	}
	if interval != "day" && interval != "hour" && interval != "minute" {
		return fmt.Errorf("invalid --interval %q", interval)
	}
	r, warn, err := timerange.Parse(timerange.Input{
		Last: f.Last, From: f.From, To: f.To, Interval: interval,
	}, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("time range: %w", err)
	}
	if warn != "" {
		fmt.Fprintln(os.Stderr, "warning:", warn)
	}
	p := api.MetricsParams{From: r.From, To: r.To}

	if gf.Raw {
		path := "/islands/" + f.Code + "/metrics/" + interval + "/" + spec.Metric
		raw, err := getRaw(ctx, client, path, p.EncodePublic())
		if err != nil {
			return err
		}
		w := &output.Writer{Out: stdout, Format: formatFrom(gf)}
		return w.WriteRaw(raw)
	}

	resp, err := client.GetIslandMetric(ctx, f.Code, interval, spec.Metric, p)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	slimmed, err := slim.MetricResponse(raw, interval)
	if err != nil {
		return err
	}
	w := &output.Writer{Out: stdout, Format: formatFrom(gf)}
	return w.WriteRaw(slimmed)
}
```

- [ ] **Step 4: Move the `singleMetricSpec` type and remove stubs**

In `stubs.go`, delete three things:
- `var singleMetricSpecs = []singleMetricSpec{}`
- `type singleMetricSpec struct { Use, Short, Metric string; DayOnly bool }`
- `func newSingleMetricCmd(_ *GlobalFlags, _ singleMetricSpec) *cobra.Command { ... }`

Add the type definition to the top of `internal/commands/islands_single_metric.go` (just under the imports):

```go
type singleMetricSpec struct {
	Use     string
	Short   string
	Metric  string
	DayOnly bool
}
```

Build to verify no duplicate declarations:

```bash
go build ./...
```

- [ ] **Step 5: Run tests**

```bash
go test ./internal/commands/
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/commands/
git commit -m "feat(cmd): implement 7 per-metric commands via shared factory"
```

---

## Task 15: Command — `islands retention`

**Files:**
- Create: `internal/commands/islands_retention.go`
- Modify: `internal/commands/stubs.go`
- Create: `internal/commands/islands_retention_test.go`

- [ ] **Step 1: Write failing test**

```go
func TestIslandsRetention(t *testing.T) {
	body := string(readRepoFile(t, "testdata/raw/retention.json"))
	srv := fortniteServer(t, map[string]string{
		"GET /islands/7022-1566-6527/metrics/day/retention": body,
	})
	defer srv.Close()

	gf := &GlobalFlags{Format: "json"}
	client := api.DefaultClient(0)
	client.BaseURL = srv.URL

	var stdout bytes.Buffer
	err := runRetention(context.Background(), client, &stdout, gf, retentionFlags{Code: "7022-1566-6527"})
	require.NoError(t, err)
	require.Contains(t, stdout.String(), `"d1":80`)
}

func TestRetentionRejectsHourInterval(t *testing.T) {
	gf := &GlobalFlags{Format: "json"}
	var stdout bytes.Buffer
	err := runRetention(context.Background(), nil, &stdout, gf, retentionFlags{
		Code: "x", Interval: "hour",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "day")
}
```

- [ ] **Step 2: Run to verify fail**

- [ ] **Step 3: Write `internal/commands/islands_retention.go`**

```go
package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/ygncode/fortnite-cli/internal/api"
	"github.com/ygncode/fortnite-cli/internal/output"
	"github.com/ygncode/fortnite-cli/internal/slim"
	"github.com/ygncode/fortnite-cli/internal/timerange"
)

type retentionFlags struct {
	Code     string
	Interval string
	Last     string
	From     string
	To       string
}

func newRetentionCmd(gf *GlobalFlags) *cobra.Command {
	var f retentionFlags
	cmd := &cobra.Command{
		Use:   "retention <code>",
		Short: "D1/D7 retention (day only).",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f.Code = args[0]
			ctx, cancel := withSignalContext()
			defer cancel()
			client := newClient(gf)
			if err := runRetention(ctx, client, os.Stdout, gf, f); err != nil {
				printError(err)
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&f.Interval, "interval", "day", "day (only day is valid)")
	cmd.Flags().StringVar(&f.Last, "last", "", "shorthand time range, e.g. 7d")
	cmd.Flags().StringVar(&f.From, "from", "", "start of range (RFC3339)")
	cmd.Flags().StringVar(&f.To, "to", "", "end of range (RFC3339)")
	return cmd
}

func runRetention(ctx context.Context, client *api.Client, stdout io.Writer, gf *GlobalFlags, f retentionFlags) error {
	interval := f.Interval
	if interval == "" {
		interval = "day"
	}
	if interval != "day" {
		return fmt.Errorf("retention is only available at day interval")
	}
	r, warn, err := timerange.Parse(timerange.Input{
		Last: f.Last, From: f.From, To: f.To, Interval: "day",
	}, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("time range: %w", err)
	}
	if warn != "" {
		fmt.Fprintln(os.Stderr, "warning:", warn)
	}
	p := api.MetricsParams{From: r.From, To: r.To}

	if gf.Raw {
		raw, err := getRaw(ctx, client, "/islands/"+f.Code+"/metrics/day/retention", p.EncodePublic())
		if err != nil {
			return err
		}
		w := &output.Writer{Out: stdout, Format: formatFrom(gf)}
		return w.WriteRaw(raw)
	}

	resp, err := client.GetRetention(ctx, f.Code, p)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	slimmed, err := slim.Retention(raw)
	if err != nil {
		return err
	}
	w := &output.Writer{Out: stdout, Format: formatFrom(gf)}
	return w.WriteRaw(slimmed)
}
```

- [ ] **Step 4: Remove `newRetentionCmd` stub**

- [ ] **Step 5: Run tests**

```bash
go test ./...
```

Expected: PASS. All 12 thin 1:1 commands now exist.

- [ ] **Step 6: Smoke test**

```bash
go build ./...
./fortnite islands list --size 3
./fortnite islands peak-ccu 7022-1566-6527 --interval hour
```

Both should exit 0 with JSON output.

- [ ] **Step 7: Commit**

```bash
git add internal/commands/
git commit -m "feat(cmd): implement islands retention (day-only)"
```

---

## Task 16: Composite — `summary`

**Files:**
- Create: `internal/commands/summary.go`
- Modify: `internal/commands/stubs.go`
- Create: `internal/commands/summary_test.go`

- [ ] **Step 1: Write failing test**

```go
func TestSummaryMergesMetadataAndMetrics(t *testing.T) {
	metaBody := string(readRepoFile(t, "testdata/raw/island_get.json"))
	metricsBody := string(readRepoFile(t, "testdata/raw/metrics_bundled.json"))
	srv := fortniteServer(t, map[string]string{
		"GET /islands/7022-1566-6527":         metaBody,
		"GET /islands/7022-1566-6527/metrics": metricsBody,
	})
	defer srv.Close()

	gf := &GlobalFlags{Format: "json"}
	client := api.DefaultClient(0)
	client.BaseURL = srv.URL

	var stdout bytes.Buffer
	err := runSummary(context.Background(), client, &stdout, gf, summaryFlags{Code: "7022-1566-6527"})
	require.NoError(t, err)
	require.Contains(t, stdout.String(), `"metadata"`)
	require.Contains(t, stdout.String(), `"metrics"`)
	require.Contains(t, stdout.String(), `"title":"BRAINROT RACE"`)
	require.Contains(t, stdout.String(), `"peakCCU"`)
}
```

- [ ] **Step 2: Run to verify fail**

- [ ] **Step 3: Write `internal/commands/summary.go`**

```go
package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/ygncode/fortnite-cli/internal/api"
	"github.com/ygncode/fortnite-cli/internal/output"
	"github.com/ygncode/fortnite-cli/internal/slim"
	"github.com/ygncode/fortnite-cli/internal/timerange"
)

type summaryFlags struct {
	Code string
	Last string
}

func newSummaryCmd(gf *GlobalFlags) *cobra.Command {
	var f summaryFlags
	cmd := &cobra.Command{
		Use:   "summary <code>",
		Short: "Full picture of one island: metadata + all bundled metrics.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f.Code = args[0]
			ctx, cancel := withSignalContext()
			defer cancel()
			client := newClient(gf)
			if err := runSummary(ctx, client, os.Stdout, gf, f); err != nil {
				printError(err)
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&f.Last, "last", "", "shorthand time range, e.g. 7d (default: previous day)")
	return cmd
}

func runSummary(ctx context.Context, client *api.Client, stdout io.Writer, gf *GlobalFlags, f summaryFlags) error {
	r, _, err := timerange.Parse(timerange.Input{Last: f.Last, Interval: "day"}, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("time range: %w", err)
	}

	meta, err := client.GetIsland(ctx, f.Code)
	if err != nil {
		return err
	}
	metrics, err := client.GetBundledMetrics(ctx, f.Code, api.MetricsParams{From: r.From, To: r.To})
	if err != nil {
		return err
	}

	// Slim the metrics piece. We intentionally do NOT slim the metadata (it's
	// already compact). --raw dumps both verbatim instead.
	if gf.Raw {
		combined := map[string]any{"metadata": meta, "metrics": metrics}
		data, err := json.Marshal(combined)
		if err != nil {
			return err
		}
		w := &output.Writer{Out: stdout, Format: formatFrom(gf)}
		return w.WriteRaw(data)
	}

	metricsBytes, err := json.Marshal(metrics)
	if err != nil {
		return err
	}
	slimmed, err := slim.BundledMetrics(metricsBytes, "day")
	if err != nil {
		return err
	}
	var slimMetrics any
	if err := json.Unmarshal(slimmed, &slimMetrics); err != nil {
		return err
	}
	combined := map[string]any{
		"metadata": meta,
		"metrics":  slimMetrics,
	}
	data, err := json.Marshal(combined)
	if err != nil {
		return err
	}
	w := &output.Writer{Out: stdout, Format: formatFrom(gf)}
	return w.WriteRaw(data)
}
```

- [ ] **Step 4: Remove `newSummaryCmd` stub**

- [ ] **Step 5: Run tests**

```bash
go test ./...
```

- [ ] **Step 6: Commit**

```bash
git add internal/commands/
git commit -m "feat(cmd): implement summary composite"
```

---

## Task 17: Composite — `top`

**Files:**
- Create: `internal/commands/top.go`
- Modify: `internal/commands/stubs.go`
- Create: `internal/commands/top_test.go`

- [ ] **Step 1: Write failing test**

```go
func TestTopReturnsSortedResults(t *testing.T) {
	listBody := `{
      "links": {"prev": null, "next": null},
      "meta": {"count": 2, "page": {"prevCursor": "", "nextCursor": ""}},
      "data": [
        {"code": "A", "title": "Alpha", "tags": [], "meta": {"page": {"cursor": "a"}}},
        {"code": "B", "title": "Bravo", "tags": [], "meta": {"page": {"cursor": "b"}}}
      ]
    }`
	metricA := `{"intervals": [{"value": 100, "timestamp": "2026-04-14T07:00:00.000Z"}]}`
	metricB := `{"intervals": [{"value": 200, "timestamp": "2026-04-14T07:00:00.000Z"}]}`
	srv := fortniteServer(t, map[string]string{
		"GET /islands":                            listBody,
		"GET /islands/A/metrics/hour/peak-ccu":    metricA,
		"GET /islands/B/metrics/hour/peak-ccu":    metricB,
	})
	defer srv.Close()

	gf := &GlobalFlags{Format: "json", Concurrency: 2}
	client := api.DefaultClient(0)
	client.BaseURL = srv.URL

	var stdout bytes.Buffer
	err := runTop(context.Background(), client, &stdout, gf, topFlags{
		By: "peakCCU", Interval: "hour", Last: "1h", Limit: 10, Scan: 100,
	})
	require.NoError(t, err)
	// B (200) should come before A (100).
	bIdx := indexOf(stdout.String(), `"B"`)
	aIdx := indexOf(stdout.String(), `"A"`)
	require.True(t, bIdx > 0 && aIdx > 0 && bIdx < aIdx, "B should appear before A in %q", stdout.String())
}

func indexOf(s, sub string) int {
	return bytes.Index([]byte(s), []byte(sub))
}
```

- [ ] **Step 2: Run to verify fail**

- [ ] **Step 3: Write `internal/commands/top.go`**

```go
package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/ygncode/fortnite-cli/internal/api"
	"github.com/ygncode/fortnite-cli/internal/output"
	"github.com/ygncode/fortnite-cli/internal/timerange"
)

type topFlags struct {
	By       string
	Interval string
	Last     string
	Limit    int
	Scan     int
}

// metricNameMap translates --by values (camelCase as in the API) to path segments.
var metricNameMap = map[string]string{
	"peakCCU":                 "peak-ccu",
	"favorites":               "favorites",
	"minutesPlayed":           "minutes-played",
	"averageMinutesPerPlayer": "average-minutes-per-player",
	"recommendations":         "recommendations",
	"uniquePlayers":           "unique-players",
	"plays":                   "plays",
}

func newTopCmd(gf *GlobalFlags) *cobra.Command {
	var f topFlags
	cmd := &cobra.Command{
		Use:   "top",
		Short: "Top islands by a metric. Expensive — scans up to --scan islands (default 500).",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := withSignalContext()
			defer cancel()
			client := newClient(gf)
			if err := runTop(ctx, client, os.Stdout, gf, f); err != nil {
				printError(err)
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&f.By, "by", "peakCCU", "metric: peakCCU | plays | uniquePlayers | minutesPlayed | favorites | recommendations | averageMinutesPerPlayer")
	cmd.Flags().StringVar(&f.Interval, "interval", "hour", "day | hour | minute")
	cmd.Flags().StringVar(&f.Last, "last", "24h", "shorthand time range, e.g. 24h, 7d")
	cmd.Flags().IntVar(&f.Limit, "limit", 10, "max results to return")
	cmd.Flags().IntVar(&f.Scan, "scan", 500, "max islands to scan (1-1000); see docs")
	return cmd
}

type topItem struct {
	Code  string  `json:"code"`
	Title string  `json:"title"`
	Value float64 `json:"value"`
}

func runTop(ctx context.Context, client *api.Client, stdout io.Writer, gf *GlobalFlags, f topFlags) error {
	metricPath, ok := metricNameMap[f.By]
	if !ok {
		return fmt.Errorf("unknown --by %q", f.By)
	}
	if f.Scan < 1 || f.Scan > 1000 {
		return fmt.Errorf("--scan must be in [1,1000]")
	}
	if f.Limit < 1 {
		return fmt.Errorf("--limit must be >= 1")
	}
	r, _, err := timerange.Parse(timerange.Input{Last: f.Last, Interval: f.Interval}, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("time range: %w", err)
	}

	// 1) Paginate /islands up to f.Scan records.
	islands, err := scanIslands(ctx, client, f.Scan)
	if err != nil {
		return err
	}

	// 2) Fetch the chosen metric for each island in parallel.
	results := make([]topItem, 0, len(islands))
	var mu sync.Mutex
	sem := make(chan struct{}, gf.Concurrency)
	var wg sync.WaitGroup
	var firstErr error
	for _, island := range islands {
		wg.Add(1)
		sem <- struct{}{}
		go func(it api.IslandListItem) {
			defer wg.Done()
			defer func() { <-sem }()
			resp, mErr := client.GetIslandMetric(ctx, it.Code, f.Interval, metricPath, api.MetricsParams{From: r.From, To: r.To})
			if mErr != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = mErr
				}
				mu.Unlock()
				return
			}
			best, ok := maxNonNullValue(resp.Intervals)
			if !ok {
				return
			}
			mu.Lock()
			results = append(results, topItem{Code: it.Code, Title: it.Title, Value: best})
			mu.Unlock()
		}(island)
	}
	wg.Wait()
	if firstErr != nil {
		return firstErr
	}

	sort.Slice(results, func(i, j int) bool { return results[i].Value > results[j].Value })
	if len(results) > f.Limit {
		results = results[:f.Limit]
	}
	data, err := json.Marshal(results)
	if err != nil {
		return err
	}
	w := &output.Writer{Out: stdout, Format: formatFrom(gf)}
	return w.WriteRaw(data)
}

// scanIslands paginates /islands accumulating up to limit items.
func scanIslands(ctx context.Context, client *api.Client, limit int) ([]api.IslandListItem, error) {
	var all []api.IslandListItem
	cursor := ""
	for len(all) < limit {
		batch := 100
		if remaining := limit - len(all); remaining < batch {
			batch = remaining
		}
		page, err := client.ListIslands(ctx, api.ListIslandsParams{Size: batch, After: cursor})
		if err != nil {
			return nil, err
		}
		all = append(all, page.Data...)
		if page.Meta.Page.NextCursor == "" {
			break
		}
		cursor = page.Meta.Page.NextCursor
	}
	return all, nil
}

// maxNonNullValue returns the largest non-null value in a metric series.
func maxNonNullValue(series []api.MetricValue) (float64, bool) {
	var best float64
	found := false
	for _, v := range series {
		if v.Value == nil {
			continue
		}
		if !found || *v.Value > best {
			best = *v.Value
			found = true
		}
	}
	return best, found
}
```

- [ ] **Step 4: Remove `newTopCmd` stub**

- [ ] **Step 5: Run tests**

```bash
go test ./...
```

- [ ] **Step 6: Commit**

```bash
git add internal/commands/
git commit -m "feat(cmd): implement top composite with bounded concurrency"
```

---

## Task 18: Composite — `compare`

**Files:**
- Create: `internal/commands/compare.go`
- Modify: `internal/commands/stubs.go`
- Create: `internal/commands/compare_test.go`

- [ ] **Step 1: Write failing test**

```go
func TestCompareAlignsSeries(t *testing.T) {
	metricA := `{"intervals": [
      {"value": 10, "timestamp": "2026-04-14T00:00:00.000Z"},
      {"value": 20, "timestamp": "2026-04-15T00:00:00.000Z"}
    ]}`
	metricB := `{"intervals": [
      {"value": 30, "timestamp": "2026-04-14T00:00:00.000Z"},
      {"value": null, "timestamp": "2026-04-15T00:00:00.000Z"}
    ]}`
	srv := fortniteServer(t, map[string]string{
		"GET /islands/A":                      `{"code":"A","title":"Alpha","tags":[]}`,
		"GET /islands/B":                      `{"code":"B","title":"Bravo","tags":[]}`,
		"GET /islands/A/metrics/day/plays":    metricA,
		"GET /islands/B/metrics/day/plays":    metricB,
	})
	defer srv.Close()

	gf := &GlobalFlags{Format: "json", Concurrency: 2}
	client := api.DefaultClient(0)
	client.BaseURL = srv.URL

	var stdout bytes.Buffer
	err := runCompare(context.Background(), client, &stdout, gf, compareFlags{
		Codes:    []string{"A", "B"},
		Metric:   "plays",
		Interval: "day",
		Last:     "2d",
	})
	require.NoError(t, err)
	require.Contains(t, stdout.String(), `"timestamps"`)
	require.Contains(t, stdout.String(), `"series"`)
	require.Contains(t, stdout.String(), `"Alpha"`)
	require.Contains(t, stdout.String(), `"Bravo"`)
}
```

- [ ] **Step 2: Run to verify fail**

- [ ] **Step 3: Write `internal/commands/compare.go`**

```go
package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/ygncode/fortnite-cli/internal/api"
	"github.com/ygncode/fortnite-cli/internal/output"
	"github.com/ygncode/fortnite-cli/internal/timerange"
)

type compareFlags struct {
	Codes    []string
	Metric   string
	Interval string
	Last     string
	From     string
	To       string
}

func newCompareCmd(gf *GlobalFlags) *cobra.Command {
	var f compareFlags
	cmd := &cobra.Command{
		Use:   "compare <code1> <code2> [...]",
		Short: "Compare two or more islands on a single metric.",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			f.Codes = args
			ctx, cancel := withSignalContext()
			defer cancel()
			client := newClient(gf)
			if err := runCompare(ctx, client, os.Stdout, gf, f); err != nil {
				printError(err)
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&f.Metric, "metric", "peakCCU", "metric to compare (same values as top --by)")
	cmd.Flags().StringVar(&f.Interval, "interval", "hour", "day | hour | minute")
	cmd.Flags().StringVar(&f.Last, "last", "24h", "shorthand time range")
	cmd.Flags().StringVar(&f.From, "from", "", "start of range (RFC3339)")
	cmd.Flags().StringVar(&f.To, "to", "", "end of range (RFC3339)")
	return cmd
}

type compareSeries struct {
	Code   string     `json:"code"`
	Title  string     `json:"title"`
	Values []*float64 `json:"values"`
}

type compareOutput struct {
	Timestamps []string        `json:"timestamps"`
	Metric     string          `json:"metric"`
	Interval   string          `json:"interval"`
	Series     []compareSeries `json:"series"`
}

func runCompare(ctx context.Context, client *api.Client, stdout io.Writer, gf *GlobalFlags, f compareFlags) error {
	metricPath, ok := metricNameMap[f.Metric]
	if !ok {
		return fmt.Errorf("unknown --metric %q", f.Metric)
	}
	r, _, err := timerange.Parse(timerange.Input{
		Last: f.Last, From: f.From, To: f.To, Interval: f.Interval,
	}, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("time range: %w", err)
	}

	type fetched struct {
		idx    int
		title  string
		values []api.MetricValue
		err    error
	}
	out := make([]fetched, len(f.Codes))
	var wg sync.WaitGroup
	sem := make(chan struct{}, gf.Concurrency)
	for i, code := range f.Codes {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int, code string) {
			defer wg.Done()
			defer func() { <-sem }()
			meta, err := client.GetIsland(ctx, code)
			if err != nil {
				out[i] = fetched{idx: i, err: err}
				return
			}
			resp, err := client.GetIslandMetric(ctx, code, f.Interval, metricPath, api.MetricsParams{From: r.From, To: r.To})
			if err != nil {
				out[i] = fetched{idx: i, err: err}
				return
			}
			out[i] = fetched{idx: i, title: meta.Title, values: resp.Intervals}
		}(i, code)
	}
	wg.Wait()
	for _, res := range out {
		if res.err != nil {
			return res.err
		}
	}

	// Build a shared timestamp axis from the first non-empty series.
	var axis []string
	for _, res := range out {
		if len(res.values) > 0 {
			axis = make([]string, len(res.values))
			for i, v := range res.values {
				axis[i] = v.Timestamp
			}
			break
		}
	}

	result := compareOutput{
		Timestamps: axis,
		Metric:     f.Metric,
		Interval:   f.Interval,
		Series:     make([]compareSeries, 0, len(out)),
	}
	for i, res := range out {
		vals := make([]*float64, len(axis))
		for j, v := range res.values {
			if j < len(vals) {
				vals[j] = v.Value
			}
		}
		result.Series = append(result.Series, compareSeries{
			Code:   f.Codes[i],
			Title:  res.title,
			Values: vals,
		})
	}

	data, err := json.Marshal(result)
	if err != nil {
		return err
	}
	w := &output.Writer{Out: stdout, Format: formatFrom(gf)}
	return w.WriteRaw(data)
}
```

- [ ] **Step 4: Remove `newCompareCmd` stub**

- [ ] **Step 5: Run tests**

```bash
go test ./...
```

- [ ] **Step 6: Commit**

```bash
git add internal/commands/
git commit -m "feat(cmd): implement compare composite"
```

---

## Task 19: Composite — `search`

**Files:**
- Create: `internal/commands/search.go`
- Modify: `internal/commands/stubs.go`
- Create: `internal/commands/search_test.go`

- [ ] **Step 1: Write failing test**

```go
func TestSearchFiltersByTagAndCreator(t *testing.T) {
	listBody := `{
      "links": {"prev": null, "next": null},
      "meta": {"count": 3, "page": {"prevCursor": "", "nextCursor": ""}},
      "data": [
        {"code": "A", "creatorCode": "alice", "title": "Race Arena", "createdIn": "UEFN", "tags": ["race", "casual"], "meta": {"page": {"cursor": "a"}}},
        {"code": "B", "creatorCode": "bob", "title": "Build Battle", "createdIn": "UEFN", "tags": ["build"], "meta": {"page": {"cursor": "b"}}},
        {"code": "C", "creatorCode": "alice", "title": "Zombie Mode", "createdIn": "UEFN", "tags": ["race", "horror"], "meta": {"page": {"cursor": "c"}}}
      ]
    }`
	srv := fortniteServer(t, map[string]string{"GET /islands": listBody})
	defer srv.Close()

	gf := &GlobalFlags{Format: "json"}
	client := api.DefaultClient(0)
	client.BaseURL = srv.URL

	var stdout bytes.Buffer
	err := runSearch(context.Background(), client, &stdout, gf, searchFlags{
		Tags: []string{"race"}, Creator: "alice", MaxScan: 100,
	})
	require.NoError(t, err)
	require.Contains(t, stdout.String(), `"A"`)
	require.Contains(t, stdout.String(), `"C"`)
	require.NotContains(t, stdout.String(), `"B"`)
}
```

- [ ] **Step 2: Run to verify fail**

- [ ] **Step 3: Write `internal/commands/search.go`**

```go
package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ygncode/fortnite-cli/internal/api"
	"github.com/ygncode/fortnite-cli/internal/output"
)

type searchFlags struct {
	Tags          []string
	Creator       string
	TitleContains string
	MaxScan       int
}

func newSearchCmd(gf *GlobalFlags) *cobra.Command {
	var f searchFlags
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Find islands matching tag/creator/title filters.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := withSignalContext()
			defer cancel()
			client := newClient(gf)
			if err := runSearch(ctx, client, os.Stdout, gf, f); err != nil {
				printError(err)
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringSliceVar(&f.Tags, "tag", nil, "filter by tag (repeatable; island must contain all)")
	cmd.Flags().StringVar(&f.Creator, "creator", "", "filter by creatorCode (exact match)")
	cmd.Flags().StringVar(&f.TitleContains, "title-contains", "", "filter by case-insensitive substring of title")
	cmd.Flags().IntVar(&f.MaxScan, "max-scan", 500, "max islands to scan (1-1000)")
	return cmd
}

type searchResult struct {
	Data []api.IslandMetadata `json:"data"`
	Meta struct {
		Scanned   int  `json:"scanned"`
		Matched   int  `json:"matched"`
		Truncated bool `json:"truncated"`
	} `json:"meta"`
}

func runSearch(ctx context.Context, client *api.Client, stdout io.Writer, gf *GlobalFlags, f searchFlags) error {
	if f.MaxScan < 1 || f.MaxScan > 1000 {
		return fmt.Errorf("--max-scan must be in [1,1000]")
	}

	islands, err := scanIslands(ctx, client, f.MaxScan)
	if err != nil {
		return err
	}

	result := searchResult{}
	result.Meta.Scanned = len(islands)

	for _, it := range islands {
		if !matchesFilters(it, f) {
			continue
		}
		result.Data = append(result.Data, it.IslandMetadata)
	}
	result.Meta.Matched = len(result.Data)
	result.Meta.Truncated = len(islands) == f.MaxScan

	data, err := json.Marshal(result)
	if err != nil {
		return err
	}
	w := &output.Writer{Out: stdout, Format: formatFrom(gf)}
	return w.WriteRaw(data)
}

func matchesFilters(it api.IslandListItem, f searchFlags) bool {
	if f.Creator != "" && it.CreatorCode != f.Creator {
		return false
	}
	if f.TitleContains != "" && !strings.Contains(strings.ToLower(it.Title), strings.ToLower(f.TitleContains)) {
		return false
	}
	for _, want := range f.Tags {
		found := false
		for _, got := range it.Tags {
			if got == want {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
```

- [ ] **Step 4: Remove `newSearchCmd` stub and delete `stubs.go`**

After this task, `stubs.go` should contain no declarations (all stubs replaced by real implementations in Tasks 11–19, and the `singleMetricSpec` type was moved in Task 14). Delete the file:

```bash
rm internal/commands/stubs.go
go build ./...
```

Expected: build succeeds.

- [ ] **Step 5: Run tests**

```bash
go test ./...
```

Expected: all tests PASS. All 4 composite verbs plus all 12 thin 1:1 commands now exist.

- [ ] **Step 6: Smoke test all composites**

```bash
go build ./...
./fortnite top --by peakCCU --last 1h --limit 5 --scan 50
./fortnite summary 7022-1566-6527
./fortnite compare 7022-1566-6527 8495-3177-6945 --metric peakCCU --last 1h
./fortnite search --tag race --max-scan 100
```

All should exit 0 and emit JSON.

- [ ] **Step 7: Commit**

```bash
git add internal/commands/
git commit -m "feat(cmd): implement search composite"
```

---

## Task 20: Skill — `SKILL.md`

**Files:**
- Create: `skill/fortnite-islands/SKILL.md`

- [ ] **Step 1: Write `skill/fortnite-islands/SKILL.md`**

```markdown
---
name: fortnite-islands
description: Query Fortnite island metadata and engagement metrics (peak CCU, plays, unique players, retention, favorites, minutes played) via the official Fortnite Ecosystem API. Use when the user asks about Fortnite Creative/UEFN islands, island popularity, player counts, play time, retention, or wants to compare or rank islands by any engagement metric. Covers only island data — not player stats, matches, cosmetics, or shop.
license: MIT
compatibility: Requires the `fortnite` CLI binary (install instructions in the repo README). Network access to api.fortnite.com.
metadata:
  version: "0.1.0"
  repository: https://github.com/ygncode/fortnite-cli
  minimum-cli-version: "0.1.0"
---

# Fortnite Islands

Query the official Fortnite Ecosystem v1 Data API: island metadata and engagement metrics (peak CCU, plays, unique players, retention, favorites, minutes played). The API's historical window is capped at **7 days**. The `null` metric value means "fewer than 5 unique players in that bucket" — it is not an error. Treat it as data.

## Preflight

Before the first call in a session, verify the CLI is installed:

```bash
fortnite --version
```

If the command is not found, the user needs to install it. Point them at the repository README's install instructions and stop.

## Prefer composite commands

When you can answer the user in one command, use these. Each one is internally doing multiple API calls, so picking the right composite saves many turns and tokens.

| User intent | Command |
| --- | --- |
| "Top N islands by <metric>" | `fortnite top --by <metric> --last <duration> --limit <N>` |
| Full picture of one island | `fortnite summary <code>` |
| Compare specific islands on one metric | `fortnite compare <code1> <code2> [...] --metric <m> --last <duration>` |
| Find islands by tag / creator / title | `fortnite search --tag <t> --creator <c> --title-contains <s>` |
| Otherwise (you want a specific endpoint) | Drop to the 1:1 layer below. |

## 1:1 endpoint cheat sheet

```
fortnite islands list [--size N] [--after CURSOR] [--before CURSOR]
fortnite islands get <code>
fortnite islands metrics <code> [--interval day|hour|minute] [--metric M]... [--last DUR | --from T --to T]
fortnite islands peak-ccu <code> --interval day|hour|minute [--last DUR]
fortnite islands favorites <code> --interval day|hour|minute [--last DUR]
fortnite islands minutes-played <code> --interval day|hour|minute [--last DUR]
fortnite islands avg-minutes <code> --interval day [--last DUR]         # day-only
fortnite islands recommendations <code> --interval day|hour|minute [--last DUR]
fortnite islands unique-players <code> --interval day|hour|minute [--last DUR]
fortnite islands plays <code> --interval day|hour|minute [--last DUR]
fortnite islands retention <code> [--last DUR]                          # day-only
```

Full flags, defaults, and response examples: `references/COMMANDS.md`.

## Output shape and parsing

- Every command returns JSON on stdout (exit 0) or a structured JSON error on stderr (exit 2 for API errors, 3 for network errors, 1 for flag errors).
- Default output is *slim JSON*: pagination envelopes are flattened, timestamps are rounded to the interval granularity (`"2026-04-14"` for day, `"2026-04-14T07:00"` for hour/minute), and metric arrays containing only `null` values collapse to a literal `null`.
- Use `--raw` if you need the exact API shape. You rarely do.
- Error shape: `{"error": "rate_limited|not_found|bad_request|upstream_error|api_error|generic", "message": "...", "status": 429}`.

Example `fortnite top --by peakCCU --last 1h --limit 3` output:

```json
[
  {"code":"1234-1234-1234","title":"Battle Royale","value":65432},
  {"code":"5555-5555-5555","title":"Fall Guys","value":45000},
  {"code":"9999-9999-9999","title":"OG","value":33000}
]
```

Example `fortnite islands peak-ccu <code> --interval hour --last 3h` output:

```json
{
  "intervals": [
    {"value": 120, "timestamp": "2026-04-14T07:00"},
    {"value": 95, "timestamp": "2026-04-14T08:00"},
    {"value": null, "timestamp": "2026-04-14T09:00"}
  ]
}
```

## Token-efficiency rules

1. **Don't call `top` when you already have island codes.** `top` scans up to 500 islands by default and fires one metric call per island. It is the most expensive command. If the user names specific islands, use `summary` or `compare` instead.
2. **Don't iterate `islands list` in a loop.** Use `search` with `--max-scan` to filter client-side in a single call. The API has no server-side search.
3. **Parse the JSON once.** Don't ask the CLI to re-run with different flags just to reformat output; adjust your parsing.
4. **Use `--last` over `--from`/`--to`** when the user is asking about "the last N hours/days" — it's shorter and less error-prone.
5. **Check exit code.** Exit 2 means the API responded with an error (rate-limit, not-found, etc.). The JSON on stderr tells you which.

## Caveats

- `null` in a metric value means "fewer than 5 unique players in that interval." It is normal, not an error. Surface it to the user as "insufficient data," not as a failure.
- The API's historical window is 7 days. Anything older returns an error.
- `avg-minutes` (average minutes per player) is only available at `day` interval.
- `retention` (d1/d7) is only available at `day` interval.
- `top` scans the newest-released-first island list, not a popularity-ranked list. The first 100 islands will skew toward recent releases. Default scan depth is 500 to mitigate this. If you truly need a global top-N, pass `--scan 1000` and expect it to take tens of seconds.

## References

- `references/COMMANDS.md` — complete per-command reference with all flags, defaults, and example output.
- `references/RECIPES.md` — copy-paste command snippets for common user tasks.
```

- [ ] **Step 2: Commit**

```bash
git add skill/fortnite-islands/SKILL.md
git commit -m "feat(skill): add SKILL.md"
```

---

## Task 21: Skill — `references/COMMANDS.md`

**Files:**
- Create: `skill/fortnite-islands/references/COMMANDS.md`

- [ ] **Step 1: Write the file**

```markdown
# fortnite command reference

Complete per-command reference. This file is loaded on demand — do not load it for every call.

## Global flags (available on every command)

- `--raw` — emit the raw API response instead of the slimmed form.
- `--format json|ndjson` — default `json`.
- `--no-retry` — disable 429/5xx exponential backoff.
- `--timeout DURATION` — HTTP timeout. Default `30s`. Go duration syntax.
- `--concurrency N` — worker-pool size for `top` / `compare`. Default `8`.

## Exit codes

- `0` success
- `1` flag validation error (printed to stderr)
- `2` API error (4xx/5xx after retries; JSON on stderr)
- `3` network or IO error

## `fortnite islands list`

List islands (cursor-paginated, newest released first).

```
fortnite islands list [--size N] [--after CURSOR] [--before CURSOR]
```

Flags:
- `--size N` — default 100, max 1000.
- `--after CURSOR` / `--before CURSOR` — from a previous response's `_pagination.next` / `_pagination.prev`.

Example output (slimmed):
```json
{
  "_pagination": {"next": "/ecosystem/v1/islands?after=...&size=100"},
  "data": [
    {"code":"1234-1234-1234","title":"...","creatorCode":"...","createdIn":"UEFN","tags":["race"],"cursor":"..."}
  ]
}
```

## `fortnite islands get <code>`

Metadata for one island.

Example:
```bash
fortnite islands get 7022-1566-6527
```

## `fortnite islands metrics <code>`

Without `--interval`: bundled day-interval metrics (all metrics at once).
With `--interval`: filterable, supports `--metric M` (repeatable).

```
fortnite islands metrics <code>
fortnite islands metrics <code> --interval day --metric peakCCU --metric plays
fortnite islands metrics <code> --interval hour --last 6h
```

Flags:
- `--interval day|hour|minute`
- `--metric M` — repeatable. Valid: peakCCU, plays, uniquePlayers, minutesPlayed, favorites, recommendations, averageMinutesPerPlayer, retention.
- `--last DURATION` — e.g. `24h`, `7d`, `60m`.
- `--from RFC3339` / `--to RFC3339`.

## `fortnite islands peak-ccu <code>`
`fortnite islands favorites <code>`
`fortnite islands minutes-played <code>`
`fortnite islands recommendations <code>`
`fortnite islands unique-players <code>`
`fortnite islands plays <code>`

Single-metric endpoints. Same flag shape:

```
fortnite islands peak-ccu <code> --interval day|hour|minute [--last DUR | --from T --to T]
```

Default `--interval` is `day`.

Example output:
```json
{
  "intervals": [
    {"value": 120, "timestamp": "2026-04-14T07:00"},
    {"value": null, "timestamp": "2026-04-14T08:00"}
  ]
}
```

## `fortnite islands avg-minutes <code>`

Average minutes per player. **Day-only.** Passing `--interval hour` or `--interval minute` exits 1.

## `fortnite islands retention <code>`

D1 / D7 retention. **Day-only.** Example output:

```json
{
  "intervals": [
    {"d1": 80, "d7": 45, "timestamp": "2026-04-14"},
    {"d1": null, "d7": null, "timestamp": "2026-04-15"}
  ]
}
```

## `fortnite top`

Top islands by a metric. **Expensive.** Scans up to `--scan` islands (default 500, max 1000), making one list pagination call per 100 plus one metric call per island. With default concurrency 8, scanning 500 islands takes ~30–60 seconds.

```
fortnite top --by peakCCU --interval hour --last 24h --limit 10 --scan 500
```

Valid `--by` values: `peakCCU`, `plays`, `uniquePlayers`, `minutesPlayed`, `favorites`, `recommendations`, `averageMinutesPerPlayer`.

Output:
```json
[
  {"code":"1234-1234-1234","title":"...","value":65432},
  ...
]
```

Islands whose metric series is entirely null are dropped.

## `fortnite summary <code>`

Metadata + bundled metrics in one call.

```
fortnite summary <code> [--last DUR]
```

Output:
```json
{
  "metadata": {"code":"...","title":"...","tags":[...]},
  "metrics": {"peakCCU":[...], "plays":[...], "uniquePlayers":[...], ...}
}
```

## `fortnite compare`

Compare 2+ islands on one metric.

```
fortnite compare <code1> <code2> [...] --metric peakCCU --interval hour --last 24h
```

Output:
```json
{
  "timestamps": ["2026-04-14T07:00:00.000Z", "..."],
  "metric": "peakCCU",
  "interval": "hour",
  "series": [
    {"code":"A","title":"Alpha","values":[10,20,null]},
    {"code":"B","title":"Bravo","values":[30,null,50]}
  ]
}
```

`timestamps` are the ISO8601 strings from the first non-empty series (not rounded — this is an exception to the slim rule because the axis is used for alignment).

## `fortnite search`

Client-side filter over paginated `islands list`. No server-side search on the API.

```
fortnite search [--tag T]... [--creator C] [--title-contains S] [--max-scan N]
```

Flags:
- `--tag T` — repeatable. Island must contain ALL specified tags.
- `--creator C` — exact `creatorCode` match.
- `--title-contains S` — case-insensitive substring of `title`.
- `--max-scan N` — default 500, max 1000. `meta.truncated=true` in the output means you hit the cap.

Output:
```json
{
  "data": [{"code":"...","title":"...","tags":[...]}],
  "meta": {"scanned": 500, "matched": 12, "truncated": false}
}
```
```

- [ ] **Step 2: Commit**

```bash
git add skill/fortnite-islands/references/COMMANDS.md
git commit -m "feat(skill): add per-command reference"
```

---

## Task 22: Skill — `references/RECIPES.md`

**Files:**
- Create: `skill/fortnite-islands/references/RECIPES.md`

- [ ] **Step 1: Write the file**

```markdown
# Recipes

Copy-paste command snippets for common tasks. Each recipe is one command that returns JSON the agent can parse directly.

## Top 10 most played islands in the last 24 hours

```bash
fortnite top --by plays --interval hour --last 24h --limit 10
```

Returns: `[{"code":"...","title":"...","value":<plays>}, ...]` sorted descending.

## Top 10 by peak CCU over the last 7 days

```bash
fortnite top --by peakCCU --interval day --last 7d --limit 10
```

## Full picture of one island

```bash
fortnite summary 7022-1566-6527
```

Returns `{metadata, metrics}` with all 8 metric series plus retention.

## How has island X trended over 7 days?

```bash
fortnite islands plays 7022-1566-6527 --interval day --last 7d
```

Returns `{intervals: [{value, timestamp}, ...]}`. Walk the array from oldest to newest to describe the trend.

## Does island X retain players (d1/d7)?

```bash
fortnite islands retention 7022-1566-6527 --last 7d
```

## Compare peak CCU for three islands over the last day

```bash
fortnite compare 7022-1566-6527 8495-3177-6945 1234-1234-1234 --metric peakCCU --interval hour --last 24h
```

Returns aligned `{timestamps, series: [{code, title, values}]}`.

## Find race-genre islands

```bash
fortnite search --tag race --max-scan 500
```

`meta.truncated` tells you if there are more beyond the scan limit.

## Find islands by a specific creator

```bash
fortnite search --creator maphits --max-scan 500
```

## Find islands whose title contains a keyword

```bash
fortnite search --title-contains zombie --max-scan 500
```

## Latest 20 published islands

```bash
fortnite islands list --size 20
```

The API sorts newest-released-first, so this is the freshest 20.

## Pull raw API data for one island (debugging)

```bash
fortnite islands metrics 7022-1566-6527 --raw
```

## Handle the "<5 players" null case

When any metric value is `null`, it means the interval had fewer than 5 unique players. Surface it to the user as "insufficient data" rather than "error". Example:

```bash
fortnite islands peak-ccu 7022-1566-6527 --interval minute --last 60m
```

If the entire `intervals` array has `null` values, say: "Island 7022-1566-6527 had fewer than 5 players in every 10-minute bucket over the last hour."
```

- [ ] **Step 2: Commit**

```bash
git add skill/fortnite-islands/references/RECIPES.md
git commit -m "feat(skill): add recipe reference"
```

---

## Task 23: README

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Replace `README.md` stub with the full version**

```markdown
# fortnite-cli

Agent-friendly Go CLI for the official Fortnite Ecosystem v1 Data API, plus a companion [agent skill](https://agentskills.io) (`fortnite-islands`).

## What it does

Queries Fortnite island (Creative / UEFN) metadata and engagement metrics: peak CCU, plays, unique players, retention, favorites, minutes played, recommendations, average minutes per player. It does **not** cover player stats, matches, cosmetics, shop, or BR stats — those are not part of the Ecosystem API.

## Install

### One-liner (macOS / Linux)

```bash
curl -fsSL https://raw.githubusercontent.com/ygncode/fortnite-cli/main/install.sh | sh
```

This drops the `fortnite` binary into `$HOME/.local/bin`. Re-run to upgrade.

### From source

```bash
go install github.com/ygncode/fortnite-cli/cmd/fortnite@latest
```

### Verify

```bash
fortnite --version
```

## Quickstart

```bash
fortnite islands list --size 5
fortnite top --by peakCCU --last 24h --limit 10
fortnite summary 7022-1566-6527
```

Output is compact JSON by default. Pass `--raw` for the raw API response.

## Using the agent skill

Copy `skill/fortnite-islands/` into your agent's skills directory. Installation paths vary by agent — see your agent's docs for where skills live. After that, ask your agent things like "What are the top 10 Fortnite islands by peak CCU right now?" and it will use the CLI.

## Command reference

| Command | Description |
| --- | --- |
| `fortnite islands list` | List islands (paginated). |
| `fortnite islands get <code>` | Island metadata. |
| `fortnite islands metrics <code>` | Bundled metrics (day) or filterable with `--interval`. |
| `fortnite islands peak-ccu <code>` | Peak concurrent users. |
| `fortnite islands favorites <code>` | Favorites count. |
| `fortnite islands minutes-played <code>` | Total minutes played. |
| `fortnite islands avg-minutes <code>` | Average minutes per player (day only). |
| `fortnite islands recommendations <code>` | Recommendation count. |
| `fortnite islands unique-players <code>` | Unique players. |
| `fortnite islands plays <code>` | Play count. |
| `fortnite islands retention <code>` | D1/D7 retention (day only). |
| `fortnite top --by <metric>` | Top islands by metric. |
| `fortnite summary <code>` | One-shot metadata + metrics. |
| `fortnite compare <code>...` | Compare islands on one metric. |
| `fortnite search` | Client-side filter by tag/creator/title. |

Run any command with `--help` for flag details.

## Design

See `docs/superpowers/specs/2026-04-15-fortnite-cli-and-skill-design.md` for the full design spec.

## License

MIT.
```

- [ ] **Step 2: Commit**

```bash
git add README.md
git commit -m "docs: expand README with install and quickstart"
```

---

## Task 24: `install.sh`

**Files:**
- Create: `install.sh`

- [ ] **Step 1: Write `install.sh`**

```bash
#!/bin/sh
# Install the latest fortnite-cli release into $HOME/.local/bin.
# Re-run to upgrade.

set -eu

REPO="ygncode/fortnite-cli"
BIN="fortnite"
INSTALL_DIR="${FORTNITE_INSTALL_DIR:-$HOME/.local/bin}"

os=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$os" in
  darwin) os=darwin ;;
  linux)  os=linux ;;
  mingw*|msys*|cygwin*) os=windows ;;
  *) echo "unsupported OS: $os" >&2; exit 1 ;;
esac

arch=$(uname -m)
case "$arch" in
  x86_64|amd64) arch=amd64 ;;
  aarch64|arm64) arch=arm64 ;;
  armv7l) arch=armv7 ;;
  *) echo "unsupported arch: $arch" >&2; exit 1 ;;
esac

ext="tar.gz"
if [ "$os" = "windows" ]; then ext="zip"; fi

# Resolve latest tag.
if command -v curl >/dev/null 2>&1; then
  fetch() { curl -fsSL "$1"; }
else
  fetch() { wget -qO- "$1"; }
fi
tag=$(fetch "https://api.github.com/repos/$REPO/releases/latest" | grep -o '"tag_name": *"[^"]*"' | head -1 | sed 's/.*"\(.*\)"/\1/')
if [ -z "${tag:-}" ]; then
  echo "failed to resolve latest release tag" >&2
  exit 1
fi

archive="${BIN}_${tag#v}_${os}_${arch}.${ext}"
url="https://github.com/$REPO/releases/download/$tag/$archive"
checksums_url="https://github.com/$REPO/releases/download/$tag/checksums.txt"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

echo "Downloading $url..."
fetch "$url" > "$tmp/$archive"
fetch "$checksums_url" > "$tmp/checksums.txt"

# Verify checksum.
if command -v sha256sum >/dev/null 2>&1; then
  sha256_cmd="sha256sum"
elif command -v shasum >/dev/null 2>&1; then
  sha256_cmd="shasum -a 256"
else
  echo "neither sha256sum nor shasum found; skipping checksum verification" >&2
  sha256_cmd=""
fi
if [ -n "$sha256_cmd" ]; then
  expected=$(grep "  $archive\$" "$tmp/checksums.txt" | awk '{print $1}')
  actual=$(cd "$tmp" && $sha256_cmd "$archive" | awk '{print $1}')
  if [ "$expected" != "$actual" ]; then
    echo "checksum mismatch: expected $expected, got $actual" >&2
    exit 1
  fi
fi

# Extract.
case "$ext" in
  tar.gz) tar -xzf "$tmp/$archive" -C "$tmp" ;;
  zip)    unzip -q "$tmp/$archive" -d "$tmp" ;;
esac

mkdir -p "$INSTALL_DIR"
install -m 0755 "$tmp/$BIN" "$INSTALL_DIR/$BIN" 2>/dev/null || cp "$tmp/$BIN" "$INSTALL_DIR/$BIN"
chmod 0755 "$INSTALL_DIR/$BIN"

echo "Installed $BIN $tag to $INSTALL_DIR/$BIN"
case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *) echo "Note: add $INSTALL_DIR to your PATH to run \`$BIN\`." ;;
esac
```

- [ ] **Step 2: Mark executable and test syntax**

```bash
chmod +x install.sh
sh -n install.sh
```

Expected: syntax check passes silently.

- [ ] **Step 3: Commit**

```bash
git add install.sh
git commit -m "feat: add curl|sh installer"
```

---

## Task 25: GoReleaser config

**Files:**
- Create: `.goreleaser.yaml`

- [ ] **Step 1: Write `.goreleaser.yaml`**

```yaml
version: 2

project_name: fortnite-cli

builds:
  - id: fortnite
    main: ./cmd/fortnite
    binary: fortnite
    env:
      - CGO_ENABLED=0
    goos:
      - darwin
      - linux
      - windows
    goarch:
      - amd64
      - arm64
      - arm
    goarm:
      - "7"
    ignore:
      - goos: windows
        goarch: arm64
      - goos: windows
        goarch: arm
      - goos: darwin
        goarch: arm
    ldflags:
      - -s -w -X main.version={{ .Version }}

archives:
  - id: fortnite
    name_template: >-
      {{ .ProjectName }}_{{ .Version }}_{{ .Os }}_
      {{- if eq .Arch "amd64" }}amd64
      {{- else if eq .Arch "arm64" }}arm64
      {{- else if eq .Arch "arm" }}armv{{ .Arm }}
      {{- else }}{{ .Arch }}{{ end }}
    format: tar.gz
    format_overrides:
      - goos: windows
        format: zip
    files:
      - LICENSE
      - README.md

checksum:
  name_template: "checksums.txt"
  algorithm: sha256

changelog:
  use: git
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"
      - "^chore:"

release:
  draft: true
  prerelease: auto
```

- [ ] **Step 2: Verify locally (optional but recommended)**

```bash
# Dry-run (requires goreleaser installed):
goreleaser check
goreleaser release --snapshot --clean --skip=publish
```

Expected: successful build into `dist/`. Skip this step if goreleaser isn't installed locally — CI will catch config issues.

- [ ] **Step 3: Commit**

```bash
git add .goreleaser.yaml
git commit -m "chore: add goreleaser config"
```

---

## Task 26: CI + Release workflows

**Files:**
- Create: `.github/workflows/ci.yaml`
- Create: `.github/workflows/release.yaml`

- [ ] **Step 1: Write `.github/workflows/ci.yaml`**

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      - name: Vet
        run: go vet ./...
      - name: Test
        run: go test ./...

  e2e:
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      - name: E2E smoke test
        run: go test -tags=e2e ./internal/api/...
```

- [ ] **Step 2: Write `.github/workflows/release.yaml`**

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write

jobs:
  goreleaser:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      - uses: goreleaser/goreleaser-action@v6
        with:
          version: '~> v2'
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

- [ ] **Step 3: Commit**

```bash
git add .github/
git commit -m "ci: add test and release workflows"
```

---

## Task 27: E2E smoke test

**Files:**
- Create: `internal/api/e2e_test.go`

- [ ] **Step 1: Write `internal/api/e2e_test.go` (build-tag gated)**

```go
//go:build e2e

package api

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestE2EListIslands hits the real Fortnite API. Only runs with -tags=e2e.
func TestE2EListIslands(t *testing.T) {
	c := DefaultClient(30 * time.Second)
	resp, err := c.ListIslands(context.Background(), ListIslandsParams{Size: 5})
	require.NoError(t, err)
	require.NotEmpty(t, resp.Data)
	require.NotEmpty(t, resp.Data[0].Code)
	require.NotEmpty(t, resp.Data[0].Title)
}
```

- [ ] **Step 2: Run locally**

```bash
go test -tags=e2e ./internal/api/
```

Expected: PASS. The test hits the real API and returns 5 islands.

- [ ] **Step 3: Commit**

```bash
git add internal/api/e2e_test.go
git commit -m "test(api): add e2e smoke test behind build tag"
```

---

## Task 28: Final verification

- [ ] **Step 1: Run full test suite**

```bash
go test ./...
```

Expected: all tests PASS.

- [ ] **Step 2: Run vet**

```bash
go vet ./...
```

Expected: no warnings.

- [ ] **Step 3: Build a production binary**

```bash
go build -ldflags "-s -w -X main.version=v0.1.0" -o fortnite ./cmd/fortnite
./fortnite --version
```

Expected: prints `fortnite version v0.1.0`.

- [ ] **Step 4: Exercise every command against the real API**

```bash
./fortnite islands list --size 3
code=$(./fortnite islands list --size 1 | python3 -c 'import sys,json; print(json.load(sys.stdin)["data"][0]["code"])')
./fortnite islands get "$code"
./fortnite islands metrics "$code"
./fortnite islands metrics "$code" --interval day --metric peakCCU
./fortnite islands peak-ccu "$code" --interval hour --last 2h
./fortnite islands favorites "$code" --interval day
./fortnite islands minutes-played "$code" --interval day
./fortnite islands avg-minutes "$code"
./fortnite islands recommendations "$code" --interval day
./fortnite islands unique-players "$code" --interval day
./fortnite islands plays "$code" --interval day
./fortnite islands retention "$code"
./fortnite top --by peakCCU --interval hour --last 1h --limit 3 --scan 20
./fortnite summary "$code"
./fortnite compare "$code" "$code" --metric peakCCU --interval hour --last 2h
./fortnite search --tag race --max-scan 50
```

All commands should exit 0 and emit valid JSON. Spot-check that slim output looks reasonable.

- [ ] **Step 5: Validate the skill with `skills-ref` if available**

```bash
# Optional — only if skills-ref is installed locally.
skills-ref validate skill/fortnite-islands
```

Expected: passes frontmatter validation.

- [ ] **Step 6: Tag v0.1.0**

```bash
git tag -a v0.1.0 -m "v0.1.0: initial release"
```

Do **not** push the tag yet. Pushing triggers the release workflow; confirm with the user first.

- [ ] **Step 7: Report completion**

Tell the user:
- All 26 tasks completed, tests green.
- Tagged `v0.1.0` locally but not pushed.
- Next step: push `main` + push `v0.1.0` tag to kick off the GoReleaser release (they will need to create the `ygncode/fortnite-cli` repo on GitHub first and add it as the `origin` remote).
