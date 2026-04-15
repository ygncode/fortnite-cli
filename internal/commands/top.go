package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/ygncode/fortnite-cli/internal/api"
	"github.com/ygncode/fortnite-cli/internal/output"
	"github.com/ygncode/fortnite-cli/internal/timerange"
)

type topFlags struct {
	By       string
	Interval string
	Last     string
	Limit    int
	Scan     int
}

// metricNameMap translates --by values (camelCase as in the API) to path segments.
var metricNameMap = map[string]string{
	"peakCCU":                 "peak-ccu",
	"favorites":               "favorites",
	"minutesPlayed":           "minutes-played",
	"averageMinutesPerPlayer": "average-minutes-per-player",
	"recommendations":         "recommendations",
	"uniquePlayers":           "unique-players",
	"plays":                   "plays",
}

func newTopCmd(gf *GlobalFlags) *cobra.Command {
	var f topFlags
	cmd := &cobra.Command{
		Use:   "top",
		Short: "Top islands by a metric. Expensive — scans up to --scan islands (default 500).",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := withSignalContext()
			defer cancel()
			client := newClient(gf)
			if err := runTop(ctx, client, os.Stdout, gf, f); err != nil {
				printError(err)
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&f.By, "by", "peakCCU", "metric: peakCCU | plays | uniquePlayers | minutesPlayed | favorites | recommendations | averageMinutesPerPlayer")
	cmd.Flags().StringVar(&f.Interval, "interval", "hour", "day | hour | minute")
	cmd.Flags().StringVar(&f.Last, "last", "24h", "shorthand time range, e.g. 24h, 7d")
	cmd.Flags().IntVar(&f.Limit, "limit", 10, "max results to return")
	cmd.Flags().IntVar(&f.Scan, "scan", 500, "max islands to scan (1-1000); see docs")
	return cmd
}

type topItem struct {
	Code  string  `json:"code"`
	Title string  `json:"title"`
	Value float64 `json:"value"`
}

func runTop(ctx context.Context, client *api.Client, stdout io.Writer, gf *GlobalFlags, f topFlags) error {
	metricPath, ok := metricNameMap[f.By]
	if !ok {
		return fmt.Errorf("unknown --by %q", f.By)
	}
	if f.Scan < 1 || f.Scan > 1000 {
		return fmt.Errorf("--scan must be in [1,1000]")
	}
	if f.Limit < 1 {
		return fmt.Errorf("--limit must be >= 1")
	}
	if gf.Concurrency < 1 {
		gf.Concurrency = 8
	}
	r, _, err := timerange.Parse(timerange.Input{Last: f.Last, Interval: f.Interval}, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("time range: %w", err)
	}

	islands, err := scanIslands(ctx, client, f.Scan)
	if err != nil {
		return err
	}

	results := make([]topItem, 0, len(islands))
	var mu sync.Mutex
	sem := make(chan struct{}, gf.Concurrency)
	var wg sync.WaitGroup
	var firstErr error
	for _, island := range islands {
		wg.Add(1)
		sem <- struct{}{}
		go func(it api.IslandListItem) {
			defer wg.Done()
			defer func() { <-sem }()
			resp, mErr := client.GetIslandMetric(ctx, it.Code, f.Interval, metricPath, api.MetricsParams{From: r.From, To: r.To})
			if mErr != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = mErr
				}
				mu.Unlock()
				return
			}
			best, ok := maxNonNullValue(resp.Intervals)
			if !ok {
				return
			}
			mu.Lock()
			results = append(results, topItem{Code: it.Code, Title: it.Title, Value: best})
			mu.Unlock()
		}(island)
	}
	wg.Wait()
	if firstErr != nil {
		return firstErr
	}

	sort.Slice(results, func(i, j int) bool { return results[i].Value > results[j].Value })
	if len(results) > f.Limit {
		results = results[:f.Limit]
	}
	data, err := json.Marshal(results)
	if err != nil {
		return err
	}
	w := &output.Writer{Out: stdout, Format: formatFrom(gf)}
	return w.WriteRaw(data)
}

// scanIslands paginates /islands accumulating up to limit items.
func scanIslands(ctx context.Context, client *api.Client, limit int) ([]api.IslandListItem, error) {
	var all []api.IslandListItem
	cursor := ""
	for len(all) < limit {
		batch := 100
		if remaining := limit - len(all); remaining < batch {
			batch = remaining
		}
		page, err := client.ListIslands(ctx, api.ListIslandsParams{Size: batch, After: cursor})
		if err != nil {
			return nil, err
		}
		all = append(all, page.Data...)
		if page.Meta.Page.NextCursor == "" {
			break
		}
		cursor = page.Meta.Page.NextCursor
	}
	return all, nil
}

// maxNonNullValue returns the largest non-null value in a metric series.
func maxNonNullValue(series []api.MetricValue) (float64, bool) {
	var best float64
	found := false
	for _, v := range series {
		if v.Value == nil {
			continue
		}
		if !found || *v.Value > best {
			best = *v.Value
			found = true
		}
	}
	return best, found
}
