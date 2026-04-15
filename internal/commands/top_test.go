package commands

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ygncode/fortnite-cli/internal/api"
)

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
		"GET /islands":                         listBody,
		"GET /islands/A/metrics/hour/peak-ccu": metricA,
		"GET /islands/B/metrics/hour/peak-ccu": metricB,
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
	out := stdout.String()
	bIdx := bytes.Index([]byte(out), []byte(`"B"`))
	aIdx := bytes.Index([]byte(out), []byte(`"A"`))
	require.True(t, bIdx > 0 && aIdx > 0 && bIdx < aIdx, "B (200) should appear before A (100) in %q", out)
}
