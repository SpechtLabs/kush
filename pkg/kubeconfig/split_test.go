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

func TestSplitFilenameCollision(t *testing.T) {
	dir := t.TempDir()
	cfg := sampleConfig()
	// "team/prod" and "team_prod" both sanitize to "team_prod"; neither may
	// overwrite the other's file.
	cfg.Contexts["team/prod"] = cfg.Contexts["prod"]
	cfg.Contexts["team_prod"] = cfg.Contexts["dev"]

	paths, err := Split(cfg, dir)
	if err != nil {
		t.Fatalf("Split() error = %v", err)
	}
	if len(paths) != 4 {
		t.Fatalf("Split() wrote %d files, want 4", len(paths))
	}

	first, err := clientcmd.LoadFromFile(filepath.Join(dir, "team_prod.yaml"))
	if err != nil {
		t.Fatalf("load team_prod.yaml: %v", err)
	}
	second, err := clientcmd.LoadFromFile(filepath.Join(dir, "team_prod-2.yaml"))
	if err != nil {
		t.Fatalf("load team_prod-2.yaml: %v", err)
	}

	// "team/prod" sorts before "team_prod" ('/' < '_'), so it claims the
	// unsuffixed name and "team_prod" gets the "-2" suffix.
	if first.CurrentContext != "team/prod" {
		t.Fatalf("team_prod.yaml current-context = %q, want team/prod", first.CurrentContext)
	}
	if second.CurrentContext != "team_prod" {
		t.Fatalf("team_prod-2.yaml current-context = %q, want team_prod", second.CurrentContext)
	}
}
