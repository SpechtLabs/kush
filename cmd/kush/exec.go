package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	humane "github.com/sierrasoftworks/humane-errors-go"
	"github.com/spechtlabs/kush/internal/config"
	"github.com/spechtlabs/kush/internal/shell"
	"github.com/spechtlabs/kush/internal/state"
	"github.com/spechtlabs/kush/internal/tempkube"
	"github.com/spechtlabs/kush/pkg/kubeconfig"
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
		return runExec(cmd.Context(), cmd.ErrOrStderr(), ctxName, execNamespace, argv)
	},
}

func init() {
	cmdExec.Flags().StringVarP(&execNamespace, "namespace", "n", "", "namespace to pin for the command")
	// Everything after `--` is the command; cobra passes it through in args.
}

func runExec(ctx context.Context, warnOut io.Writer, ctxName, namespace string, argv []string) error {
	if err := state.GuardNesting(); err != nil {
		return humane.Wrap(err, "cannot run exec", "exit the current kush shell first")
	}
	if dir, err := tempkube.TempDir(); err == nil {
		tempkube.SweepStale(dir)
	}

	cfg, err := resolveLoad(warnOut)
	if err != nil {
		return humane.Wrap(err, "cannot load kubeconfig", "verify your kubeconfig locations with 'kush lint'")
	}

	ctxDef, ok := cfg.Contexts[ctxName]
	if !ok {
		_, err = kubeconfig.Extract(cfg, ctxName, namespace)
		if err != nil {
			return humane.Wrap(err, fmt.Sprintf("cannot isolate context %q", ctxName), "run 'kush lint' to find broken context references")
		}
	}
	hookNamespace := namespace
	if hookNamespace == "" && ok {
		hookNamespace = ctxDef.Namespace
	}
	err = runPreExecHook(ctx, ctxName, hookNamespace)
	if err != nil {
		return err
	}
	if len(config.PreExecHooks(ctxName)) > 0 {
		cfg, err = resolveLoad(warnOut)
		if err != nil {
			return humane.Wrap(err, "cannot reload kubeconfig after pre-exec hook", "check whether the hook changed or removed a configured kubeconfig")
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
		//nolint:gocritic // exitAfterDefer: the temp file is removed explicitly above before os.Exit
		os.Exit(exitErr.ExitCode())
	}
	if err != nil {
		return humane.Wrap(err, "exec failed", "the command could not be run against the isolated context")
	}
	return nil
}
