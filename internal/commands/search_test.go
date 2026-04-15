package commands

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ygncode/fortnite-cli/internal/api"
)

func TestSearchFiltersByTagAndCreator(t *testing.T) {
	listBody := `{
      "links": {"prev": null, "next": null},
      "meta": {"count": 3, "page": {"prevCursor": "", "nextCursor": ""}},
      "data": [
        {"code": "A", "creatorCode": "alice", "title": "Race Arena", "createdIn": "UEFN", "tags": ["race", "casual"], "meta": {"page": {"cursor": "a"}}},
        {"code": "B", "creatorCode": "bob", "title": "Build Battle", "createdIn": "UEFN", "tags": ["build"], "meta": {"page": {"cursor": "b"}}},
        {"code": "C", "creatorCode": "alice", "title": "Zombie Mode", "createdIn": "UEFN", "tags": ["race", "horror"], "meta": {"page": {"cursor": "c"}}}
      ]
    }`
	srv := fortniteServer(t, map[string]string{"GET /islands": listBody})
	defer srv.Close()

	gf := &GlobalFlags{Format: "json"}
	client := api.DefaultClient(0)
	client.BaseURL = srv.URL

	var stdout bytes.Buffer
	err := runSearch(context.Background(), client, &stdout, gf, searchFlags{
		Tags: []string{"race"}, Creator: "alice", MaxScan: 100,
	})
	require.NoError(t, err)
	out := stdout.String()
	require.Contains(t, out, `"A"`)
	require.Contains(t, out, `"C"`)
	require.NotContains(t, out, `"B"`)
}
