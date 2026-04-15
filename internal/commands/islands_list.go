package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/ygncode/fortnite-cli/internal/api"
	"github.com/ygncode/fortnite-cli/internal/output"
	"github.com/ygncode/fortnite-cli/internal/slim"
)

// listFlags captures --size/--after/--before for the islands list command.
type listFlags struct {
	Size   int
	After  string
	Before string
}

func newListCmd(gf *GlobalFlags) *cobra.Command {
	var f listFlags
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List islands (cursor-paginated, newest-released first).",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := withSignalContext()
			defer cancel()
			client := newClient(gf)
			if err := runIslandsList(ctx, client, os.Stdout, gf, f); err != nil {
				printError(err)
				return err
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&f.Size, "size", 100, "max results per page (1-1000)")
	cmd.Flags().StringVar(&f.After, "after", "", "cursor: return results after this position")
	cmd.Flags().StringVar(&f.Before, "before", "", "cursor: return results before this position")
	return cmd
}

// runIslandsList is the testable core of the command.
func runIslandsList(ctx context.Context, client *api.Client, stdout io.Writer, gf *GlobalFlags, f listFlags) error {
	if f.Size < 1 || f.Size > 1000 {
		return fmt.Errorf("--size must be in [1,1000], got %d", f.Size)
	}

	// Raw mode: bypass typed decoding and stream the original JSON.
	if gf.Raw {
		raw, err := getRaw(ctx, client, "/islands", listQueryValues(f))
		if err != nil {
			return err
		}
		var buf bytes.Buffer
		if err := json.Compact(&buf, raw); err != nil {
			return fmt.Errorf("compact raw response: %w", err)
		}
		w := &output.Writer{Out: stdout, Format: formatFrom(gf)}
		return w.WriteRaw(buf.Bytes())
	}

	resp, err := client.ListIslands(ctx, api.ListIslandsParams{Size: f.Size, After: f.After, Before: f.Before})
	if err != nil {
		return err
	}
	rawRespBytes, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	slimmed, err := slim.IslandList(rawRespBytes)
	if err != nil {
		return err
	}
	w := &output.Writer{Out: stdout, Format: formatFrom(gf)}
	return w.WriteRaw(slimmed)
}

// formatFrom translates GlobalFlags.Format into an output.Format.
// Defined here because islands_list is the first command that needs it; reused by later commands.
func formatFrom(gf *GlobalFlags) output.Format {
	if gf.Format == "ndjson" {
		return output.FormatNDJSON
	}
	return output.FormatJSON
}
