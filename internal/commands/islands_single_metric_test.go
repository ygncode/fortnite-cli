package commands

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ygncode/fortnite-cli/internal/api"
)

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
	require.Contains(t, stdout.String(), `"2026-04-14T07:00"`)
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
