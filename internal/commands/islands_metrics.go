package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/ygncode/fortnite-cli/internal/api"
	"github.com/ygncode/fortnite-cli/internal/output"
	"github.com/ygncode/fortnite-cli/internal/slim"
	"github.com/ygncode/fortnite-cli/internal/timerange"
)

type metricsFlags struct {
	Code     string
	Interval string
	Metrics  []string
	Last     string
	From     string
	To       string
}

func newMetricsCmd(gf *GlobalFlags) *cobra.Command {
	var f metricsFlags
	cmd := &cobra.Command{
		Use:   "metrics <code>",
		Short: "Island metrics. Without --interval: bundled day metrics. With --interval: filterable.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f.Code = args[0]
			ctx, cancel := withSignalContext()
			defer cancel()
			client := newClient(gf)
			if err := runIslandsMetrics(ctx, client, os.Stdout, gf, f); err != nil {
				printError(err)
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&f.Interval, "interval", "", "day | hour | minute (omit for bundled day response)")
	cmd.Flags().StringSliceVar(&f.Metrics, "metric", nil, "filter to specific metrics (repeatable; requires --interval)")
	cmd.Flags().StringVar(&f.Last, "last", "", "shorthand time range, e.g. 24h, 7d")
	cmd.Flags().StringVar(&f.From, "from", "", "start of range (RFC3339)")
	cmd.Flags().StringVar(&f.To, "to", "", "end of range (RFC3339)")
	return cmd
}

func runIslandsMetrics(ctx context.Context, client *api.Client, stdout io.Writer, gf *GlobalFlags, f metricsFlags) error {
	interval := f.Interval
	effective := interval
	if effective == "" {
		effective = "day"
	}
	r, warn, err := timerange.Parse(timerange.Input{
		Last: f.Last, From: f.From, To: f.To, Interval: effective,
	}, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("time range: %w", err)
	}
	if warn != "" {
		fmt.Fprintln(os.Stderr, "warning:", warn)
	}
	if len(f.Metrics) > 0 && interval == "" {
		return fmt.Errorf("--metric requires --interval")
	}

	p := api.MetricsParams{From: r.From, To: r.To, Metrics: f.Metrics}

	if gf.Raw {
		path := "/islands/" + f.Code + "/metrics"
		if interval != "" {
			path += "/" + interval
		}
		raw, err := getRaw(ctx, client, path, p.EncodePublic())
		if err != nil {
			return err
		}
		var buf bytes.Buffer
		if err := json.Compact(&buf, raw); err != nil {
			return err
		}
		w := &output.Writer{Out: stdout, Format: formatFrom(gf)}
		return w.WriteRaw(buf.Bytes())
	}

	var resp *api.BundledMetrics
	if interval == "" {
		resp, err = client.GetBundledMetrics(ctx, f.Code, p)
	} else {
		resp, err = client.GetFilterableMetrics(ctx, f.Code, interval, p)
	}
	if err != nil {
		return err
	}
	raw, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	slimmed, err := slim.BundledMetrics(raw, effective)
	if err != nil {
		return err
	}
	w := &output.Writer{Out: stdout, Format: formatFrom(gf)}
	return w.WriteRaw(slimmed)
}
