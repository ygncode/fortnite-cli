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
