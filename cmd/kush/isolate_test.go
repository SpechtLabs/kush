package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spechtlabs/kush/internal/config"
	"github.com/spechtlabs/kush/internal/state"
	"github.com/spf13/viper"
	"k8s.io/client-go/tools/clientcmd"
)

// namespacedKubeconfig has a context with a namespace of its own (prod) and
// one without (dev), and dev as the current context.
const namespacedKubeconfig = `apiVersion: v1
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
    namespace: payments
- name: dev
  context:
    cluster: c1
    user: u1
current-context: dev
`

// isolatedEnv points kush at a fresh kubeconfig holding content, a private
// runtime dir and an empty kush config (in an empty home), outside any kush shell, with
// /bin/sh as the shell, and resets viper afterwards. It returns the runtime
// dir's kush subdirectory, where the temp kubeconfigs go.
func isolatedEnv(t *testing.T, content string) string {
	t.Helper()
	kubeconfig := filepath.Join(t.TempDir(), "config")
	if err := os.WriteFile(kubeconfig, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	runtime := t.TempDir()
	t.Setenv("KUBECONFIG", kubeconfig)
	t.Setenv("XDG_RUNTIME_DIR", runtime)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	t.Setenv("SHELL", "/bin/sh")
	t.Setenv(state.EnvActive, "")
	t.Cleanup(viper.Reset)
	return filepath.Join(runtime, "kush")
}

// truePath returns the path of true(1), which the tests use as a subshell
// that exits at once.
func truePath(t *testing.T) string {
	t.Helper()
	path, err := exec.LookPath("true")
	if err != nil {
		t.Skip("true not available")
	}
	return path
}

func TestHookNamespace(t *testing.T) {
	cfg, err := clientcmd.Load([]byte(namespacedKubeconfig))
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		ctxName   string
		namespace string
		want      string
		wantErr   string
	}{
		{name: "the context's own namespace", ctxName: "prod", want: "payments"},
		{name: "the requested namespace wins", ctxName: "prod", namespace: "audit", want: "audit"},
		{name: "a context without a namespace", ctxName: "dev", want: ""},
		{name: "an unknown context fails", ctxName: "staging", wantErr: `cannot isolate context "staging"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := hookNamespace(cfg, tt.ctxName, tt.namespace)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("hookNamespace() error = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("hookNamespace() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPrepareIsolationRefusesInsideAKushShell(t *testing.T) {
	isolatedEnv(t, namespacedKubeconfig)
	t.Setenv(state.EnvActive, "1")

	_, err := prepareIsolation(&bytes.Buffer{}, "cannot do the thing")
	if err == nil || !strings.Contains(err.Error(), "cannot do the thing") {
		t.Fatalf("prepareIsolation() error = %v, want the refusal", err)
	}
}

func TestIsolate(t *testing.T) {
	tests := []struct {
		name          string
		ctxName       string
		namespace     string
		preExecHooks  []string
		wantNamespace string
		wantHookOut   string
		wantErr       string
	}{
		{name: "the context at its own namespace", ctxName: "prod", wantNamespace: "payments"},
		{name: "pinned to the requested namespace", ctxName: "prod", namespace: "audit", wantNamespace: "audit"},
		{
			name:          "pre-exec hooks run with the context and namespace",
			ctxName:       "prod",
			preExecHooks:  []string{`printf '%s/%s' "$KUSH_CONTEXT" "$KUSH_NAMESPACE" > "$HOOK_OUT"`},
			wantNamespace: "payments",
			wantHookOut:   "prod/payments",
		},
		{name: "a failing pre-exec hook stops it", ctxName: "prod", preExecHooks: []string{"exit 3"}, wantErr: "pre-exec hook 1 failed"},
		{name: "an unknown context fails before any hook", ctxName: "staging", preExecHooks: []string{"exit 3"}, wantErr: `cannot isolate context "staging"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := isolatedEnv(t, namespacedKubeconfig)
			hookOut := filepath.Join(t.TempDir(), "hook.out")
			t.Setenv("HOOK_OUT", hookOut)
			viper.Set(config.KeyPreExecHook, tt.preExecHooks)

			cfg, err := prepareIsolation(&bytes.Buffer{}, "cannot isolate")
			if err != nil {
				t.Fatal(err)
			}
			path, st, err := isolate(context.Background(), &bytes.Buffer{}, cfg, tt.ctxName, tt.namespace)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("isolate() error = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}

			if filepath.Dir(path) != tempDir {
				t.Errorf("temp kubeconfig %q is not in %q", path, tempDir)
			}
			want := state.State{Context: tt.ctxName, Namespace: tt.wantNamespace, Kubeconfig: path}
			if st != want {
				t.Errorf("state = %+v, want %+v", st, want)
			}
			written, loadErr := clientcmd.LoadFromFile(path)
			if loadErr != nil {
				t.Fatal(loadErr)
			}
			if written.CurrentContext != tt.ctxName || written.Contexts[tt.ctxName].Namespace != tt.wantNamespace {
				t.Errorf("temp kubeconfig has %q at namespace %q", written.CurrentContext, written.Contexts[tt.ctxName].Namespace)
			}
			if tt.wantHookOut != "" {
				got, err := os.ReadFile(hookOut)
				if err != nil {
					t.Fatal(err)
				}
				if string(got) != tt.wantHookOut {
					t.Errorf("hook saw %q, want %q", got, tt.wantHookOut)
				}
			}
		})
	}
}

func TestRunExec(t *testing.T) {
	tests := []struct {
		name    string
		argv    []string
		nested  bool
		wantErr string
	}{
		{name: "a command that succeeds", argv: []string{truePath(t)}},
		{name: "a command that doesn't exist", argv: []string{filepath.Join(t.TempDir(), "missing")}, wantErr: "exec failed"},
		{name: "inside a kush shell", argv: []string{truePath(t)}, nested: true, wantErr: "cannot run exec"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := isolatedEnv(t, namespacedKubeconfig)
			if tt.nested {
				t.Setenv(state.EnvActive, "1")
			}

			err := runExec(context.Background(), &bytes.Buffer{}, "prod", "", tt.argv)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("runExec() error = %v, want %q", err, tt.wantErr)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			assertNoTempKubeconfigs(t, tempDir)
		})
	}
}

func TestRunCtx(t *testing.T) {
	tests := []struct {
		name       string
		kubeconfig string
		ctxName    string
		picker     string
		wantErr    string
	}{
		{name: "a named context", kubeconfig: namespacedKubeconfig, ctxName: "prod"},
		{name: "an unknown context", kubeconfig: namespacedKubeconfig, ctxName: "staging", wantErr: `cannot isolate context "staging"`},
		{name: "the picker without contexts", kubeconfig: "apiVersion: v1\nkind: Config\n", wantErr: "no contexts found"},
		{name: "the picker misconfigured", kubeconfig: namespacedKubeconfig, picker: "dmenu", wantErr: "cannot determine the context picker"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := isolatedEnv(t, tt.kubeconfig)
			viper.Set(config.KeyShell, truePath(t))
			viper.Set(config.KeyPicker, tt.picker)

			err := runCtx(context.Background(), &bytes.Buffer{}, tt.ctxName, "")
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("runCtx() error = %v, want %q", err, tt.wantErr)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			assertNoTempKubeconfigs(t, tempDir)
		})
	}
}

// assertNoTempKubeconfigs fails when a temp kubeconfig outlived the command.
func assertNoTempKubeconfigs(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("temp kubeconfigs left behind in %s: %v", dir, entries)
	}
}
