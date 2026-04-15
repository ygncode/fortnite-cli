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

func TestSlimBundledMetricsDayInterval(t *testing.T) {
	raw := readFile(t, "testdata/raw/metrics_bundled.json")
	expected := readFile(t, "testdata/slim/metrics_bundled.json")

	got, err := BundledMetrics(raw, "day")
	require.NoError(t, err)
	assertJSONEqual(t, expected, got)
}

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

func TestRoundTimestamp(t *testing.T) {
	require.Equal(t, "2026-04-14", roundTimestamp("2026-04-14T00:00:00.000Z", "day"))
	require.Equal(t, "2026-04-14T07:00", roundTimestamp("2026-04-14T07:00:00.000Z", "hour"))
	require.Equal(t, "2026-04-14T07:20", roundTimestamp("2026-04-14T07:20:00.000Z", "minute"))
	require.Equal(t, "garbage", roundTimestamp("garbage", "day"))
}
