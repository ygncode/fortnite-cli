package commands

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ygncode/fortnite-cli/internal/api"
)

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
		"GET /islands/A":                   `{"code":"A","title":"Alpha","tags":[]}`,
		"GET /islands/B":                   `{"code":"B","title":"Bravo","tags":[]}`,
		"GET /islands/A/metrics/day/plays": metricA,
		"GET /islands/B/metrics/day/plays": metricB,
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
