//go:build unix

// Package shell forks an interactive subshell (or a single command) with a
// pinned KUBECONFIG and forwards stdio and signals. Unix only.
package shell

import (
	"context"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	humane "github.com/sierrasoftworks/humane-errors-go"
)

func loginShell() string {
	if s := os.Getenv("SHELL"); s != "" {
		return s
	}
	return "/bin/bash"
}

// Run forks the user's login shell with KUBECONFIG set to kubeconfig and
// extraEnv appended, inheriting stdio, and blocks until it exits.
func Run(ctx context.Context, kubeconfig string, extraEnv []string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	cmd := exec.CommandContext(ctx, loginShell())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "KUBECONFIG="+kubeconfig)
	cmd.Env = append(cmd.Env, extraEnv...)

	setupSignalHandler(ctx, cancel, cmd)

	if err := cmd.Run(); err != nil {
		return humane.Wrap(err, "subshell exited with an error", "this usually just reflects the last command's exit code inside the shell")
	}
	return nil
}

// setupSignalHandler tears the subshell down on external termination requests
// (SIGTERM/SIGHUP) and otherwise gets out of the way. The child shell shares our
// foreground process group and controlling tty, so tty-generated signals
// (SIGINT/SIGQUIT/SIGTSTP) are delivered to it directly; we catch them only to
// suppress the parent's default action (which would kill kush and orphan the
// child) and then do nothing.
func setupSignalHandler(ctx context.Context, cancel context.CancelFunc, cmd *exec.Cmd) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs,
		syscall.SIGTERM, syscall.SIGHUP, // external teardown requests
		syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTSTP, // tty signals: swallow, let child handle
	)

	go func() {
		defer signal.Stop(sigs)
		for {
			select {
			case <-ctx.Done():
				return
			case sig := <-sigs:
				switch sig {
				case syscall.SIGTERM, syscall.SIGHUP:
					terminateChild(cmd, cancel)
					return
				default:
					// SIGINT/SIGQUIT/SIGTSTP: the child owns these; do nothing.
				}
			}
		}
	}()
}

func isRunning(cmd *exec.Cmd) bool {
	return cmd != nil && cmd.Process != nil && (cmd.ProcessState == nil || !cmd.ProcessState.Exited())
}

func terminateChild(cmd *exec.Cmd, cancel context.CancelFunc) {
	if !isRunning(cmd) {
		cancel()
		return
	}
	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		_ = cmd.Process.Kill()
		cancel()
		return
	}
	deadline := time.NewTimer(250 * time.Millisecond)
	defer deadline.Stop()
	tick := time.NewTicker(50 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-deadline.C:
			_ = cmd.Process.Kill()
			cancel()
			return
		case <-tick.C:
			if !isRunning(cmd) {
				cancel()
				return
			}
		}
	}
}
