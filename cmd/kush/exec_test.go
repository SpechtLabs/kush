package main

import (
	"context"
	"fmt"
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

func TestExecRunsPreExecHook(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh not available")
	}

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

	xdgConfig := t.TempDir()
	kushConfigDir := filepath.Join(xdgConfig, "kush")
	if err := os.MkdirAll(kushConfigDir, 0o700); err != nil {
		t.Fatal(err)
	}
	hookOut := filepath.Join(t.TempDir(), "hook.out")
	configContent := fmt.Sprintf("pre_exec_hook: |\n  printf '%%s' \"$KUSH_CONTEXT\" > '%s'\n", hookOut)
	if err := os.WriteFile(filepath.Join(kushConfigDir, "config.yaml"), []byte(configContent), 0o600); err != nil {
		t.Fatal(err)
	}

	runtime := t.TempDir()
	cmd := exec.CommandContext(context.Background(),
		"go", "run", "../../cmd/kush", "exec", "prod", "--", "sh", "-c", "printf ok")
	cmd.Env = append(os.Environ(),
		"KUBECONFIG="+kubeconfig,
		"XDG_CONFIG_HOME="+xdgConfig,
		"XDG_RUNTIME_DIR="+runtime,
		"SHELL=/bin/sh",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("kush exec failed: %v\n%s", err, out)
	}
	if got := string(out); got != "ok" {
		t.Fatalf("command output = %q, want ok", got)
	}
	hookBytes, err := os.ReadFile(hookOut)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(hookBytes); got != "prod" {
		t.Fatalf("hook output = %q, want prod", got)
	}
}
