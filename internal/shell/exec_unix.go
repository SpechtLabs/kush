//go:build unix

package shell

import (
	"context"
	"errors"
	"os"
	"os/exec"

	humane "github.com/sierrasoftworks/humane-errors-go"
)

// Exec runs argv with KUBECONFIG=kubeconfig and extraEnv appended, forwarding
// stdio. If the command exits non-zero, the returned error IS the raw
// *exec.ExitError (not wrapped) so the caller can read and propagate the code.
func Exec(ctx context.Context, kubeconfig string, extraEnv, argv []string) error {
	if len(argv) == 0 {
		return humane.New("no command given to exec", "pass a command after `--`, e.g. kush exec prod -- kubectl get pods")
	}
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "KUBECONFIG="+kubeconfig)
	cmd.Env = append(cmd.Env, extraEnv...)

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return err // raw, so runExec can read ExitCode()
		}
		return humane.Wrap(err, "failed to run "+argv[0], "check the command exists on PATH")
	}
	return nil
}
