package commands

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ygncode/fortnite-cli/internal/api"
)

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
