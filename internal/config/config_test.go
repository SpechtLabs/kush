package config

import (
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestParsePicker(t *testing.T) {
	tests := []struct {
		in      string
		want    PickerMode
		wantErr bool
	}{
		{"", PickerAuto, false},
		{"auto", PickerAuto, false},
		{"AUTO", PickerAuto, false},
		{"builtin", PickerBuiltin, false},
		{"fzf", PickerFzf, false},
		{" fzf ", PickerFzf, false},
		{"nonsense", PickerAuto, true},
	}
	for _, tt := range tests {
		got, err := ParsePicker(tt.in)
		if (err != nil) != tt.wantErr {
			t.Fatalf("ParsePicker(%q) err = %v, wantErr %v", tt.in, err, tt.wantErr)
		}
		if got != tt.want {
			t.Fatalf("ParsePicker(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestPickerDefault(t *testing.T) {
	viper.Reset()
	got, err := Picker()
	if err != nil || got != PickerAuto {
		t.Fatalf("Picker() = %v, %v; want PickerAuto, nil", got, err)
	}
}

func TestPickerFromViper(t *testing.T) {
	viper.Reset()
	viper.Set(KeyPicker, "fzf")
	got, _ := Picker()
	if got != PickerFzf {
		t.Fatalf("Picker() = %v, want PickerFzf", got)
	}
}

func TestShell(t *testing.T) {
	viper.Reset()
	if got := Shell(); got != "" {
		t.Fatalf("Shell() = %q, want empty when unset", got)
	}
	viper.Set(KeyShell, "/opt/homebrew/bin/fish")
	if got := Shell(); got != "/opt/homebrew/bin/fish" {
		t.Fatalf("Shell() = %q, want /opt/homebrew/bin/fish", got)
	}
}

func TestLookupLocations(t *testing.T) {
	viper.Reset()
	if got := LookupLocations(); got != nil {
		t.Fatalf("LookupLocations() = %v, want nil when unset", got)
	}
	viper.Set(KeyLookupLocations, []string{"$KUBECONFIG", "~/.kube/config"})
	got := LookupLocations()
	if len(got) != 2 || got[0] != "$KUBECONFIG" {
		t.Fatalf("LookupLocations() = %v", got)
	}
}

func TestPreExecHook(t *testing.T) {
	viper.Reset()
	if got := PreExecHook("prod"); got != "" {
		t.Fatalf("PreExecHook() = %q, want empty when unset", got)
	}

	viper.Set(KeyPreExecHook, " tsh join $KUSH_CONTEXT ")
	if got := PreExecHook("prod"); got != "tsh join $KUSH_CONTEXT" {
		t.Fatalf("PreExecHook() = %q, want global hook", got)
	}

	viper.Set(KeyContexts, map[string]any{
		"prod": map[string]any{
			KeyContextPreExecHook: "tsh join prod",
		},
	})
	if got := PreExecHook("prod"); got != "tsh join prod" {
		t.Fatalf("PreExecHook() = %q, want per-context hook", got)
	}
	if got := PreExecHook("dev"); got != "tsh join $KUSH_CONTEXT" {
		t.Fatalf("PreExecHook() = %q, want global fallback", got)
	}
}

func TestPreExecHookFromYAML(t *testing.T) {
	viper.Reset()
	viper.SetConfigType("yaml")
	err := viper.ReadConfig(strings.NewReader(`
pre_exec_hook: "tsh join $KUSH_CONTEXT"
contexts:
  cluster-123:
    pre_exec_hook: "tsh join cluster-123"
`))
	if err != nil {
		t.Fatal(err)
	}

	if got := PreExecHook("cluster-123"); got != "tsh join cluster-123" {
		t.Fatalf("PreExecHook() = %q, want per-context hook from YAML", got)
	}
	if got := PreExecHook("cluster-456"); got != "tsh join $KUSH_CONTEXT" {
		t.Fatalf("PreExecHook() = %q, want global fallback from YAML", got)
	}
}
