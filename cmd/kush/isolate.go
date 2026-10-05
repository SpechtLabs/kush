package main

import (
	"context"
	"fmt"
	"io"

	humane "github.com/sierrasoftworks/humane-errors-go"
	"github.com/spechtlabs/kush/internal/config"
	"github.com/spechtlabs/kush/internal/state"
	"github.com/spechtlabs/kush/internal/tempkube"
	"github.com/spechtlabs/kush/pkg/kubeconfig"
	"k8s.io/client-go/tools/clientcmd/api"
)

// prepareIsolation is what every command that isolates a context does first:
// refuse to run inside a kush shell (refusal is the message for that), sweep
// stale temp kubeconfigs, and load the merged kubeconfig.
func prepareIsolation(warnOut io.Writer, refusal string) (*api.Config, humane.Error) {
	if err := state.GuardNesting(); err != nil {
		return nil, humane.Wrap(err, refusal, "exit the current kush shell first")
	}

	// Opportunistic stale-file cleanup; never blocks the invocation.
	if dir, err := tempkube.TempDir(); err == nil {
		tempkube.SweepStale(dir)
	}

	cfg, err := resolveLoad(warnOut)
	if err != nil {
		return nil, humane.Wrap(err, "cannot load kubeconfig", "verify your kubeconfig locations with 'kush lint'")
	}
	return cfg, nil
}

// isolate runs ctxName's pre-exec hooks, then writes the context, pinned to
// namespace when one is given, to a private temp kubeconfig. It returns the
// file's path, which the caller removes, and the state the child process
// gets. When there are hooks it reloads the kubeconfig after them, because a
// hook may have changed it.
func isolate(ctx context.Context, warnOut io.Writer, cfg *api.Config, ctxName, namespace string) (string, state.State, humane.Error) {
	hookNS, err := hookNamespace(cfg, ctxName, namespace)
	if err != nil {
		return "", state.State{}, err
	}
	if err = runPreExecHook(ctx, ctxName, hookNS); err != nil {
		return "", state.State{}, err
	}
	if len(config.PreExecHooks(ctxName)) > 0 {
		if cfg, err = resolveLoad(warnOut); err != nil {
			return "", state.State{}, humane.Wrap(err, "cannot reload kubeconfig after pre-exec hook", "check whether the hook changed or removed a configured kubeconfig")
		}
	}

	out, extractErr := kubeconfig.Extract(cfg, ctxName, namespace)
	if extractErr != nil {
		return "", state.State{}, humane.Wrap(extractErr, fmt.Sprintf("cannot isolate context %q", ctxName), "run 'kush lint' to find broken context references")
	}
	path, err := tempkube.WriteTemp(out, ctxName)
	if err != nil {
		return "", state.State{}, humane.Wrap(err, "cannot write the temporary kubeconfig", "check space and permissions in $XDG_RUNTIME_DIR")
	}

	st := state.State{
		Context:    ctxName,
		Namespace:  out.Contexts[ctxName].Namespace,
		Kubeconfig: path,
	}
	return path, st, nil
}

// hookNamespace is the namespace the pre-exec hooks see: the requested one,
// or else the context's own. A context that doesn't exist fails here, before
// any hook runs.
func hookNamespace(cfg *api.Config, ctxName, namespace string) (string, humane.Error) {
	ctxDef, ok := cfg.Contexts[ctxName]
	if !ok {
		if _, err := kubeconfig.Extract(cfg, ctxName, namespace); err != nil {
			return "", humane.Wrap(err, fmt.Sprintf("cannot isolate context %q", ctxName), "run 'kush lint' to find broken context references")
		}
	}
	if namespace == "" && ok {
		return ctxDef.Namespace, nil
	}
	return namespace, nil
}
