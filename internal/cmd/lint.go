package cmd

import (
	"fmt"

	humane "github.com/sierrasoftworks/humane-errors-go"
	"github.com/spechtlabs/kush/internal/kubeconfig"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

var cmdLint = &cobra.Command{
	Use:   "lint",
	Short: "Check KUBECONFIG(s) for common problems",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		rules := clientcmd.NewDefaultClientConfigLoadingRules()
		cfg, err := kubeconfig.LoadFrom(rules)
		if err != nil {
			return err
		}

		findings := kubeconfig.Lint(cfg, rules.Precedence)
		out := cmd.OutOrStdout()
		if len(findings) == 0 {
			_, err := fmt.Fprintln(out, "ok: no problems found")
			return err
		}

		hasError := false
		for _, f := range findings {
			if f.Level == kubeconfig.LevelError {
				hasError = true
			}
			if _, err := fmt.Fprintf(out, "%s: %s\n", f.Level, f.Message); err != nil {
				return err
			}
		}
		if hasError {
			return humane.New("kubeconfig lint found errors", "fix the reported problems above, then re-run `kush lint`")
		}
		return nil
	},
}
