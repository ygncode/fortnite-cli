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
