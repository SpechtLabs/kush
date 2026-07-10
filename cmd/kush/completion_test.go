package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

const twoContextKubeconfig = `apiVersion: v1
kind: Config
clusters:
- name: c1
  cluster:
    server: https://x:6443
users:
- name: u1
  user:
    token: t
contexts:
- name: prod
  context:
    cluster: c1
    user: u1
- name: dev
  context:
    cluster: c1
    user: u1
current-context: dev
`

func writeKubeconfig(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config")
	if err := os.WriteFile(path, []byte(twoContextKubeconfig), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestCompleteContextsAll(t *testing.T) {
	t.Setenv("KUBECONFIG", writeKubeconfig(t))

	got, directive := completeContexts(nil, nil, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("directive = %v, want NoFileComp", directive)
	}
	if len(got) != 2 || got[0] != "dev" || got[1] != "prod" {
		t.Fatalf("completions = %v, want [dev prod]", got)
	}
}

func TestCompleteContextsPrefix(t *testing.T) {
	t.Setenv("KUBECONFIG", writeKubeconfig(t))

	got, _ := completeContexts(nil, nil, "pr")
	if len(got) != 1 || got[0] != "prod" {
		t.Fatalf("completions = %v, want [prod]", got)
	}
}

func TestCompleteContextsOnlyFirstArg(t *testing.T) {
	t.Setenv("KUBECONFIG", writeKubeconfig(t))

	got, directive := completeContexts(nil, []string{"prod"}, "")
	if got != nil {
		t.Fatalf("completions = %v, want nil once a context is given", got)
	}
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("directive = %v, want NoFileComp", directive)
	}
}
