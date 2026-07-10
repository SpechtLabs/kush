package kubeconfig

import (
	"fmt"
	"os"

	"k8s.io/client-go/tools/clientcmd/api"
)

const (
	LevelError = "error"
	LevelWarn  = "warning"
)

// Finding is one lint result.
type Finding struct {
	Level   string
	Message string
}

// Lint runs read-only checks over the merged config and its source files.
func Lint(cfg *api.Config, files []string) []Finding {
	var out []Finding

	if cfg.CurrentContext == "" {
		out = append(out, Finding{LevelWarn, "current-context is empty"})
	} else if _, ok := cfg.Contexts[cfg.CurrentContext]; !ok {
		out = append(out, Finding{LevelError, fmt.Sprintf("current-context %q does not exist", cfg.CurrentContext)})
	}

	for name, kctx := range cfg.Contexts {
		if kctx.Cluster == "" {
			out = append(out, Finding{LevelError, fmt.Sprintf("context %q has no cluster", name)})
		} else if _, ok := cfg.Clusters[kctx.Cluster]; !ok {
			out = append(out, Finding{LevelError, fmt.Sprintf("context %q references missing cluster %q", name, kctx.Cluster)})
		}
		if kctx.AuthInfo == "" {
			out = append(out, Finding{LevelWarn, fmt.Sprintf("context %q has no user", name)})
		} else if _, ok := cfg.AuthInfos[kctx.AuthInfo]; !ok {
			out = append(out, Finding{LevelError, fmt.Sprintf("context %q references missing user %q", name, kctx.AuthInfo)})
		}
	}

	for _, f := range files {
		if f == "" {
			continue
		}
		if _, err := os.Stat(f); err != nil {
			out = append(out, Finding{LevelWarn, fmt.Sprintf("kubeconfig file entry is unreachable: %s", f)})
		}
	}

	return out
}
