package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spechtlabs/kush/internal/config"
	"github.com/spechtlabs/kush/internal/state"
	"github.com/spf13/viper"
	"k8s.io/client-go/tools/clientcmd"
)

// runKush runs the kush command tree with args, as main wires it, and returns
// what it printed to stdout.
func runKush(t *testing.T, args ...string) (string, error) {
	t.Helper()
	root := NewRootCmd()
	AddSubcommands(root)
	var out, errOut bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&errOut)
	root.SetArgs(args)
	err := root.ExecuteContext(context.Background())
	return out.String(), err
}

func TestCommands(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantOut    string
		exactOut   bool
		wantErr    string
		wantSplits []string
	}{
		{name: "list from the root", args: []string{"--list"}, wantOut: "dev (current)\nprod\n", exactOut: true},
		{name: "list from ctx", args: []string{"ctx", "-l"}, wantOut: "dev (current)\nprod\n", exactOut: true},
		{name: "current outside a kush shell prints nothing", args: []string{"current"}, exactOut: true},
		{name: "init bash", args: []string{"init", "bash"}, wantOut: initBash, exactOut: true},
		{name: "init zsh", args: []string{"init", "zsh"}, wantOut: initZsh, exactOut: true},
		{name: "init fish", args: []string{"init", "fish"}, wantOut: initFish, exactOut: true},
		{name: "init with an unsupported shell", args: []string{"init", "tcsh"}, wantErr: `unsupported shell "tcsh"`},
		{name: "lint a clean kubeconfig", args: []string{"lint"}, wantOut: "ok: no problems found"},
		{name: "split into a directory", args: []string{"split", "-o", "SPLIT_DIR"}, wantSplits: []string{"dev", "prod"}},
		{name: "a named context", args: []string{"ctx", "prod"}},
		{name: "exec in a namespace", args: []string{"exec", "prod", "-n", "audit", "--", "TRUE"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isolatedEnv(t, namespacedKubeconfig)
			viper.Set(config.KeyShell, truePath(t))
			splitDir := t.TempDir()
			args := make([]string, len(tt.args))
			for i, arg := range tt.args {
				arg = strings.ReplaceAll(arg, "SPLIT_DIR", splitDir)
				args[i] = strings.ReplaceAll(arg, "TRUE", truePath(t))
			}

			out, err := runKush(t, args...)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("kush %v error = %v, want %q", tt.args, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("kush %v: %v", tt.args, err)
			}
			if (tt.exactOut && out != tt.wantOut) || !strings.Contains(out, tt.wantOut) {
				t.Errorf("kush %v printed %q, want %q", tt.args, out, tt.wantOut)
			}
			for _, ctxName := range tt.wantSplits {
				path := filepath.Join(splitDir, ctxName+".yaml")
				if !strings.Contains(out, path) {
					t.Errorf("kush split didn't print %s:\n%s", path, out)
				}
				if _, err := clientcmd.LoadFromFile(path); err != nil {
					t.Errorf("kush split wrote no kubeconfig for %s: %v", ctxName, err)
				}
			}
		})
	}
}

func TestNs(t *testing.T) {
	t.Run("inside a kush shell it re-pins the namespace in place", func(t *testing.T) {
		isolatedEnv(t, namespacedKubeconfig)
		temp := filepath.Join(t.TempDir(), "prod.yaml")
		if err := os.WriteFile(temp, []byte(namespacedKubeconfig), 0o600); err != nil {
			t.Fatal(err)
		}
		t.Setenv(state.EnvActive, "1")
		t.Setenv(state.EnvContext, "dev")
		t.Setenv(state.EnvKubeconfig, temp)

		out, err := runKush(t, "ns", "audit")
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out, `namespace re-pinned to "audit"`) {
			t.Errorf("kush ns printed %q", out)
		}
		cfg, err := clientcmd.LoadFromFile(temp)
		if err != nil {
			t.Fatal(err)
		}
		if got := cfg.Contexts["dev"].Namespace; got != "audit" {
			t.Errorf("dev's namespace = %q, want audit", got)
		}
	})

	t.Run("inside a kush shell with a missing kubeconfig", func(t *testing.T) {
		isolatedEnv(t, namespacedKubeconfig)
		t.Setenv(state.EnvActive, "1")
		t.Setenv(state.EnvContext, "dev")
		t.Setenv(state.EnvKubeconfig, filepath.Join(t.TempDir(), "gone.yaml"))

		if _, err := runKush(t, "ns", "audit"); err == nil || !strings.Contains(err.Error(), "cannot re-pin the namespace") {
			t.Fatalf("kush ns error = %v, want the re-pin failure", err)
		}
	})

	t.Run("outside a kush shell it enters the current context", func(t *testing.T) {
		tempDir := isolatedEnv(t, namespacedKubeconfig)
		viper.Set(config.KeyShell, truePath(t))

		if _, err := runKush(t, "ns", "audit"); err != nil {
			t.Fatal(err)
		}
		assertNoTempKubeconfigs(t, tempDir)
	})

	t.Run("outside a kush shell without a current context", func(t *testing.T) {
		isolatedEnv(t, strings.Replace(namespacedKubeconfig, "current-context: dev\n", "", 1))

		if _, err := runKush(t, "ns", "audit"); err == nil || !strings.Contains(err.Error(), "no current context set") {
			t.Fatalf("kush ns error = %v, want the missing current context", err)
		}
	})

	t.Run("a misconfigured picker when prompting", func(t *testing.T) {
		isolatedEnv(t, namespacedKubeconfig)
		viper.Set(config.KeyPicker, "dmenu")

		if _, err := runKush(t, "ns"); err == nil || !strings.Contains(err.Error(), "cannot determine the namespace picker") {
			t.Fatalf("kush ns error = %v, want the picker failure", err)
		}
	})
}

func TestLintReportsProblems(t *testing.T) {
	broken := strings.Replace(namespacedKubeconfig, "    cluster: c1\n    user: u1\n    namespace: payments", "    cluster: missing\n    user: u1\n    namespace: payments", 1)
	isolatedEnv(t, broken)

	out, err := runKush(t, "lint")
	if err == nil || !strings.Contains(err.Error(), "kubeconfig lint found errors") {
		t.Fatalf("kush lint error = %v, want the lint failure", err)
	}
	if !strings.Contains(out, "error: ") || !strings.Contains(out, "missing") {
		t.Errorf("kush lint printed %q, want the dangling cluster reported", out)
	}
}

func TestCurrentInsideAKushShell(t *testing.T) {
	isolatedEnv(t, namespacedKubeconfig)
	t.Setenv(state.EnvActive, "1")
	t.Setenv(state.EnvContext, "prod")
	t.Setenv(state.EnvNamespace, "payments")

	out, err := runKush(t, "current")
	if err != nil {
		t.Fatal(err)
	}
	if out != "prod (payments)\n" {
		t.Errorf("kush current printed %q, want %q", out, "prod (payments)\n")
	}
}

func TestSplitDefaultsToKubeKush(t *testing.T) {
	isolatedEnv(t, namespacedKubeconfig)
	home := os.Getenv("HOME")

	if _, err := runKush(t, "split"); err != nil {
		t.Fatal(err)
	}
	for _, ctxName := range []string{"dev", "prod"} {
		if _, err := clientcmd.LoadFromFile(filepath.Join(home, ".kube", "kush", ctxName+".yaml")); err != nil {
			t.Errorf("kush split wrote no kubeconfig for %s under ~/.kube/kush: %v", ctxName, err)
		}
	}
}
