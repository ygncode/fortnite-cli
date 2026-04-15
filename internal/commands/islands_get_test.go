package commands

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ygncode/fortnite-cli/internal/api"
)

func TestIslandsGet(t *testing.T) {
	body := string(readRepoFile(t, "testdata/raw/island_get.json"))
	srv := fortniteServer(t, map[string]string{
		"GET /islands/7022-1566-6527": body,
	})
	defer srv.Close()

	gf := &GlobalFlags{Format: "json"}
	client := api.DefaultClient(0)
	client.BaseURL = srv.URL

	var stdout bytes.Buffer
	err := runIslandsGet(context.Background(), client, &stdout, gf, "7022-1566-6527")
	require.NoError(t, err)
	require.Contains(t, stdout.String(), `"title":"BRAINROT RACE"`)
}
