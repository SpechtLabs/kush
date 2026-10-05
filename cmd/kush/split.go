package main

import (
	"fmt"
	"os"
	"path/filepath"

	humane "github.com/sierrasoftworks/humane-errors-go"
	"github.com/spechtlabs/kush/pkg/kubeconfig"
	"github.com/spf13/cobra"
)

// newSplitCmd builds `kush split`, which writes one kubeconfig per context.
func newSplitCmd() *cobra.Command {
	var outDir string

	splitCmd := &cobra.Command{
		Use:   "split [-o dir]",
		Short: "Split a monolithic kubeconfig into one self-contained file per context",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := outDir
			if dir == "" {
				home, err := os.UserHomeDir()
				if err != nil {
					return humane.Wrap(err, "failed to resolve home dir", "pass an explicit output dir with -o <dir>")
				}
				dir = filepath.Join(home, ".kube", "kush")
			}

			cfg, herr := resolveLoad(cmd.ErrOrStderr())
			if herr != nil {
				return herr
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
	splitCmd.Flags().StringVarP(&outDir, "out", "o", "", "output directory (default ~/.kube/kush)")
	return splitCmd
}
