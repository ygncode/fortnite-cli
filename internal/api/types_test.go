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
