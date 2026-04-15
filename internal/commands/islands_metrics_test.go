package commands

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ygncode/fortnite-cli/internal/api"
)

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
