package api

import (
	"context"
	"net/url"
	"strconv"
	"time"
)

// ListIslandsParams is the query for GET /islands.
type ListIslandsParams struct {
	Size   int
	After  string
	Before string
}

// MetricsParams is the shared query struct for metric endpoints.
// Not every field applies to every endpoint; unused fields are zero and ignored.
type MetricsParams struct {
	From    time.Time
	To      time.Time
	Metrics []string // only used by GetFilterableMetrics
}

func (p MetricsParams) encode() url.Values {
	v := url.Values{}
	if !p.From.IsZero() {
		v.Set("from", p.From.UTC().Format(time.RFC3339Nano))
	}
	if !p.To.IsZero() {
		v.Set("to", p.To.UTC().Format(time.RFC3339Nano))
	}
	for _, m := range p.Metrics {
		v.Add("metrics", m)
	}
	return v
}

// ListIslands calls GET /islands.
func (c *Client) ListIslands(ctx context.Context, p ListIslandsParams) (*IslandListResponse, error) {
	q := url.Values{}
	if p.Size > 0 {
		q.Set("size", strconv.Itoa(p.Size))
	}
	if p.After != "" {
		q.Set("after", p.After)
	}
	if p.Before != "" {
		q.Set("before", p.Before)
	}
	var out IslandListResponse
	if err := c.Get(ctx, "/islands", q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetIsland calls GET /islands/{code}.
func (c *Client) GetIsland(ctx context.Context, code string) (*IslandMetadata, error) {
	var out IslandMetadata
	if err := c.Get(ctx, "/islands/"+code, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetBundledMetrics calls GET /islands/{code}/metrics (day interval, all metrics).
func (c *Client) GetBundledMetrics(ctx context.Context, code string, p MetricsParams) (*BundledMetrics, error) {
	var out BundledMetrics
	if err := c.Get(ctx, "/islands/"+code+"/metrics", p.encode(), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetFilterableMetrics calls GET /islands/{code}/metrics/{interval}.
func (c *Client) GetFilterableMetrics(ctx context.Context, code, interval string, p MetricsParams) (*BundledMetrics, error) {
	var out BundledMetrics
	if err := c.Get(ctx, "/islands/"+code+"/metrics/"+interval, p.encode(), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetIslandMetric calls one of the seven single-metric endpoints. metric must be one of:
// "peak-ccu", "favorites", "minutes-played", "average-minutes-per-player",
// "recommendations", "unique-players", "plays" — NOT "retention" (use GetRetention).
func (c *Client) GetIslandMetric(ctx context.Context, code, interval, metric string, p MetricsParams) (*MetricResponse, error) {
	var out MetricResponse
	path := "/islands/" + code + "/metrics/" + interval + "/" + metric
	if err := c.Get(ctx, path, p.encode(), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetRetention calls GET /islands/{code}/metrics/day/retention. The API only
// supports day interval for retention; callers should not pass any other.
func (c *Client) GetRetention(ctx context.Context, code string, p MetricsParams) (*RetentionResponse, error) {
	var out RetentionResponse
	if err := c.Get(ctx, "/islands/"+code+"/metrics/day/retention", p.encode(), &out); err != nil {
		return nil, err
	}
	return &out, nil
}
