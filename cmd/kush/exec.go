package main

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"

	humane "github.com/sierrasoftworks/humane-errors-go"
	"github.com/spechtlabs/kush/internal/shell"
	"github.com/spf13/cobra"
)

// newExecCmd builds `kush exec`, which runs one command against an isolated
// context.
func newExecCmd() *cobra.Command {
	var namespace string

	execCmd := &cobra.Command{
		Use:               "exec <context> [-n namespace] -- <command> [args...]",
		Short:             "Run one command against an isolated context, no interactive shell",
		Args:              cobra.MinimumNArgs(2),
		ValidArgsFunction: completeContexts,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctxName := args[0]
			argv := args[1:]
			return runExec(cmd.Context(), cmd.ErrOrStderr(), ctxName, namespace, argv)
		},
	}
	execCmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace to pin for the command")
	// Everything after `--` is the command; cobra passes it through in args.
	return execCmd
}

func runExec(ctx context.Context, warnOut io.Writer, ctxName, namespace string, argv []string) humane.Error {
	cfg, herr := prepareIsolation(warnOut, "cannot run exec")
	if herr != nil {
		return herr
	}

	path, st, herr := isolate(ctx, warnOut, cfg, ctxName, namespace)
	if herr != nil {
		return herr
	}
	// On SIGINT (Ctrl-C) the parent process dies before this defer runs, so
	// cleanup falls to SweepStale on the next invocation (the sweep is the
	// safety net for exactly this case, not just crashes).
	defer func() { _ = os.Remove(path) }()

	err := shell.Exec(ctx, path, st.Env(), argv)

	// Propagate the child's exit code.
	if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
		_ = os.Remove(path) // os.Exit skips the deferred cleanup; delete the creds now
		//nolint:gocritic // exitAfterDefer: the temp file is removed explicitly above before os.Exit
		os.Exit(exitErr.ExitCode())
	}
	if err != nil {
		return humane.Wrap(err, "exec failed", "the command could not be run against the isolated context")
	}
	return nil
}
