package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Build metadata, overridable via -ldflags "-X main.Version=...".
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the kush version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "kush %s (commit %s, built %s)\n", Version, Commit, Date)
			return err
		},
	}
}
