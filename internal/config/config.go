// Package config exposes kush's user configuration (viper-backed).
// cmd/kush initializes viper from ~/.config/kush/config.yaml and the KUSH_ env
// prefix; this package is the typed, defaulted read side.
package config

import (
	"fmt"
	"strings"

	humane "github.com/sierrasoftworks/humane-errors-go"
	"github.com/spf13/viper"
)

// PickerMode selects which picker kush uses when no context arg is given.
type PickerMode int

const (
	// PickerAuto uses fzf if present, otherwise the built-in TUI.
	PickerAuto PickerMode = iota
	// PickerBuiltin always uses the built-in TUI.
	PickerBuiltin
	// PickerFzf always uses fzf and errors if it is not installed.
	PickerFzf
)

const (
	// KeyLookupLocations is the config key holding the context lookup locations.
	KeyLookupLocations = "context_lookup_locations"
	// KeyPicker is the config key holding the picker mode.
	KeyPicker = "picker"
	// KeyShell is the config key holding the subshell override.
	KeyShell = "shell"
	// KeyPreExecHook is the config key holding the global pre-exec hook.
	KeyPreExecHook = "pre_exec_hook"
	// KeyContexts is the config key holding per-context settings.
	KeyContexts = "contexts"
	// KeyContextPreExecHook is the per-context key holding a pre-exec hook.
	KeyContextPreExecHook = "pre_exec_hook"
)

// Shell returns the shell kush should fork for a subshell, or "" to fall back
// to $SHELL (then /bin/bash). Set it when your interactive shell differs from
// your login $SHELL (e.g. you run fish but $SHELL is zsh) so subshell history
// and atuin land where you expect.
func Shell() string {
	return viper.GetString(KeyShell)
}

// ParsePicker maps a config string to a PickerMode. "" and "auto" → PickerAuto;
// an unrecognized value is an error.
func ParsePicker(s string) (PickerMode, humane.Error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "auto":
		return PickerAuto, nil
	case "builtin":
		return PickerBuiltin, nil
	case "fzf":
		return PickerFzf, nil
	default:
		return PickerAuto, humane.New(fmt.Sprintf("invalid picker mode %q", s), "set picker to one of: auto, builtin, fzf")
	}
}

// Picker returns the configured picker mode (default PickerAuto).
func Picker() (PickerMode, humane.Error) {
	return ParsePicker(viper.GetString(KeyPicker))
}

// LookupLocations returns the configured context lookup locations, or nil when
// unset/empty (nil signals "use kubeconfig defaults").
func LookupLocations() []string {
	locs := viper.GetStringSlice(KeyLookupLocations)
	if len(locs) == 0 {
		return nil
	}
	return locs
}

// PreExecHook returns the hook configured for ctxName. A per-context hook
// overrides the global hook; an empty value means "run no hook".
func PreExecHook(ctxName string) string {
	if hook := contextPreExecHook(ctxName); hook != "" {
		return hook
	}
	return strings.TrimSpace(viper.GetString(KeyPreExecHook))
}

func contextPreExecHook(ctxName string) string {
	if ctxName == "" {
		return ""
	}
	contexts := viper.GetStringMap(KeyContexts)
	raw, ok := contexts[ctxName]
	if !ok {
		raw, ok = contexts[strings.ToLower(ctxName)]
	}
	if !ok {
		return ""
	}
	settings, ok := raw.(map[string]any)
	if !ok {
		return ""
	}
	hook, ok := settings[KeyContextPreExecHook].(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(hook)
}
