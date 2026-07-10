package kubeconfig

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	humane "github.com/sierrasoftworks/humane-errors-go"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// TempDir resolves and creates the private kush temp directory (mode 0700):
// $XDG_RUNTIME_DIR/kush when set, else os.TempDir()/kush.
func TempDir() (string, error) {
	base := os.Getenv("XDG_RUNTIME_DIR")
	if base == "" {
		base = os.TempDir()
	}
	dir := filepath.Join(base, "kush")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", humane.Wrap(err, "failed to create kush temp dir "+dir, "check filesystem permissions on the parent directory")
	}
	return dir, nil
}

// sanitizeContext makes a context name safe for use inside a single filename
// component. The PID and random suffix stay parseable because they contain no
// '-' collisions the sweep relies on positionally (last two '-'-fields).
func sanitizeContext(name string) string {
	repl := strings.NewReplacer("/", "_", string(os.PathSeparator), "_", " ", "_")
	return repl.Replace(name)
}

// WriteTemp writes cfg to a fresh 0600 file named <ctx>-<pid>-<rand>.yaml inside
// the kush temp dir and returns its path.
func WriteTemp(cfg *api.Config, ctxName string) (string, error) {
	dir, err := TempDir()
	if err != nil {
		return "", err
	}
	pattern := sanitizeContext(ctxName) + "-" + strconv.Itoa(os.Getpid()) + "-*.yaml"
	f, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return "", humane.Wrap(err, "failed to create temp kubeconfig", "check available space and permissions in the kush temp dir")
	}
	path := f.Name()
	if err := f.Close(); err != nil {
		_ = os.Remove(path)
		return "", humane.Wrap(err, "failed to close temp kubeconfig", "the temp file may be left behind; remove it manually: "+path)
	}
	if err := clientcmd.WriteToFile(*cfg, path); err != nil {
		_ = os.Remove(path)
		return "", humane.Wrap(err, "failed to write temp kubeconfig "+path, "check available space in the kush temp dir")
	}
	return path, nil
}
