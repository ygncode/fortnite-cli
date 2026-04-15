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
