// Package slim transforms raw Fortnite Ecosystem API responses into agent-friendly
// compact JSON. All functions are pure — no I/O, no HTTP.
package slim

import (
	"encoding/json"
	"fmt"
	"time"
)

// IslandList transforms a GET /islands response into slim form:
// - Envelope {data, links, meta} is flattened into {_pagination?, data}.
// - Each island's nested meta.page.cursor is lifted to a top-level "cursor" field.
// - Null optional fields (displayName, creatorCode, category) are omitted by the source marshaler.
// - `_pagination` is omitted entirely when both links are null.
func IslandList(raw []byte) ([]byte, error) {
	var in struct {
		Links struct {
			Prev *string `json:"prev"`
			Next *string `json:"next"`
		} `json:"links"`
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return nil, fmt.Errorf("slim island list: decode: %w", err)
	}

	out := map[string]any{}

	if in.Links.Prev != nil || in.Links.Next != nil {
		pag := map[string]any{}
		if in.Links.Prev != nil {
			pag["prev"] = *in.Links.Prev
		}
		if in.Links.Next != nil {
			pag["next"] = *in.Links.Next
		}
		out["_pagination"] = pag
	}

	slimmed := make([]map[string]any, 0, len(in.Data))
	for _, item := range in.Data {
		o := map[string]any{}
		for k, v := range item {
			if k == "meta" {
				continue
			}
			if s, ok := v.(string); ok && s == "" {
				continue // omit empty optional strings
			}
			o[k] = v
		}
		if meta, ok := item["meta"].(map[string]any); ok {
			if page, ok := meta["page"].(map[string]any); ok {
				if c, ok := page["cursor"].(string); ok && c != "" {
					o["cursor"] = c
				}
			}
		}
		slimmed = append(slimmed, o)
	}
	out["data"] = slimmed

	return json.Marshal(out)
}

// BundledMetrics slims a BundledMetrics response (from /metrics or /metrics/{interval}).
// interval must be "day", "hour", or "minute" and controls timestamp rounding.
func BundledMetrics(raw []byte, interval string) ([]byte, error) {
	var in map[string]any
	if err := json.Unmarshal(raw, &in); err != nil {
		return nil, fmt.Errorf("slim bundled metrics: decode: %w", err)
	}

	// Known metric arrays. Retention is special (d1/d7 instead of value).
	metricKeys := []string{
		"averageMinutesPerPlayer",
		"peakCCU",
		"favorites",
		"minutesPlayed",
		"recommendations",
		"plays",
		"uniquePlayers",
	}
	out := map[string]any{}
	for _, k := range metricKeys {
		arr, ok := in[k].([]any)
		if !ok {
			out[k] = nil
			continue
		}
		out[k] = slimValueArray(arr, interval)
	}
	// Retention.
	if retArr, ok := in["retention"].([]any); ok {
		out["retention"] = slimRetentionArray(retArr, interval)
	} else {
		out["retention"] = nil
	}

	return json.Marshal(out)
}

// slimValueArray rounds timestamps and collapses all-null arrays to nil.
func slimValueArray(arr []any, interval string) any {
	allNull := true
	slim := make([]map[string]any, 0, len(arr))
	for _, item := range arr {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		value := m["value"]
		ts, _ := m["timestamp"].(string)
		if value != nil {
			allNull = false
		}
		slim = append(slim, map[string]any{
			"value":     value,
			"timestamp": roundTimestamp(ts, interval),
		})
	}
	if allNull {
		return nil
	}
	return slim
}

// slimRetentionArray collapses fully-null retention arrays (where both d1 and d7 are null) to nil.
func slimRetentionArray(arr []any, interval string) any {
	allNull := true
	slim := make([]map[string]any, 0, len(arr))
	for _, item := range arr {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		d1 := m["d1"]
		d7 := m["d7"]
		ts, _ := m["timestamp"].(string)
		if d1 != nil || d7 != nil {
			allNull = false
		}
		slim = append(slim, map[string]any{
			"d1":        d1,
			"d7":        d7,
			"timestamp": roundTimestamp(ts, interval),
		})
	}
	if allNull {
		return nil
	}
	return slim
}

// roundTimestamp truncates an ISO8601 timestamp to the given interval granularity.
// day → "2026-04-14", hour → "2026-04-14T07:00", minute → "2026-04-14T07:20".
// Returns the input unchanged if it can't be parsed.
func roundTimestamp(ts, interval string) string {
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return ts
	}
	switch interval {
	case "day":
		return t.UTC().Format("2006-01-02")
	case "hour":
		return t.UTC().Format("2006-01-02T15:04")
	case "minute":
		return t.UTC().Format("2006-01-02T15:04")
	default:
		return ts
	}
}
