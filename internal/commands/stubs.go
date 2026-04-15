package commands

import "github.com/spf13/cobra"

// Temporary stubs — replaced in subsequent tasks.

func newRetentionCmd(_ *GlobalFlags) *cobra.Command { return &cobra.Command{Use: "retention", Hidden: true} }
func newTopCmd(_ *GlobalFlags) *cobra.Command       { return &cobra.Command{Use: "top", Hidden: true} }
func newSummaryCmd(_ *GlobalFlags) *cobra.Command   { return &cobra.Command{Use: "summary", Hidden: true} }
func newCompareCmd(_ *GlobalFlags) *cobra.Command   { return &cobra.Command{Use: "compare", Hidden: true} }
func newSearchCmd(_ *GlobalFlags) *cobra.Command    { return &cobra.Command{Use: "search", Hidden: true} }
