package main

import (
	"fmt"
	"io"

	"github.com/spechtlabs/kush/internal/config"
	"github.com/spechtlabs/kush/internal/kubeconfig"
	"github.com/spechtlabs/kush/internal/picker"
	"k8s.io/client-go/tools/clientcmd/api"
)

// resolveLoad loads the merged kubeconfig from the configured lookup locations
// (or the kubeconfig defaults when unset), printing any duplicate-context
// warnings to warnOut. Pass cmd.ErrOrStderr() as warnOut.
func resolveLoad(warnOut io.Writer) (*api.Config, error) {
	cfg, warnings, err := kubeconfig.LoadResolved(config.LookupLocations())
	if err != nil {
		return nil, err
	}
	for _, w := range warnings {
		fmt.Fprintln(warnOut, "warning:", w.Message)
	}
	return cfg, nil
}

// pickerMode translates the configured picker mode into a picker.Mode.
func pickerMode() (picker.Mode, error) {
	m, err := config.Picker()
	if err != nil {
		return picker.Auto, err
	}
	switch m {
	case config.PickerBuiltin:
		return picker.Builtin, nil
	case config.PickerFzf:
		return picker.Fzf, nil
	default:
		return picker.Auto, nil
	}
}
