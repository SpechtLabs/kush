package cmd

import (
	"context"
	"os"

	humane "github.com/sierrasoftworks/humane-errors-go"
	"github.com/spechtlabs/kush/internal/kubeconfig"
	"github.com/spechtlabs/kush/internal/shell"
	"github.com/spechtlabs/kush/internal/state"
	"github.com/spf13/cobra"
)

var cmdCtx = &cobra.Command{
	Use:   "ctx [name]",
	Short: "Enter an isolated subshell pinned to a context (no arg opens the picker)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := ""
		if len(args) == 1 {
			name = args[0]
		}
		return runCtx(cmd.Context(), name, "")
	},
}

// runCtx enters an isolated subshell for ctxName at the optional namespace.
// An empty ctxName means "use the picker" (wired in Phase 2).
func runCtx(ctx context.Context, ctxName, namespace string) error {
	if err := state.GuardNesting(); err != nil {
		return err
	}

	// Opportunistic stale-file cleanup; never blocks the invocation.
	if dir, err := kubeconfig.TempDir(); err == nil {
		kubeconfig.SweepStale(dir)
	}

	cfg, err := kubeconfig.Load()
	if err != nil {
		return err
	}

	if ctxName == "" {
		ctxName, err = pickContext(ctx, cfg) // Phase 1: returns an error; replaced in Phase 2.
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

// pickContext is replaced by the real picker in Phase 2. Phase 1 requires an
// explicit context argument.
func pickContext(ctx context.Context, cfg interface{}) (string, error) {
	return "", humane.New("no context given", "run `kush <context>` (interactive picker is added in Phase 2)")
}
