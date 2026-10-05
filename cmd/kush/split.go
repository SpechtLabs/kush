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
				return humane.Wrap(err, "failed to split the kubeconfig into "+dir, "fix the cause below, then re-run `kush split`")
			}
			for _, p := range paths {
				if _, err := fmt.Fprintln(cmd.OutOrStdout(), p); err != nil {
					return humane.Wrap(err, "failed to print the path of "+p, "check that stdout is still open, e.g. that a pipe reader did not exit early")
				}
			}
			return nil
		},
	}
	splitCmd.Flags().StringVarP(&outDir, "out", "o", "", "output directory (default ~/.kube/kush)")
	return splitCmd
}
