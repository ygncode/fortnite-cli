package commands

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ygncode/fortnite-cli/internal/api"
)

// fortniteServer spins up an httptest.Server backed by testdata files keyed by method+path.
func fortniteServer(t *testing.T, routes map[string]string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Method + " " + r.URL.Path
		body, ok := routes[key]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
}

// readRepoFile reads a file relative to the repo root (two levels up from internal/commands).
func readRepoFile(t *testing.T, rel string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", rel))
	require.NoError(t, err)
	return data
}

func TestIslandsListSlimmed(t *testing.T) {
	body := string(readRepoFile(t, "testdata/raw/islands_list.json"))
	srv := fortniteServer(t, map[string]string{"GET /islands": body})
	defer srv.Close()

	gf := &GlobalFlags{Format: "json"}
	client := api.DefaultClient(0)
	client.BaseURL = srv.URL

	var stdout bytes.Buffer
	err := runIslandsList(context.Background(), client, &stdout, gf, listFlags{Size: 2})
	require.NoError(t, err)
	require.Contains(t, stdout.String(), `"7022-1566-6527"`)
	require.Contains(t, stdout.String(), `"_pagination"`)
	require.NotContains(t, stdout.String(), `"meta":{"page"`)
}

func TestIslandsListRawPassesThrough(t *testing.T) {
	body := string(readRepoFile(t, "testdata/raw/islands_list.json"))
	srv := fortniteServer(t, map[string]string{"GET /islands": body})
	defer srv.Close()

	gf := &GlobalFlags{Raw: true, Format: "json"}
	client := api.DefaultClient(0)
	client.BaseURL = srv.URL

	var stdout bytes.Buffer
	err := runIslandsList(context.Background(), client, &stdout, gf, listFlags{Size: 2})
	require.NoError(t, err)
	require.Contains(t, stdout.String(), `"links"`)
	require.Contains(t, stdout.String(), `"page":{"prevCursor"`)
}
