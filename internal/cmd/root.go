package cmd

import (
	"strings"

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
		Args:          cobra.MaximumNArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := ""
			if len(args) == 1 {
				name = args[0]
			}
			return runCtx(cmd.Context(), name, "")
		},
	}

	root.Version = Version
	root.AddCommand(newVersionCmd())
	return root
}

func initConfig() {
	viper.SetEnvPrefix("KUSH")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	viper.AutomaticEnv()
}

// AddSubcommands wires every non-root subcommand. Phase 1 registers nothing new
// beyond version; later phases extend this.
func AddSubcommands(root *cobra.Command) {
	root.AddCommand(cmdCtx)
	root.AddCommand(cmdCurrent)
	root.AddCommand(cmdNs)
	root.AddCommand(cmdInit)
	root.AddCommand(cmdExec)
}
