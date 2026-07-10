package cmd

import (
	"fmt"

	"github.com/spechtlabs/kush/internal/state"
	"github.com/spf13/cobra"
)

var cmdCurrent = &cobra.Command{
	Use:   "current",
	Short: "Print the active kush context/namespace (empty if not in a kush shell)",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		st, ok := state.Current()
		if !ok {
			return nil // not in a kush shell: print nothing, exit 0
		}
		_, err := fmt.Fprintf(cmd.OutOrStdout(), "%s (%s)\n", st.Context, st.Namespace)
		return err
	},
}
