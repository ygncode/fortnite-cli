// Package timerange parses --last / --from / --to flags into a validated time range.
package timerange

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// maxHistory is the Fortnite API's historical window cap.
const maxHistory = 7 * 24 * time.Hour

// Input captures the raw flag values from a command.
type Input struct {
	Last     string // e.g. "24h", "7d", "60m"
	From     string // RFC3339
	To       string // RFC3339
	Interval string // "day" | "hour" | "minute"
}

// Range is the parsed, validated time range.
type Range struct {
	From time.Time
	To   time.Time
}

// Parse converts Input into a Range. It returns (range, warnings, error).
// Warnings are non-empty strings intended for stderr (e.g. "both --last and --from provided").
// The caller supplies `now` so tests can pin time.
func Parse(in Input, now time.Time) (Range, string, error) {
	var warn string

	// --last wins over --from/--to.
	if in.Last != "" {
		dur, err := parseDuration(in.Last)
		if err != nil {
			return Range{}, "", fmt.Errorf("invalid --last: %w", err)
		}
		if dur > maxHistory {
			return Range{}, "", errors.New("invalid --last: exceeds API 7-day historical limit")
		}
		if in.From != "" || in.To != "" {
			warn = "--last takes precedence; --from/--to ignored"
		}
		return Range{From: now.Add(-dur), To: now}, warn, nil
	}

	// Explicit --from/--to.
	if in.From != "" || in.To != "" {
		if in.From == "" || in.To == "" {
			return Range{}, "", errors.New("--from and --to must be provided together (or use --last)")
		}
		from, err := time.Parse(time.RFC3339, in.From)
		if err != nil {
			return Range{}, "", fmt.Errorf("invalid --from: %w", err)
		}
		to, err := time.Parse(time.RFC3339, in.To)
		if err != nil {
			return Range{}, "", fmt.Errorf("invalid --to: %w", err)
		}
		if !from.Before(to) {
			return Range{}, "", errors.New("--from must be before --to")
		}
		if to.After(now) {
			warn = "--to is in the future; clamping to now"
			to = now
		}
		if now.Sub(from) > maxHistory {
			return Range{}, "", errors.New("--from exceeds API 7-day historical limit")
		}
		return Range{From: from, To: to}, warn, nil
	}

	// Defaults per interval.
	switch in.Interval {
	case "day", "":
		return Range{From: now.Add(-24 * time.Hour), To: now}, "", nil
	case "hour":
		return Range{From: now.Add(-24 * time.Hour), To: now}, "", nil
	case "minute":
		return Range{From: now.Add(-60 * time.Minute), To: now}, "", nil
	default:
		return Range{}, "", fmt.Errorf("unknown interval %q", in.Interval)
	}
}

// parseDuration accepts "24h", "60m", "7d", "30s" — i.e. Go duration syntax plus "d" for days.
func parseDuration(s string) (time.Duration, error) {
	if strings.HasSuffix(s, "d") {
		prefix := strings.TrimSuffix(s, "d")
		days, err := strconv.Atoi(prefix)
		if err != nil || days < 0 {
			return 0, fmt.Errorf("bad duration %q", s)
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, err
	}
	if d < 0 {
		return 0, errors.New("duration must be positive")
	}
	return d, nil
}
