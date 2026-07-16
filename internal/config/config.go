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
	// KeyPostExecHook is the config key holding the global post-exec hook.
	KeyPostExecHook = "post_exec_hook"
	// KeyContexts is the config key holding per-context settings.
	KeyContexts = "contexts"
	// KeyContextPreExecHook is the per-context key holding a pre-exec hook.
	KeyContextPreExecHook = "pre_exec_hook"
	// KeyContextPostExecHook is the per-context key holding a post-exec hook.
	KeyContextPostExecHook = "post_exec_hook"
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

// PreExecHooks returns the hooks configured for ctxName. Per-context hooks
// override the global hooks; an empty list means "run no hooks".
func PreExecHooks(ctxName string) []string {
	return hooks(ctxName, KeyPreExecHook, KeyContextPreExecHook)
}

// PostExecHooks returns the hooks configured for ctxName. Per-context hooks
// override the global hooks; an empty list means "run no hooks".
func PostExecHooks(ctxName string) []string {
	return hooks(ctxName, KeyPostExecHook, KeyContextPostExecHook)
}

func hooks(ctxName, globalKey, contextKey string) []string {
	if configured := contextHooks(ctxName, contextKey); len(configured) > 0 {
		return configured
	}
	return hookList(viper.Get(globalKey))
}

func contextHooks(ctxName, key string) []string {
	if ctxName == "" {
		return nil
	}
	contexts := viper.GetStringMap(KeyContexts)
	raw, ok := contexts[ctxName]
	if !ok {
		raw, ok = contexts[strings.ToLower(ctxName)]
	}
	if !ok {
		return nil
	}
	settings, ok := raw.(map[string]any)
	if !ok {
		return nil
	}
	return hookList(settings[key])
}

// hookList accepts the documented list form and the old scalar form so
// existing configurations continue to work while users migrate.
func hookList(raw any) []string {
	var values []string
	switch value := raw.(type) {
	case string:
		values = []string{value}
	case []string:
		values = value
	case []any:
		values = make([]string, 0, len(value))
		for _, item := range value {
			if hook, ok := item.(string); ok {
				values = append(values, hook)
			}
		}
	default:
		return nil
	}

	hooks := make([]string, 0, len(values))
	for _, value := range values {
		if hook := strings.TrimSpace(value); hook != "" {
			hooks = append(hooks, hook)
		}
	}
	return hooks
}
