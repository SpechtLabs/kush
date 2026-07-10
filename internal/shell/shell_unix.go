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
	"golang.org/x/sys/unix"
	"golang.org/x/term"
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

// setupSignalHandler forwards or acts on catchable signals for clean teardown.
// Lifted from tka's shell handler minus credential revocation.
func setupSignalHandler(ctx context.Context, cancel context.CancelFunc, cmd *exec.Cmd) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs,
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP,
		syscall.SIGABRT, syscall.SIGFPE, syscall.SIGILL, syscall.SIGSEGV, syscall.SIGBUS,
		syscall.SIGPIPE, syscall.SIGTSTP, syscall.SIGTTIN, syscall.SIGTTOU,
		syscall.SIGWINCH, syscall.SIGURG,
	)

	go func() {
		defer signal.Stop(sigs)
		for {
			select {
			case <-ctx.Done():
				return
			case sig := <-sigs:
				switch sig {
				case syscall.SIGWINCH:
					resizeChild(cmd)
				case syscall.SIGTSTP, syscall.SIGTTIN, syscall.SIGTTOU, syscall.SIGURG:
					if isRunning(cmd) {
						_ = cmd.Process.Signal(sig)
					}
				default:
					// Termination / fatal / pipe signals: tear the child down.
					terminateChild(cmd, cancel)
					return
				}
			}
		}
	}()
}

func resizeChild(cmd *exec.Cmd) {
	if !isRunning(cmd) {
		return
	}
	if fd := int(os.Stdin.Fd()); term.IsTerminal(fd) {
		if w, h, err := term.GetSize(fd); err == nil {
			_ = unix.IoctlSetWinsize(cmd.Process.Pid, syscall.TIOCSWINSZ, &unix.Winsize{
				Row: uint16(h), Col: uint16(w),
			})
		}
	}
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
