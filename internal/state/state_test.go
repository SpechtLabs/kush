package state

import (
	"testing"
)

func TestActive(t *testing.T) {
	tests := []struct {
		name string
		val  string
		want bool
	}{
		{"set to one", "1", true},
		{"empty", "", false},
		{"other value", "true", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(EnvActive, tt.val)
			if got := Active(); got != tt.want {
				t.Fatalf("Active() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCurrent(t *testing.T) {
	t.Setenv(EnvActive, "1")
	t.Setenv(EnvContext, "prod")
	t.Setenv(EnvNamespace, "web")
	t.Setenv(EnvKubeconfig, "/run/kush/prod.yaml")

	got, ok := Current()
	if !ok {
		t.Fatal("Current() ok = false, want true")
	}
	want := State{Context: "prod", Namespace: "web", Kubeconfig: "/run/kush/prod.yaml"}
	if got != want {
		t.Fatalf("Current() = %+v, want %+v", got, want)
	}
}

func TestCurrentInactive(t *testing.T) {
	t.Setenv(EnvActive, "")
	if _, ok := Current(); ok {
		t.Fatal("Current() ok = true when inactive, want false")
	}
}

func TestGuardNesting(t *testing.T) {
	t.Run("blocks when active", func(t *testing.T) {
		t.Setenv(EnvActive, "1")
		t.Setenv(EnvContext, "prod")
		if err := GuardNesting(); err == nil {
			t.Fatal("GuardNesting() = nil, want error")
		}
	})
	t.Run("allows when inactive", func(t *testing.T) {
		t.Setenv(EnvActive, "")
		if err := GuardNesting(); err != nil {
			t.Fatalf("GuardNesting() = %v, want nil", err)
		}
	})
}

func TestEnv(t *testing.T) {
	s := State{Context: "prod", Namespace: "web", Kubeconfig: "/run/kush/prod.yaml"}
	got := s.Env()
	want := []string{
		"KUSH_ACTIVE=1",
		"KUSH_CONTEXT=prod",
		"KUSH_NAMESPACE=web",
		"KUSH_KUBECONFIG=/run/kush/prod.yaml",
	}
	if len(got) != len(want) {
		t.Fatalf("Env() len = %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Env()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
