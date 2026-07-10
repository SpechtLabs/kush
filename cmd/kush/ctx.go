package main

import (
	"context"
	"fmt"
	"io"
	"os"

	humane "github.com/sierrasoftworks/humane-errors-go"
	"github.com/spechtlabs/kush/internal/config"
	"github.com/spechtlabs/kush/internal/picker"
	"github.com/spechtlabs/kush/internal/shell"
	"github.com/spechtlabs/kush/internal/state"
	"github.com/spechtlabs/kush/internal/tempkube"
	"github.com/spechtlabs/kush/pkg/kubeconfig"
	"github.com/spf13/cobra"
)

var cmdCtx = &cobra.Command{
	Use:               "ctx [name]",
	Short:             "Enter an isolated subshell pinned to a context (no arg opens the picker)",
	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: completeContexts,
	RunE: func(cmd *cobra.Command, args []string) error {
		if list, _ := cmd.Flags().GetBool("list"); list {
			return listContexts(cmd)
		}
		name := ""
		if len(args) == 1 {
			name = args[0]
		}
		return runCtx(cmd.Context(), cmd.ErrOrStderr(), name, "")
	},
}

func init() {
	cmdCtx.Flags().BoolP("list", "l", false, "list all discovered contexts and exit (no subshell)")
}

// listContexts prints every context discovered across the configured lookup
// locations, one per line, marking the current context. It is also the quickest
// way to see exactly which contexts kush's config resolves to.
func listContexts(cmd *cobra.Command) error {
	cfg, err := resolveLoad(cmd.ErrOrStderr())
	if err != nil {
		return humane.Wrap(err, "cannot list contexts", "verify your kubeconfig locations with 'kush lint'")
	}
	out := cmd.OutOrStdout()
	for _, name := range kubeconfig.Contexts(cfg) {
		if name == cfg.CurrentContext {
			_, _ = fmt.Fprintf(out, "%s (current)\n", name)
			continue
		}
		_, _ = fmt.Fprintln(out, name)
	}
	return nil
}

// runCtx enters an isolated subshell for ctxName at the optional namespace.
// An empty ctxName means "open the picker".
func runCtx(ctx context.Context, warnOut io.Writer, ctxName, namespace string) error {
	if err := state.GuardNesting(); err != nil {
		return humane.Wrap(err, "cannot enter a context", "exit the current kush shell first")
	}

	// Opportunistic stale-file cleanup; never blocks the invocation.
	if dir, err := tempkube.TempDir(); err == nil {
		tempkube.SweepStale(dir)
	}

	cfg, err := resolveLoad(warnOut)
	if err != nil {
		return humane.Wrap(err, "cannot load kubeconfig", "verify your kubeconfig locations with 'kush lint'")
	}

	if ctxName == "" {
		names := kubeconfig.Contexts(cfg)
		if len(names) == 0 {
			return humane.New("no contexts found in KUBECONFIG", "check that KUBECONFIG points at a kubeconfig with at least one context")
		}
		var mode picker.Mode
		mode, err = pickerMode()
		if err != nil {
			return humane.Wrap(err, "cannot determine the context picker", "check the 'picker' config value or KUSH_PICKER")
		}
		ctxName, err = picker.Select(ctx, mode, "kush ctx> ", names)
		if err != nil {
			return humane.Wrap(err, "context selection failed", "pick a context or pass one as an argument")
		}
	}

	out, err := kubeconfig.Extract(cfg, ctxName, namespace)
	if err != nil {
		return humane.Wrap(err, fmt.Sprintf("cannot isolate context %q", ctxName), "run 'kush lint' to find broken context references")
	}

	path, err := tempkube.WriteTemp(out, ctxName)
	if err != nil {
		return humane.Wrap(err, "cannot write the temporary kubeconfig", "check space and permissions in $XDG_RUNTIME_DIR")
	}
	defer func() { _ = os.Remove(path) }()

	st := state.State{
		Context:    ctxName,
		Namespace:  out.Contexts[ctxName].Namespace,
		Kubeconfig: path,
	}
	return shell.Run(ctx, config.Shell(), path, st.Env())
}
