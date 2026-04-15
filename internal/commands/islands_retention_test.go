package commands

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ygncode/fortnite-cli/internal/api"
)

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
