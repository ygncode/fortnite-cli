package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/ygncode/fortnite-cli/internal/api"
	"github.com/ygncode/fortnite-cli/internal/output"
)

func newGetCmd(gf *GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <code>",
		Short: "Get metadata for a single island.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := withSignalContext()
			defer cancel()
			client := newClient(gf)
			if err := runIslandsGet(ctx, client, os.Stdout, gf, args[0]); err != nil {
				printError(err)
				return err
			}
			return nil
		},
	}
	return cmd
}

func runIslandsGet(ctx context.Context, client *api.Client, stdout io.Writer, gf *GlobalFlags, code string) error {
	if gf.Raw {
		raw, err := getRaw(ctx, client, "/islands/"+code, nil)
		if err != nil {
			return err
		}
		// Compact any whitespace (mirrors islands list raw path).
		var buf bytes.Buffer
		if err := json.Compact(&buf, raw); err != nil {
			return err
		}
		w := &output.Writer{Out: stdout, Format: formatFrom(gf)}
		return w.WriteRaw(buf.Bytes())
	}
	meta, err := client.GetIsland(ctx, code)
	if err != nil {
		return err
	}
	data, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	w := &output.Writer{Out: stdout, Format: formatFrom(gf)}
	return w.WriteRaw(data)
}
