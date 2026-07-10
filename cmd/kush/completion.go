package main

import (
	"strings"

	"github.com/spechtlabs/kush/internal/config"
	"github.com/spechtlabs/kush/internal/kubeconfig"
	"github.com/spf13/cobra"
)

// completeContexts is a cobra ValidArgsFunction that completes the first
// positional context argument with the context names discovered across the
// configured lookup locations (or the kubeconfig defaults when unset). It
// only completes the first positional; once a context is given it returns no
// completions. Any load error yields no completions rather than a crash, and
// duplicate-context warnings are dropped so tab-completion stays clean.
func completeContexts(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) >= 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	cfg, _, err := kubeconfig.LoadResolved(config.LookupLocations())
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
