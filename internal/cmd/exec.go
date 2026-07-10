package cmd

import (
	"context"
	"errors"
	"os"
	"os/exec"

	humane "github.com/sierrasoftworks/humane-errors-go"
	"github.com/spechtlabs/kush/internal/kubeconfig"
	"github.com/spechtlabs/kush/internal/shell"
	"github.com/spechtlabs/kush/internal/state"
	"github.com/spf13/cobra"
)

var execNamespace string

var cmdExec = &cobra.Command{
	Use:               "exec <context> [-n namespace] -- <command> [args...]",
	Short:             "Run one command against an isolated context, no interactive shell",
	Args:              cobra.MinimumNArgs(2),
	ValidArgsFunction: completeContexts,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctxName := args[0]
		argv := args[1:]
		return runExec(cmd.Context(), ctxName, execNamespace, argv)
	},
}

func init() {
	cmdExec.Flags().StringVarP(&execNamespace, "namespace", "n", "", "namespace to pin for the command")
	// Everything after `--` is the command; cobra passes it through in args.
}

func runExec(ctx context.Context, ctxName, namespace string, argv []string) error {
	if err := state.GuardNesting(); err != nil {
		return err
	}
	if dir, err := kubeconfig.TempDir(); err == nil {
		kubeconfig.SweepStale(dir)
	}

	cfg, err := kubeconfig.Load()
	if err != nil {
		return err
	}
	out, err := kubeconfig.Extract(cfg, ctxName, namespace)
	if err != nil {
		return err
	}
	path, err := kubeconfig.WriteTemp(out, ctxName)
	if err != nil {
		return err
	}
	// On SIGINT (Ctrl-C) the parent process dies before this defer runs, so
	// cleanup falls to SweepStale on the next invocation (the sweep is the
	// safety net for exactly this case, not just crashes).
	defer func() { _ = os.Remove(path) }()

	st := state.State{Context: ctxName, Namespace: out.Contexts[ctxName].Namespace, Kubeconfig: path}
	err = shell.Exec(ctx, path, st.Env(), argv)

	// Propagate the child's exit code.
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		_ = os.Remove(path) // os.Exit skips the deferred cleanup; delete the creds now
		os.Exit(exitErr.ExitCode())
	}
	if err != nil {
		return humane.Wrap(err, "exec failed", "the command could not be run against the isolated context")
	}
	return nil
}
