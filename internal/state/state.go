// Package state reads and writes the KUSH_* environment variables that mark an
// active kush subshell and drives the one-env-var nesting guard.
package state

import (
	"fmt"
	"os"

	humane "github.com/sierrasoftworks/humane-errors-go"
)

const (
	EnvActive     = "KUSH_ACTIVE"
	EnvContext    = "KUSH_CONTEXT"
	EnvNamespace  = "KUSH_NAMESPACE"
	EnvKubeconfig = "KUSH_KUBECONFIG"

	activeValue = "1"
)

// State is the identity of an active kush subshell.
type State struct {
	Context    string
	Namespace  string
	Kubeconfig string
}

// Active reports whether the current process runs inside a kush subshell.
func Active() bool {
	return os.Getenv(EnvActive) == activeValue
}

// Current returns the active subshell's state, or ok=false when not in a kush shell.
func Current() (State, bool) {
	if !Active() {
		return State{}, false
	}
	return State{
		Context:    os.Getenv(EnvContext),
		Namespace:  os.Getenv(EnvNamespace),
		Kubeconfig: os.Getenv(EnvKubeconfig),
	}, true
}

// GuardNesting returns an error if we are already inside a kush subshell.
func GuardNesting() error {
	if Active() {
		return humane.New(
			fmt.Sprintf("already in a kush shell (%s)", os.Getenv(EnvContext)),
			"exit the current shell first, then enter the other context",
		)
	}
	return nil
}

// Env renders the state as KUBECONFIG-adjacent env entries for the child shell.
func (s State) Env() []string {
	return []string{
		EnvActive + "=" + activeValue,
		EnvContext + "=" + s.Context,
		EnvNamespace + "=" + s.Namespace,
		EnvKubeconfig + "=" + s.Kubeconfig,
	}
}
