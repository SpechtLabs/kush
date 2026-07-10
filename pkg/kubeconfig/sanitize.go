package kubeconfig

import (
	"os"
	"strings"
)

// sanitizeContext makes a context name safe for use as a single filename
// component, used when Split writes one file per context.
func sanitizeContext(name string) string {
	repl := strings.NewReplacer("/", "_", string(os.PathSeparator), "_", " ", "_")
	return repl.Replace(name)
}
