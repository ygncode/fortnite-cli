// Package api defines request and response types for the Fortnite Ecosystem v1 Data API.
package api

// IslandMetadata represents an island's metadata (summary form).
type IslandMetadata struct {
	Code        string   `json:"code"`
	CreatorCode string   `json:"creatorCode,omitempty"`
	DisplayName string   `json:"displayName,omitempty"`
	Title       string   `json:"title"`
	Category    string   `json:"category,omitempty"`
	CreatedIn   string   `json:"createdIn,omitempty"`
	Tags        []string `json:"tags"`
}

// IslandListItem is an island in a list response, including per-item pagination cursor.
type IslandListItem struct {
	IslandMetadata
	Meta struct {
		Page struct {
			Cursor string `json:"cursor"`
		} `json:"page"`
	} `json:"meta"`
}

// PaginationLinks carries path-style next/prev links.
type PaginationLinks struct {
	Prev *string `json:"prev"`
	Next *string `json:"next"`
}

// PaginationMeta carries count + cursor metadata.
type PaginationMeta struct {
	Count int `json:"count"`
	Page  struct {
		PrevCursor string `json:"prevCursor"`
		NextCursor string `json:"nextCursor"`
	} `json:"page"`
}

// IslandListResponse is the envelope returned by GET /islands.
type IslandListResponse struct {
	Links PaginationLinks  `json:"links"`
	Meta  PaginationMeta   `json:"meta"`
	Data  []IslandListItem `json:"data"`
}

// MetricValue is a single (timestamp, value) bucket where value may be null.
type MetricValue struct {
	Value     *float64 `json:"value"`
	Timestamp string   `json:"timestamp"`
}

// BundledMetrics is the response from GET /islands/{code}/metrics (and the filterable variant).
type BundledMetrics struct {
	AverageMinutesPerPlayer []MetricValue    `json:"averageMinutesPerPlayer,omitempty"`
	PeakCCU                 []MetricValue    `json:"peakCCU,omitempty"`
	Favorites               []MetricValue    `json:"favorites,omitempty"`
	MinutesPlayed           []MetricValue    `json:"minutesPlayed,omitempty"`
	Recommendations         []MetricValue    `json:"recommendations,omitempty"`
	Plays                   []MetricValue    `json:"plays,omitempty"`
	UniquePlayers           []MetricValue    `json:"uniquePlayers,omitempty"`
	Retention               []RetentionValue `json:"retention,omitempty"`
}

// MetricResponse is the response from a single-metric endpoint.
type MetricResponse struct {
	Intervals []MetricValue `json:"intervals"`
}

// RetentionValue is a d1/d7 retention bucket.
type RetentionValue struct {
	D1        *float64 `json:"d1"`
	D7        *float64 `json:"d7"`
	Timestamp string   `json:"timestamp"`
}

// RetentionResponse is the response from .../metrics/day/retention.
type RetentionResponse struct {
	Intervals []RetentionValue `json:"intervals"`
}

// ErrorResponse is the API's error body (4xx).
type ErrorResponse struct {
	ErrorCode    string `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
	UUID         string `json:"uuid"`
}
