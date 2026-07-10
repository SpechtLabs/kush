package kubeconfig

import (
	"os"
	"path/filepath"

	humane "github.com/sierrasoftworks/humane-errors-go"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// Split writes one self-contained kubeconfig per context into dir (created with
// mode 0700; files 0600) using extract-one semantics. The source config is
// never mutated. Returns the written file paths.
func Split(cfg *api.Config, dir string) ([]string, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, humane.Wrap(err, "failed to create split dir "+dir, "check write permissions on the parent directory")
	}

	var written []string
	for _, name := range Contexts(cfg) {
		out, err := Extract(cfg, name, "")
		if err != nil {
			return written, humane.Wrap(err, "failed to split context "+name, "run `kush lint` to find broken references")
		}
		path := filepath.Join(dir, sanitizeContext(name)+".yaml")
		if err := clientcmd.WriteToFile(*out, path); err != nil {
			return written, humane.Wrap(err, "failed to write "+path, "check available space and write permissions in the output dir")
		}
		written = append(written, path)
	}
	return written, nil
}
