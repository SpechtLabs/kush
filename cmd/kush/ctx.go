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
	"github.com/spechtlabs/kush/pkg/kubeconfig"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd/api"
)

// newCtxCmd builds `kush ctx`, the explicit form of the bare `kush [context]`.
func newCtxCmd() *cobra.Command {
	ctxCmd := &cobra.Command{
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
	ctxCmd.Flags().BoolP("list", "l", false, "list all discovered contexts and exit (no subshell)")
	return ctxCmd
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
func runCtx(ctx context.Context, warnOut io.Writer, ctxName, namespace string) humane.Error {
	cfg, err := prepareIsolation(warnOut, "cannot enter a context")
	if err != nil {
		return err
	}

	if ctxName == "" {
		if ctxName, err = pickContext(ctx, cfg); err != nil {
			return err
		}
	}

	path, st, err := isolate(ctx, warnOut, cfg, ctxName, namespace)
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(path) }()

	return shell.Run(ctx, config.Shell(), path, st.Env(), config.PostExecHooks(ctxName))
}

// pickContext asks the user to pick one of cfg's contexts with the configured
// picker.
func pickContext(ctx context.Context, cfg *api.Config) (string, humane.Error) {
	names := kubeconfig.Contexts(cfg)
	if len(names) == 0 {
		return "", humane.New("no contexts found in KUBECONFIG", "check that KUBECONFIG points at a kubeconfig with at least one context")
	}
	mode, err := pickerMode()
	if err != nil {
		return "", humane.Wrap(err, "cannot determine the context picker", "check the 'picker' config value or KUSH_PICKER")
	}
	name, err := picker.Select(ctx, mode, "kush ctx> ", names)
	if err != nil {
		return "", humane.Wrap(err, "context selection failed", "pick a context or pass one as an argument")
	}
	return name, nil
}
