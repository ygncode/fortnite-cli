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

type retentionFlags struct {
	Code     string
	Interval string
	Last     string
	From     string
	To       string
}

func newRetentionCmd(gf *GlobalFlags) *cobra.Command {
	var f retentionFlags
	cmd := &cobra.Command{
		Use:   "retention <code>",
		Short: "D1/D7 retention (day only).",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f.Code = args[0]
			ctx, cancel := withSignalContext()
			defer cancel()
			client := newClient(gf)
			if err := runRetention(ctx, client, os.Stdout, gf, f); err != nil {
				printError(err)
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&f.Interval, "interval", "day", "day (only day is valid)")
	cmd.Flags().StringVar(&f.Last, "last", "", "shorthand time range, e.g. 7d")
	cmd.Flags().StringVar(&f.From, "from", "", "start of range (RFC3339)")
	cmd.Flags().StringVar(&f.To, "to", "", "end of range (RFC3339)")
	return cmd
}

func runRetention(ctx context.Context, client *api.Client, stdout io.Writer, gf *GlobalFlags, f retentionFlags) error {
	interval := f.Interval
	if interval == "" {
		interval = "day"
	}
	if interval != "day" {
		return fmt.Errorf("retention is only available at day interval")
	}
	r, warn, err := timerange.Parse(timerange.Input{
		Last: f.Last, From: f.From, To: f.To, Interval: "day",
	}, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("time range: %w", err)
	}
	if warn != "" {
		fmt.Fprintln(os.Stderr, "warning:", warn)
	}
	p := api.MetricsParams{From: r.From, To: r.To}

	if gf.Raw {
		raw, err := getRaw(ctx, client, "/islands/"+f.Code+"/metrics/day/retention", p.EncodePublic())
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

	resp, err := client.GetRetention(ctx, f.Code, p)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	slimmed, err := slim.Retention(raw)
	if err != nil {
		return err
	}
	w := &output.Writer{Out: stdout, Format: formatFrom(gf)}
	return w.WriteRaw(slimmed)
}
