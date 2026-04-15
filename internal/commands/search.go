package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ygncode/fortnite-cli/internal/api"
	"github.com/ygncode/fortnite-cli/internal/output"
)

type searchFlags struct {
	Tags          []string
	Creator       string
	TitleContains string
	MaxScan       int
}

func newSearchCmd(gf *GlobalFlags) *cobra.Command {
	var f searchFlags
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Find islands matching tag/creator/title filters.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := withSignalContext()
			defer cancel()
			client := newClient(gf)
			if err := runSearch(ctx, client, os.Stdout, gf, f); err != nil {
				printError(err)
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringSliceVar(&f.Tags, "tag", nil, "filter by tag (repeatable; island must contain all)")
	cmd.Flags().StringVar(&f.Creator, "creator", "", "filter by creatorCode (exact match)")
	cmd.Flags().StringVar(&f.TitleContains, "title-contains", "", "filter by case-insensitive substring of title")
	cmd.Flags().IntVar(&f.MaxScan, "max-scan", 500, "max islands to scan (1-1000)")
	return cmd
}

type searchResult struct {
	Data []api.IslandMetadata `json:"data"`
	Meta struct {
		Scanned   int  `json:"scanned"`
		Matched   int  `json:"matched"`
		Truncated bool `json:"truncated"`
	} `json:"meta"`
}

func runSearch(ctx context.Context, client *api.Client, stdout io.Writer, gf *GlobalFlags, f searchFlags) error {
	if f.MaxScan < 1 || f.MaxScan > 1000 {
		return fmt.Errorf("--max-scan must be in [1,1000]")
	}

	islands, err := scanIslands(ctx, client, f.MaxScan)
	if err != nil {
		return err
	}

	result := searchResult{}
	result.Meta.Scanned = len(islands)

	for _, it := range islands {
		if !matchesFilters(it, f) {
			continue
		}
		result.Data = append(result.Data, it.IslandMetadata)
	}
	result.Meta.Matched = len(result.Data)
	result.Meta.Truncated = len(islands) == f.MaxScan

	data, err := json.Marshal(result)
	if err != nil {
		return err
	}
	w := &output.Writer{Out: stdout, Format: formatFrom(gf)}
	return w.WriteRaw(data)
}

func matchesFilters(it api.IslandListItem, f searchFlags) bool {
	if f.Creator != "" && it.CreatorCode != f.Creator {
		return false
	}
	if f.TitleContains != "" && !strings.Contains(strings.ToLower(it.Title), strings.ToLower(f.TitleContains)) {
		return false
	}
	for _, want := range f.Tags {
		found := false
		for _, got := range it.Tags {
			if got == want {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
