//go:build e2e

package api

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestE2EListIslands hits the real Fortnite API. Only runs with -tags=e2e.
func TestE2EListIslands(t *testing.T) {
	c := DefaultClient(30 * time.Second)
	resp, err := c.ListIslands(context.Background(), ListIslandsParams{Size: 5})
	require.NoError(t, err)
	require.NotEmpty(t, resp.Data)
	require.NotEmpty(t, resp.Data[0].Code)
	require.NotEmpty(t, resp.Data[0].Title)
}
