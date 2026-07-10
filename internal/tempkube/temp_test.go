package tempkube

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// sampleConfig returns a minimal single-context kubeconfig for exercising the
// temp-file writer. tempkube stays independent of pkg/kubeconfig on purpose.
func sampleConfig() *api.Config {
	cfg := api.NewConfig()
	cfg.Clusters["prod-cluster"] = &api.Cluster{Server: "https://prod:6443"}
	cfg.AuthInfos["prod-user"] = &api.AuthInfo{Token: "prod-token"}
	cfg.Contexts["prod"] = &api.Context{Cluster: "prod-cluster", AuthInfo: "prod-user", Namespace: "web"}
	cfg.CurrentContext = "prod"
	return cfg
}

func TestTempDirUsesXDGRuntime(t *testing.T) {
	base := t.TempDir()
	t.Setenv("XDG_RUNTIME_DIR", base)
	dir, err := TempDir()
	if err != nil {
		t.Fatalf("TempDir() error = %v", err)
	}
	if dir != filepath.Join(base, "kush") {
		t.Fatalf("TempDir() = %q, want %q", dir, filepath.Join(base, "kush"))
	}
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("stat temp dir: %v", err)
	}
	if info.Mode().Perm() != 0o700 {
		t.Fatalf("temp dir mode = %o, want 700", info.Mode().Perm())
	}
}

func TestTempDirEnforcesMode(t *testing.T) {
	base := t.TempDir()
	t.Setenv("XDG_RUNTIME_DIR", base)
	dir := filepath.Join(base, "kush")
	if err := os.MkdirAll(dir, 0o777); err != nil {
		t.Fatalf("pre-create dir: %v", err)
	}

	if _, err := TempDir(); err != nil {
		t.Fatalf("TempDir() error = %v", err)
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("stat temp dir: %v", err)
	}
	if info.Mode().Perm() != 0o700 {
		t.Fatalf("temp dir mode = %o, want 700", info.Mode().Perm())
	}
}

func TestWriteTemp(t *testing.T) {
	base := t.TempDir()
	t.Setenv("XDG_RUNTIME_DIR", base)

	path, err := WriteTemp(sampleConfig(), "prod")
	if err != nil {
		t.Fatalf("WriteTemp() error = %v", err)
	}
	defer os.Remove(path)

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat temp file: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("temp file mode = %o, want 600", info.Mode().Perm())
	}
	if !strings.HasPrefix(filepath.Base(path), "prod-") {
		t.Fatalf("temp file name = %q, want prefix prod-", filepath.Base(path))
	}

	// Round-trips as a valid kubeconfig.
	reloaded, err := clientcmd.LoadFromFile(path)
	if err != nil {
		t.Fatalf("reload temp file: %v", err)
	}
	if reloaded.CurrentContext != "prod" {
		t.Fatalf("reloaded current-context = %q, want prod", reloaded.CurrentContext)
	}
}

func TestWriteTempSanitizesContextName(t *testing.T) {
	base := t.TempDir()
	t.Setenv("XDG_RUNTIME_DIR", base)
	cfg := sampleConfig()
	cfg.Contexts["team/prod"] = cfg.Contexts["prod"]
	path, err := WriteTemp(cfg, "team/prod")
	if err != nil {
		t.Fatalf("WriteTemp() error = %v", err)
	}
	defer os.Remove(path)
	if strings.ContainsRune(filepath.Base(path), '/') {
		t.Fatalf("temp file name contains path separator: %q", filepath.Base(path))
	}
}
