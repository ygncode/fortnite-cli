package timerange

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseLast(t *testing.T) {
	now := time.Date(2026, 4, 15, 12, 0, 0, 0, time.UTC)
	r, warn, err := Parse(Input{Last: "24h", Interval: "hour"}, now)
	require.NoError(t, err)
	require.Empty(t, warn)
	require.Equal(t, now.Add(-24*time.Hour), r.From)
	require.Equal(t, now, r.To)
}

func TestParseFromTo(t *testing.T) {
	now := time.Date(2026, 4, 15, 12, 0, 0, 0, time.UTC)
	r, _, err := Parse(Input{From: "2026-04-14T00:00:00Z", To: "2026-04-15T00:00:00Z", Interval: "day"}, now)
	require.NoError(t, err)
	require.Equal(t, "2026-04-14T00:00:00Z", r.From.Format(time.RFC3339))
	require.Equal(t, "2026-04-15T00:00:00Z", r.To.Format(time.RFC3339))
}

func TestParseLastWinsOverFromTo(t *testing.T) {
	now := time.Date(2026, 4, 15, 12, 0, 0, 0, time.UTC)
	r, warn, err := Parse(Input{Last: "1h", From: "2026-04-14T00:00:00Z", To: "2026-04-15T00:00:00Z", Interval: "hour"}, now)
	require.NoError(t, err)
	require.NotEmpty(t, warn)
	require.Equal(t, now.Add(-1*time.Hour), r.From)
}

func TestParseRejectsFromAfterTo(t *testing.T) {
	now := time.Date(2026, 4, 15, 12, 0, 0, 0, time.UTC)
	_, _, err := Parse(Input{From: "2026-04-15T00:00:00Z", To: "2026-04-14T00:00:00Z", Interval: "day"}, now)
	require.Error(t, err)
}

func TestParseRejectsOver7Days(t *testing.T) {
	now := time.Date(2026, 4, 15, 12, 0, 0, 0, time.UTC)
	_, _, err := Parse(Input{Last: "8d", Interval: "day"}, now)
	require.Error(t, err)
	require.Contains(t, err.Error(), "7-day")
}

func TestParseRejectsOver7DaysFromExplicit(t *testing.T) {
	now := time.Date(2026, 4, 15, 12, 0, 0, 0, time.UTC)
	_, _, err := Parse(Input{From: "2026-04-01T00:00:00Z", To: "2026-04-15T00:00:00Z", Interval: "day"}, now)
	require.Error(t, err)
	require.Contains(t, err.Error(), "7-day")
}

func TestParseClampsToFuture(t *testing.T) {
	now := time.Date(2026, 4, 15, 12, 0, 0, 0, time.UTC)
	r, warn, err := Parse(Input{From: "2026-04-15T00:00:00Z", To: "2026-04-16T00:00:00Z", Interval: "day"}, now)
	require.NoError(t, err)
	require.NotEmpty(t, warn)
	require.Equal(t, now, r.To)
}

func TestParseDefaults(t *testing.T) {
	now := time.Date(2026, 4, 15, 12, 0, 0, 0, time.UTC)

	rDay, _, err := Parse(Input{Interval: "day"}, now)
	require.NoError(t, err)
	require.Equal(t, now.Add(-24*time.Hour), rDay.From)

	rHour, _, err := Parse(Input{Interval: "hour"}, now)
	require.NoError(t, err)
	require.Equal(t, now.Add(-24*time.Hour), rHour.From)

	rMin, _, err := Parse(Input{Interval: "minute"}, now)
	require.NoError(t, err)
	require.Equal(t, now.Add(-60*time.Minute), rMin.From)
}
