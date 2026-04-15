package commands

import (
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

type summaryFlags struct {
	Code string
	Last string
}

func newSummaryCmd(gf *GlobalFlags) *cobra.Command {
	var f summaryFlags
	cmd := &cobra.Command{
		Use:   "summary <code>",
		Short: "Full picture of one island: metadata + all bundled metrics.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f.Code = args[0]
			ctx, cancel := withSignalContext()
			defer cancel()
			client := newClient(gf)
			if err := runSummary(ctx, client, os.Stdout, gf, f); err != nil {
				printError(err)
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&f.Last, "last", "", "shorthand time range, e.g. 7d (default: previous day)")
	return cmd
}

func runSummary(ctx context.Context, client *api.Client, stdout io.Writer, gf *GlobalFlags, f summaryFlags) error {
	r, _, err := timerange.Parse(timerange.Input{Last: f.Last, Interval: "day"}, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("time range: %w", err)
	}

	meta, err := client.GetIsland(ctx, f.Code)
	if err != nil {
		return err
	}
	metrics, err := client.GetBundledMetrics(ctx, f.Code, api.MetricsParams{From: r.From, To: r.To})
	if err != nil {
		return err
	}

	if gf.Raw {
		combined := map[string]any{"metadata": meta, "metrics": metrics}
		data, err := json.Marshal(combined)
		if err != nil {
			return err
		}
		w := &output.Writer{Out: stdout, Format: formatFrom(gf)}
		return w.WriteRaw(data)
	}

	metricsBytes, err := json.Marshal(metrics)
	if err != nil {
		return err
	}
	slimmed, err := slim.BundledMetrics(metricsBytes, "day")
	if err != nil {
		return err
	}
	var slimMetrics any
	if err := json.Unmarshal(slimmed, &slimMetrics); err != nil {
		return err
	}
	combined := map[string]any{
		"metadata": meta,
		"metrics":  slimMetrics,
	}
	data, err := json.Marshal(combined)
	if err != nil {
		return err
	}
	w := &output.Writer{Out: stdout, Format: formatFrom(gf)}
	return w.WriteRaw(data)
}
