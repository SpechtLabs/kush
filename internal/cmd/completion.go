package cmd

import (
	"strings"

	"github.com/spechtlabs/kush/internal/kubeconfig"
	"github.com/spf13/cobra"
)

// completeContexts is a cobra ValidArgsFunction that completes the first
// positional context argument with the context names discovered in the user's
// kubeconfig(s). It only completes the first positional; once a context is
// given it returns no completions. Any load error yields no completions rather
// than a crash, so tab-completion degrades quietly.
//
// TODO(phase5): source contexts from the configured lookup locations via
// resolveLoad once the config-discovery feature lands.
func completeContexts(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) >= 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	cfg, err := kubeconfig.Load()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var out []string
	for _, name := range kubeconfig.Contexts(cfg) {
		if strings.HasPrefix(name, toComplete) {
			out = append(out, name)
		}
	}
	return out, cobra.ShellCompDirectiveNoFileComp
}
