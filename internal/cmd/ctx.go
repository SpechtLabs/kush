package cmd

import (
	"context"
	"io"
	"os"

	humane "github.com/sierrasoftworks/humane-errors-go"
	"github.com/spechtlabs/kush/internal/kubeconfig"
	"github.com/spechtlabs/kush/internal/picker"
	"github.com/spechtlabs/kush/internal/shell"
	"github.com/spechtlabs/kush/internal/state"
	"github.com/spf13/cobra"
)

var cmdCtx = &cobra.Command{
	Use:               "ctx [name]",
	Short:             "Enter an isolated subshell pinned to a context (no arg opens the picker)",
	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: completeContexts,
	RunE: func(cmd *cobra.Command, args []string) error {
		name := ""
		if len(args) == 1 {
			name = args[0]
		}
		return runCtx(cmd.Context(), cmd.ErrOrStderr(), name, "")
	},
}

// runCtx enters an isolated subshell for ctxName at the optional namespace.
// An empty ctxName means "use the picker" (wired in Phase 2).
func runCtx(ctx context.Context, warnOut io.Writer, ctxName, namespace string) error {
	if err := state.GuardNesting(); err != nil {
		return err
	}

	// Opportunistic stale-file cleanup; never blocks the invocation.
	if dir, err := kubeconfig.TempDir(); err == nil {
		kubeconfig.SweepStale(dir)
	}

	cfg, err := resolveLoad(warnOut)
	if err != nil {
		return err
	}

	if ctxName == "" {
		names := kubeconfig.Contexts(cfg)
		if len(names) == 0 {
			return humane.New("no contexts found in KUBECONFIG", "check that KUBECONFIG points at a kubeconfig with at least one context")
		}
		mode, err := pickerMode()
		if err != nil {
			return err
		}
		ctxName, err = picker.Select(ctx, mode, "kush ctx> ", names)
		if err != nil {
			return err
		}
	}

	out, err := kubeconfig.Extract(cfg, ctxName, namespace)
	if err != nil {
		return err
	}

	path, err := kubeconfig.WriteTemp(out, ctxName)
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(path) }()

	st := state.State{
		Context:    ctxName,
		Namespace:  out.Contexts[ctxName].Namespace,
		Kubeconfig: path,
	}
	return shell.Run(ctx, path, st.Env())
}
