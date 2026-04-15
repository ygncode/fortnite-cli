// Package slim transforms raw Fortnite Ecosystem API responses into agent-friendly
// compact JSON. All functions are pure — no I/O, no HTTP.
package slim

import (
	"encoding/json"
	"fmt"
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
