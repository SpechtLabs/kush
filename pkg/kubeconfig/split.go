package kubeconfig

import (
	"os"
	"path/filepath"
	"strconv"

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

	used := make(map[string]bool)
	var written []string
	for _, name := range Contexts(cfg) {
		out, err := Extract(cfg, name, "")
		if err != nil {
			return written, humane.Wrap(err, "failed to split context "+name, "run `kush lint` to find broken references")
		}
		base := uniqueBase(sanitizeContext(name), used)
		path := filepath.Join(dir, base+".yaml")
		if err := clientcmd.WriteToFile(*out, path); err != nil {
			return written, humane.Wrap(err, "failed to write "+path, "check available space and write permissions in the output dir")
		}
		written = append(written, path)
	}
	return written, nil
}

// uniqueBase returns base, or base disambiguated with a "-2", "-3", ... suffix
// if it collides with a filename already used in this Split run, and records
// whichever name it returns as used.
func uniqueBase(base string, used map[string]bool) string {
	candidate := base
	for n := 2; used[candidate]; n++ {
		candidate = base + "-" + strconv.Itoa(n)
	}
	used[candidate] = true
	return candidate
}
