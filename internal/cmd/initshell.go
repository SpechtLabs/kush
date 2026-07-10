package cmd

import (
	"fmt"

	humane "github.com/sierrasoftworks/humane-errors-go"
	"github.com/spf13/cobra"
)

// Opt-in prompt fallback for plain shells with no starship/oh-my-posh. Engine
// users can ignore this; starship's kubernetes module already renders from
// KUBECONFIG. The markers key on KUSH_ACTIVE/KUSH_CONTEXT.
const (
	initBash = `# kush prompt fallback (bash). Add to ~/.bashrc:  eval "$(kush init bash)"
if [ -n "$KUSH_ACTIVE" ]; then
  PS1="(kush:${KUSH_CONTEXT}) ${PS1}"
fi
`
	initZsh = `# kush prompt fallback (zsh). Add to ~/.zshrc:  eval "$(kush init zsh)"
if [ -n "$KUSH_ACTIVE" ]; then
  PROMPT="(kush:${KUSH_CONTEXT}) ${PROMPT}"
fi
`
	initFish = `# kush prompt fallback (fish). Add to config.fish:  kush init fish | source
if set -q KUSH_ACTIVE
  functions -q __kush_orig_fish_prompt; or functions -c fish_prompt __kush_orig_fish_prompt
  function fish_prompt
    printf '(kush:%s) ' $KUSH_CONTEXT
    __kush_orig_fish_prompt
  end
end
`
)

var cmdInit = &cobra.Command{
	Use:       "init <bash|zsh|fish>",
	Short:     "Emit opt-in prompt-fallback shell glue",
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"bash", "zsh", "fish"},
	RunE: func(cmd *cobra.Command, args []string) error {
		var snippet string
		switch args[0] {
		case "bash":
			snippet = initBash
		case "zsh":
			snippet = initZsh
		case "fish":
			snippet = initFish
		default:
			return humane.New(fmt.Sprintf("unsupported shell %q", args[0]), "use one of: bash, zsh, fish")
		}
		_, err := fmt.Fprint(cmd.OutOrStdout(), snippet)
		return err
	},
}
