package kubeconfig

import (
	"os"
	"path/filepath"
	"testing"

	"k8s.io/client-go/tools/clientcmd"
)

func TestSplitRoundTrip(t *testing.T) {
	dir := t.TempDir()
	paths, err := Split(sampleConfig(), dir)
	if err != nil {
		t.Fatalf("Split() error = %v", err)
	}
	if len(paths) != 2 {
		t.Fatalf("Split() wrote %d files, want 2", len(paths))
	}

	// prod.yaml must be self-contained: exactly prod's cluster + user, nothing from dev.
	prod, err := clientcmd.LoadFromFile(filepath.Join(dir, "prod.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if prod.CurrentContext != "prod" {
		t.Fatalf("prod.yaml current-context = %q, want prod", prod.CurrentContext)
	}
	if len(prod.Contexts) != 1 || len(prod.Clusters) != 1 || len(prod.AuthInfos) != 1 {
		t.Fatal("prod.yaml is not self-contained (extra entries leaked)")
	}
	if _, ok := prod.Clusters["dev-cluster"]; ok {
		t.Fatal("dev-cluster leaked into prod.yaml")
	}
	if prod.Contexts["prod"].Namespace != "web" {
		t.Fatal("prod.yaml lost its namespace")
	}
}

func TestSplitSanitizesFilenames(t *testing.T) {
	dir := t.TempDir()
	cfg := sampleConfig()
	cfg.Contexts["team/prod"] = cfg.Contexts["prod"]
	paths, err := Split(cfg, dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("split file missing: %s (%v)", p, err)
		}
	}
}
