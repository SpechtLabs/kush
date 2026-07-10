package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestExecSetsKubeconfig runs `kush exec <ctx> -- sh -c 'echo $KUBECONFIG'` end
// to end and asserts the child saw a kush temp kubeconfig. Uses go run so it
// exercises the real binary path including temp-file lifecycle.
func TestExecKubeconfigPointsAtTemp(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh not available")
	}

	// Build a minimal kubeconfig with one context.
	kubeconfig := filepath.Join(t.TempDir(), "config")
	content := `apiVersion: v1
kind: Config
clusters:
- name: c1
  cluster:
    server: https://one:6443
users:
- name: u1
  user:
    token: t1
contexts:
- name: prod
  context:
    cluster: c1
    user: u1
current-context: prod
`
	if err := os.WriteFile(kubeconfig, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	runtime := t.TempDir()
	cmd := exec.CommandContext(context.Background(),
		"go", "run", "../../cmd/kush", "exec", "prod", "--", "sh", "-c", "echo $KUBECONFIG")
	cmd.Env = append(os.Environ(), "KUBECONFIG="+kubeconfig, "XDG_RUNTIME_DIR="+runtime)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("kush exec failed: %v\n%s", err, out)
	}
	got := strings.TrimSpace(string(out))
	if !strings.Contains(got, filepath.Join(runtime, "kush")) {
		t.Fatalf("child KUBECONFIG = %q, want a path under %q", got, filepath.Join(runtime, "kush"))
	}
	if !strings.Contains(got, "prod-") {
		t.Fatalf("child KUBECONFIG = %q, want prod-*.yaml", got)
	}
}
