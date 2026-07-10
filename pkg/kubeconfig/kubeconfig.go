// Package kubeconfig is the pure, testable core: load/merge kubeconfigs, extract
// a single self-contained context, write private temp copies, edit namespaces,
// lint, and split. It performs no shell, TTY, or Kubernetes API operations.
package kubeconfig

import (
	"fmt"
	"sort"

	humane "github.com/sierrasoftworks/humane-errors-go"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// Load merges all kubeconfigs per kubectl's default loading rules (KUBECONFIG
// colon-list, then ~/.kube/config).
func Load() (*api.Config, error) {
	return LoadFrom(clientcmd.NewDefaultClientConfigLoadingRules())
}

// LoadFrom merges kubeconfigs using the supplied loading rules.
func LoadFrom(rules *clientcmd.ClientConfigLoadingRules) (*api.Config, error) {
	cfg, err := rules.Load()
	if err != nil {
		return nil, humane.Wrap(err, "failed to load kubeconfig", "check that KUBECONFIG points at readable files")
	}
	return cfg, nil
}

// Extract builds a new minimal config containing only ctxName plus the single
// cluster and user it references, with current-context set to ctxName. A
// non-empty namespace overrides the context's namespace. The source config is
// never mutated.
func Extract(cfg *api.Config, ctxName, namespace string) (*api.Config, error) {
	kctx, ok := cfg.Contexts[ctxName]
	if !ok {
		return nil, humane.New(
			fmt.Sprintf("context %q not found", ctxName),
			fmt.Sprintf("available contexts: %v", Contexts(cfg)),
		)
	}

	out := api.NewConfig()
	newCtx := kctx.DeepCopy()
	if namespace != "" {
		newCtx.Namespace = namespace
	}
	out.Contexts[ctxName] = newCtx
	out.CurrentContext = ctxName

	if kctx.Cluster != "" {
		cl, ok := cfg.Clusters[kctx.Cluster]
		if !ok {
			return nil, humane.New(
				fmt.Sprintf("context %q references missing cluster %q", ctxName, kctx.Cluster),
				"run `kush lint` to find broken references in your kubeconfig",
			)
		}
		out.Clusters[kctx.Cluster] = cl.DeepCopy()
	}
	if kctx.AuthInfo != "" {
		au, ok := cfg.AuthInfos[kctx.AuthInfo]
		if !ok {
			return nil, humane.New(
				fmt.Sprintf("context %q references missing user %q", ctxName, kctx.AuthInfo),
				"run `kush lint` to find broken references in your kubeconfig",
			)
		}
		out.AuthInfos[kctx.AuthInfo] = au.DeepCopy()
	}
	return out, nil
}

// Contexts returns the sorted context names in cfg.
func Contexts(cfg *api.Config) []string {
	names := make([]string, 0, len(cfg.Contexts))
	for name := range cfg.Contexts {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
