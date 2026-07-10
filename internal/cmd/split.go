package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	humane "github.com/sierrasoftworks/humane-errors-go"
	"github.com/spechtlabs/kush/internal/kubeconfig"
	"github.com/spf13/cobra"
)

var splitOutDir string

var cmdSplit = &cobra.Command{
	Use:   "split [-o dir]",
	Short: "Split a monolithic kubeconfig into one self-contained file per context",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := splitOutDir
		if dir == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return humane.Wrap(err, "failed to resolve home dir", "pass an explicit output dir with -o <dir>")
			}
			dir = filepath.Join(home, ".kube", "kush")
		}

		cfg, err := resolveLoad(cmd.ErrOrStderr())
		if err != nil {
			return err
		}
		paths, err := kubeconfig.Split(cfg, dir)
		if err != nil {
			return err
		}
		for _, p := range paths {
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), p); err != nil {
				return err
			}
		}
		return nil
	},
}

func init() {
	cmdSplit.Flags().StringVarP(&splitOutDir, "out", "o", "", "output directory (default ~/.kube/kush)")
}
