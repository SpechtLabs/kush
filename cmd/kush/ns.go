package main

import (
	"fmt"

	humane "github.com/sierrasoftworks/humane-errors-go"
	"github.com/spechtlabs/kush/internal/picker"
	"github.com/spechtlabs/kush/internal/state"
	"github.com/spechtlabs/kush/pkg/kubeconfig"
	"github.com/spf13/cobra"
)

var cmdNs = &cobra.Command{
	Use:   "ns [name]",
	Short: "Re-pin the namespace (inside a kush shell) or enter one for the current context",
	Long: "Inside a kush shell, `kush ns <name>` re-pins the namespace in place " +
		"(no new shell). Outside, it enters a subshell for the current context " +
		"at that namespace. With no name it prompts for one.",
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := ""
		if len(args) == 1 {
			name = args[0]
		}
		return runNs(cmd, name)
	},
}

func runNs(cmd *cobra.Command, namespace string) error {
	ctx := cmd.Context()
	if st, inShell := state.Current(); inShell {
		// Inside a kush shell: edit the temp kubeconfig in place. Exempt from the
		// nesting guard — no subshell is spawned.
		if namespace == "" {
			mode, err := pickerMode()
			if err != nil {
				return err
			}
			namespace, err = picker.Prompt(ctx, mode, "kush ns> ")
			if err != nil {
				return err
			}
		}
		if err := kubeconfig.SetNamespace(st.Kubeconfig, namespace); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "namespace re-pinned to %q (kubectl and prompt update from the kubeconfig file)\n", namespace)
		return nil
	}

	// Outside a kush shell: spawn a subshell for the CURRENT context at namespace.
	cfg, err := resolveLoad(cmd.ErrOrStderr())
	if err != nil {
		return err
	}
	if cfg.CurrentContext == "" {
		return humane.New("no current context set", "run `kush <context>` or `kubectl config use-context <ctx>` first")
	}
	if namespace == "" {
		mode, err := pickerMode()
		if err != nil {
			return err
		}
		namespace, err = picker.Prompt(ctx, mode, "kush ns> ")
		if err != nil {
			return err
		}
	}
	return runCtx(ctx, cmd.ErrOrStderr(), cfg.CurrentContext, namespace)
}
