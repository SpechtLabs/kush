package kubeconfig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

func writeCfg(t *testing.T, path, ctx, cluster, user string) {
	t.Helper()
	cfg := api.NewConfig()
	cfg.Clusters[cluster] = &api.Cluster{Server: "https://" + cluster + ":6443"}
	cfg.AuthInfos[user] = &api.AuthInfo{Token: user}
	cfg.Contexts[ctx] = &api.Context{Cluster: cluster, AuthInfo: user}
	cfg.CurrentContext = ctx
	if err := clientcmd.WriteToFile(*cfg, path); err != nil {
		t.Fatal(err)
	}
}

func TestExpandLocationsGlobAndTilde(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.yaml")
	b := filepath.Join(dir, "b.yaml")
	writeCfg(t, a, "actx", "ac", "au")
	writeCfg(t, b, "bctx", "bc", "bu")
	// non-yaml file must not match the glob
	if err := os.WriteFile(filepath.Join(dir, "note.txt"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := ExpandLocations([]string{filepath.Join(dir, "*.yaml")})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d files, want 2: %v", len(got), got)
	}
}

func TestExpandLocationsKubeconfigSplitAndDedupe(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.yaml")
	b := filepath.Join(dir, "b.yaml")
	writeCfg(t, a, "actx", "ac", "au")
	writeCfg(t, b, "bctx", "bc", "bu")
	t.Setenv("KUBECONFIG", a+string(os.PathListSeparator)+b)

	// $KUBECONFIG splits into two; a.yaml listed again → deduped; order preserved.
	got, err := ExpandLocations([]string{"$KUBECONFIG", a})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != a || got[1] != b {
		t.Fatalf("got %v, want [%s %s]", got, a, b)
	}
}

func TestExpandLocationsSkipsMissing(t *testing.T) {
	dir := t.TempDir()
	got, err := ExpandLocations([]string{filepath.Join(dir, "nope.yaml"), filepath.Join(dir, "*.yaml")})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("got %v, want empty", got)
	}
}

func TestLoadResolvedNilUsesDefault(t *testing.T) {
	dir := t.TempDir()
	kc := filepath.Join(dir, "config")
	writeCfg(t, kc, "only", "c", "u")
	t.Setenv("KUBECONFIG", kc)

	cfg, warns, err := LoadResolved(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(warns) != 0 {
		t.Fatalf("warnings = %v, want none", warns)
	}
	if _, ok := cfg.Contexts["only"]; !ok {
		t.Fatal("default load did not surface the context")
	}
}

func TestLoadResolvedSkipsInvalidFiles(t *testing.T) {
	dir := t.TempDir()
	good := filepath.Join(dir, "good")
	writeCfg(t, good, "keeper", "c", "u")
	// an empty file (valid: empty kubeconfig) and a junk file (invalid YAML)
	if err := os.WriteFile(filepath.Join(dir, "empty"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "junk"), []byte("not: [valid yaml"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, warns, err := LoadResolved([]string{filepath.Join(dir, "*")})
	if err != nil {
		t.Fatalf("LoadResolved() error = %v; a junk file should be skipped, not fatal", err)
	}
	if _, ok := cfg.Contexts["keeper"]; !ok {
		t.Fatal("valid context was dropped")
	}
	var skipped bool
	for _, w := range warns {
		if strings.Contains(w.Message, "junk") && strings.Contains(w.Message, "skipping") {
			skipped = true
		}
	}
	if !skipped {
		t.Fatalf("expected a skip warning for the junk file, got %v", warns)
	}
}

func TestLoadResolvedMergesFirstWinsWithWarning(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.yaml")
	b := filepath.Join(dir, "b.yaml")
	// both define context "dup"; a is first → wins
	writeCfg(t, a, "dup", "ac", "au")
	writeCfg(t, b, "dup", "bc", "bu")

	cfg, warns, err := LoadResolved([]string{a, b})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Contexts["dup"].Cluster != "ac" {
		t.Fatalf("dup cluster = %q, want ac (first-wins)", cfg.Contexts["dup"].Cluster)
	}
	if len(warns) != 1 || !strings.Contains(warns[0].Message, "dup") || !strings.Contains(warns[0].Message, a) {
		t.Fatalf("warnings = %v, want one naming dup + %s", warns, a)
	}
}
