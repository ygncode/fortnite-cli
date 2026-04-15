package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/ygncode/fortnite-cli/internal/api"
	"github.com/ygncode/fortnite-cli/internal/output"
	"github.com/ygncode/fortnite-cli/internal/timerange"
)

type compareFlags struct {
	Codes    []string
	Metric   string
	Interval string
	Last     string
	From     string
	To       string
}

func newCompareCmd(gf *GlobalFlags) *cobra.Command {
	var f compareFlags
	cmd := &cobra.Command{
		Use:   "compare <code1> <code2> [...]",
		Short: "Compare two or more islands on a single metric.",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			f.Codes = args
			ctx, cancel := withSignalContext()
			defer cancel()
			client := newClient(gf)
			if err := runCompare(ctx, client, os.Stdout, gf, f); err != nil {
				printError(err)
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&f.Metric, "metric", "peakCCU", "metric to compare (same values as top --by)")
	cmd.Flags().StringVar(&f.Interval, "interval", "hour", "day | hour | minute")
	cmd.Flags().StringVar(&f.Last, "last", "24h", "shorthand time range")
	cmd.Flags().StringVar(&f.From, "from", "", "start of range (RFC3339)")
	cmd.Flags().StringVar(&f.To, "to", "", "end of range (RFC3339)")
	return cmd
}

type compareSeries struct {
	Code   string     `json:"code"`
	Title  string     `json:"title"`
	Values []*float64 `json:"values"`
}

type compareOutput struct {
	Timestamps []string        `json:"timestamps"`
	Metric     string          `json:"metric"`
	Interval   string          `json:"interval"`
	Series     []compareSeries `json:"series"`
}

func runCompare(ctx context.Context, client *api.Client, stdout io.Writer, gf *GlobalFlags, f compareFlags) error {
	metricPath, ok := metricNameMap[f.Metric]
	if !ok {
		return fmt.Errorf("unknown --metric %q", f.Metric)
	}
	if gf.Concurrency < 1 {
		gf.Concurrency = 8
	}
	r, _, err := timerange.Parse(timerange.Input{
		Last: f.Last, From: f.From, To: f.To, Interval: f.Interval,
	}, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("time range: %w", err)
	}

	type fetched struct {
		idx    int
		title  string
		values []api.MetricValue
		err    error
	}
	out := make([]fetched, len(f.Codes))
	var wg sync.WaitGroup
	sem := make(chan struct{}, gf.Concurrency)
	for i, code := range f.Codes {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int, code string) {
			defer wg.Done()
			defer func() { <-sem }()
			meta, err := client.GetIsland(ctx, code)
			if err != nil {
				out[i] = fetched{idx: i, err: err}
				return
			}
			resp, err := client.GetIslandMetric(ctx, code, f.Interval, metricPath, api.MetricsParams{From: r.From, To: r.To})
			if err != nil {
				out[i] = fetched{idx: i, err: err}
				return
			}
			out[i] = fetched{idx: i, title: meta.Title, values: resp.Intervals}
		}(i, code)
	}
	wg.Wait()
	for _, res := range out {
		if res.err != nil {
			return res.err
		}
	}

	// Build a shared timestamp axis from the first non-empty series.
	var axis []string
	for _, res := range out {
		if len(res.values) > 0 {
			axis = make([]string, len(res.values))
			for i, v := range res.values {
				axis[i] = v.Timestamp
			}
			break
		}
	}

	result := compareOutput{
		Timestamps: axis,
		Metric:     f.Metric,
		Interval:   f.Interval,
		Series:     make([]compareSeries, 0, len(out)),
	}
	for i, res := range out {
		vals := make([]*float64, len(axis))
		for j, v := range res.values {
			if j < len(vals) {
				vals[j] = v.Value
			}
		}
		result.Series = append(result.Series, compareSeries{
			Code:   f.Codes[i],
			Title:  res.title,
			Values: vals,
		})
	}

	data, err := json.Marshal(result)
	if err != nil {
		return err
	}
	w := &output.Writer{Out: stdout, Format: formatFrom(gf)}
	return w.WriteRaw(data)
}
