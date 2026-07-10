package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	humane "github.com/sierrasoftworks/humane-errors-go"
	"github.com/spechtlabs/kush/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewRootCmd builds the kush root command. The bare form `kush [ctx]` enters a
// subshell for the named context (see ctx.go); with no arg it opens the picker.
func NewRootCmd() *cobra.Command {
	cobra.OnInitialize(initConfig)

	root := &cobra.Command{
		Use:   "kush [context]",
		Short: "Ephemeral, isolated kube-context subshells",
		Long: "kush drops you into a throwaway subshell pinned to exactly one " +
			"Kubernetes context. Prod in one terminal, dev in another, with no " +
			"leakage between them or back into ~/.kube/config.",
		Args:              cobra.MaximumNArgs(1),
		SilenceUsage:      true,
		SilenceErrors:     true,
		ValidArgsFunction: completeContexts,
		RunE: func(cmd *cobra.Command, args []string) error {
			if list, _ := cmd.Flags().GetBool("list"); list {
				return listContexts(cmd)
			}
			name := ""
			if len(args) == 1 {
				name = args[0]
			}
			return runCtx(cmd.Context(), cmd.ErrOrStderr(), name, "")
		},
	}

	root.Flags().BoolP("list", "l", false, "list all discovered contexts and exit (no subshell)")
	root.Version = Version
	root.AddCommand(newVersionCmd())
	return root
}

func initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		viper.AddConfigPath(filepath.Join(xdg, "kush"))
	}
	viper.AddConfigPath("$HOME/.config/kush")
	viper.AddConfigPath("/etc/kush")

	viper.SetEnvPrefix("KUSH")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	// Bind only the picker mode to an env override (KUSH_PICKER). We deliberately
	// do NOT use viper.AutomaticEnv(): it would bind KUSH_CONTEXT_LOOKUP_LOCATIONS
	// and whitespace-split it (not PathListSeparator), which is spec-inconsistent —
	// lookup locations are configured via the file only.
	_ = viper.BindEnv(config.KeyPicker)

	if err := viper.ReadInConfig(); err != nil {
		if _, notFound := err.(viper.ConfigFileNotFoundError); !notFound {
			fmt.Fprintln(os.Stderr, humane.Wrap(err, "failed to read kush config", "check the syntax of your ~/.config/kush/config.yaml").Display())
			os.Exit(2)
		}
	}
}

// AddSubcommands wires every non-root subcommand. Phase 1 registers nothing new
// beyond version; later phases extend this.
func AddSubcommands(root *cobra.Command) {
	root.AddCommand(cmdCtx)
	root.AddCommand(cmdCurrent)
	root.AddCommand(cmdNs)
	root.AddCommand(cmdInit)
	root.AddCommand(cmdExec)
	root.AddCommand(cmdLint)
	root.AddCommand(cmdSplit)
}
