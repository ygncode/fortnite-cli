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

// singleMetricSpec describes a per-metric subcommand.
type singleMetricSpec struct {
	Use     string
	Short   string
	Metric  string
	DayOnly bool
}

// singleMetricSpecs is the catalogue of non-retention per-metric commands.
var singleMetricSpecs = []singleMetricSpec{
	{Use: "peak-ccu", Short: "Peak concurrent players.", Metric: "peak-ccu"},
	{Use: "favorites", Short: "Times the island was favorited.", Metric: "favorites"},
	{Use: "minutes-played", Short: "Total player minutes.", Metric: "minutes-played"},
	{Use: "avg-minutes", Short: "Average minutes per player (day only).", Metric: "average-minutes-per-player", DayOnly: true},
	{Use: "recommendations", Short: "Player recommendations.", Metric: "recommendations"},
	{Use: "unique-players", Short: "Unique players.", Metric: "unique-players"},
	{Use: "plays", Short: "Play count.", Metric: "plays"},
}

type singleMetricFlags struct {
	Code     string
	Interval string
	Last     string
	From     string
	To       string
}

func newSingleMetricCmd(gf *GlobalFlags, spec singleMetricSpec) *cobra.Command {
	var f singleMetricFlags
	cmd := &cobra.Command{
		Use:   spec.Use + " <code>",
		Short: spec.Short,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f.Code = args[0]
			ctx, cancel := withSignalContext()
			defer cancel()
			client := newClient(gf)
			if err := runSingleMetric(ctx, client, os.Stdout, gf, spec, f); err != nil {
				printError(err)
				return err
			}
			return nil
		},
	}
	if !spec.DayOnly {
		cmd.Flags().StringVar(&f.Interval, "interval", "day", "day | hour | minute")
	} else {
		cmd.Flags().StringVar(&f.Interval, "interval", "day", "day only (this metric is day-only)")
	}
	cmd.Flags().StringVar(&f.Last, "last", "", "shorthand time range, e.g. 24h, 7d")
	cmd.Flags().StringVar(&f.From, "from", "", "start of range (RFC3339)")
	cmd.Flags().StringVar(&f.To, "to", "", "end of range (RFC3339)")
	return cmd
}

func runSingleMetric(ctx context.Context, client *api.Client, stdout io.Writer, gf *GlobalFlags, spec singleMetricSpec, f singleMetricFlags) error {
	interval := f.Interval
	if interval == "" {
		interval = "day"
	}
	if spec.DayOnly && interval != "day" {
		return fmt.Errorf("%s is only available at day interval", spec.Metric)
	}
	if interval != "day" && interval != "hour" && interval != "minute" {
		return fmt.Errorf("invalid --interval %q", interval)
	}
	r, warn, err := timerange.Parse(timerange.Input{
		Last: f.Last, From: f.From, To: f.To, Interval: interval,
	}, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("time range: %w", err)
	}
	if warn != "" {
		fmt.Fprintln(os.Stderr, "warning:", warn)
	}
	p := api.MetricsParams{From: r.From, To: r.To}

	if gf.Raw {
		path := "/islands/" + f.Code + "/metrics/" + interval + "/" + spec.Metric
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

	resp, err := client.GetIslandMetric(ctx, f.Code, interval, spec.Metric, p)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	slimmed, err := slim.MetricResponse(raw, interval)
	if err != nil {
		return err
	}
	w := &output.Writer{Out: stdout, Format: formatFrom(gf)}
	return w.WriteRaw(slimmed)
}
