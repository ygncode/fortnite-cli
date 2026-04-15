// Package commands wires cobra commands to the pure packages.
package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/ygncode/fortnite-cli/internal/api"
	"github.com/ygncode/fortnite-cli/internal/output"
)

// GlobalFlags are cross-cutting flags available on every command.
type GlobalFlags struct {
	Raw         bool
	Format      string // "json" | "ndjson"
	NoRetry     bool
	Timeout     time.Duration
	Concurrency int
}

// NewRoot builds the root `fortnite` command wired with all subcommands.
func NewRoot(version string) *cobra.Command {
	gf := &GlobalFlags{}
	root := &cobra.Command{
		Use:          "fortnite",
		Short:        "Query the Fortnite Ecosystem v1 Data API (island metadata and metrics).",
		SilenceUsage: true,
		Version:      version,
	}

	root.PersistentFlags().BoolVar(&gf.Raw, "raw", false, "emit the API response as-is (no slimming)")
	root.PersistentFlags().StringVar(&gf.Format, "format", "json", "output format: json or ndjson")
	root.PersistentFlags().BoolVar(&gf.NoRetry, "no-retry", false, "disable 429/5xx retry")
	root.PersistentFlags().DurationVar(&gf.Timeout, "timeout", 30*time.Second, "HTTP timeout")
	root.PersistentFlags().IntVar(&gf.Concurrency, "concurrency", 8, "worker pool size for composite verbs")

	// islands subtree
	islands := &cobra.Command{Use: "islands", Short: "Island metadata and metrics endpoints."}
	islands.AddCommand(newListCmd(gf))
	islands.AddCommand(newGetCmd(gf))
	islands.AddCommand(newMetricsCmd(gf))
	for _, m := range singleMetricSpecs {
		islands.AddCommand(newSingleMetricCmd(gf, m))
	}
	islands.AddCommand(newRetentionCmd(gf))
	root.AddCommand(islands)

	// composite verbs
	root.AddCommand(newTopCmd(gf))
	root.AddCommand(newSummaryCmd(gf))
	root.AddCommand(newCompareCmd(gf))
	root.AddCommand(newSearchCmd(gf))

	return root
}

// newClient builds a configured api.Client from the global flags.
func newClient(gf *GlobalFlags) *api.Client {
	c := api.DefaultClient(gf.Timeout)
	c.NoRetry = gf.NoRetry
	return c
}

// newWriter builds an output.Writer from the global flags.
func newWriter(gf *GlobalFlags) *output.Writer {
	f := output.FormatJSON
	if gf.Format == "ndjson" {
		f = output.FormatNDJSON
	}
	return &output.Writer{Out: os.Stdout, Format: f}
}

// withSignalContext returns a context cancelled on Ctrl-C.
func withSignalContext() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt)
}

// printError writes a structured error to stderr.
func printError(err error) {
	out := map[string]any{"error": "generic", "message": err.Error()}
	var apiErr *api.APIError
	if asAPIErr(err, &apiErr) {
		out["error"] = classify(apiErr)
		out["message"] = apiErr.Message
		out["status"] = apiErr.Status
		if apiErr.UUID != "" {
			out["uuid"] = apiErr.UUID
		}
	}
	data, _ := json.Marshal(out)
	fmt.Fprintln(os.Stderr, string(data))
}

func classify(apiErr *api.APIError) string {
	switch {
	case apiErr.Status == 429:
		return "rate_limited"
	case apiErr.Status == 404:
		return "not_found"
	case apiErr.Status == 400:
		return "bad_request"
	case apiErr.Status >= 500:
		return "upstream_error"
	default:
		return "api_error"
	}
}

// asAPIErr is a tiny shim so root.go does not need to import "errors" just to alias errors.As.
func asAPIErr(err error, target **api.APIError) bool {
	for err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			*target = apiErr
			return true
		}
		u, ok := err.(interface{ Unwrap() error })
		if !ok {
			return false
		}
		err = u.Unwrap()
	}
	return false
}

// ExitCodeFor classifies an error into the documented exit codes.
// 0=success, 1=usage/validation, 2=API error, 3=network/IO.
func ExitCodeFor(err error) int {
	if err == nil {
		return 0
	}
	var apiErr *api.APIError
	if asAPIErr(err, &apiErr) {
		return 2
	}
	return 3
}

// getRaw performs a GET and returns the raw response body. Uses the client's
// retry/error machinery but skips JSON decoding.
func getRaw(ctx context.Context, client *api.Client, path string, q url.Values) ([]byte, error) {
	var raw json.RawMessage
	if err := client.Get(ctx, path, q, &raw); err != nil {
		return nil, err
	}
	return []byte(raw), nil
}

// listQueryValues builds url.Values for the islands list endpoint.
func listQueryValues(f listFlags) url.Values {
	q := url.Values{}
	if f.Size > 0 {
		q.Set("size", strconv.Itoa(f.Size))
	}
	if f.After != "" {
		q.Set("after", f.After)
	}
	if f.Before != "" {
		q.Set("before", f.Before)
	}
	return q
}
