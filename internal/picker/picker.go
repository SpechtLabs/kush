// Package picker selects a context (from a list) or a free-text value. It uses
// the user's fzf when present, else a built-in charm/huh TUI. fzf inherits the
// full environment so the user's FZF_DEFAULT_OPTS/styling win; we pass only
// functional flags (--prompt, --no-multi).
package picker

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/huh"
	humane "github.com/sierrasoftworks/humane-errors-go"
)

func fzfPath() (string, bool) {
	p, err := exec.LookPath("fzf")
	return p, err == nil
}

// Select returns the item the user chooses from items.
func Select(ctx context.Context, prompt string, items []string) (string, error) {
	if len(items) == 0 {
		return "", humane.New("nothing to choose from", "the list passed to the picker was empty")
	}
	if path, ok := fzfPath(); ok {
		return fzfSelect(ctx, path, prompt, items)
	}
	return huhSelect(prompt, items)
}

// Prompt returns a free-text value the user types.
func Prompt(ctx context.Context, prompt string) (string, error) {
	if path, ok := fzfPath(); ok {
		return fzfPrompt(ctx, path, prompt)
	}
	return huhInput(prompt)
}

func fzfSelect(ctx context.Context, path, prompt string, items []string) (string, error) {
	cmd := exec.CommandContext(ctx, path, "--prompt="+prompt, "--no-multi")
	cmd.Env = os.Environ()
	cmd.Stdin = strings.NewReader(strings.Join(items, "\n"))
	cmd.Stderr = os.Stderr
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", humane.Wrap(err, "fzf selection cancelled", "pick an entry or set a context explicitly: kush <context>")
	}
	sel := strings.TrimSpace(out.String())
	if sel == "" {
		return "", humane.New("no selection made", "pick an entry or set a context explicitly: kush <context>")
	}
	return sel, nil
}

func fzfPrompt(ctx context.Context, path, prompt string) (string, error) {
	// Empty input + --print-query returns whatever the user typed on the query line.
	cmd := exec.CommandContext(ctx, path, "--prompt="+prompt, "--print-query", "--no-multi")
	cmd.Env = os.Environ()
	cmd.Stdin = strings.NewReader("")
	cmd.Stderr = os.Stderr
	var out bytes.Buffer
	cmd.Stdout = &out
	// fzf exits non-zero (1) when the list has no match; the query line is still printed.
	_ = cmd.Run()
	val := strings.TrimSpace(strings.SplitN(out.String(), "\n", 2)[0])
	if val == "" {
		return "", humane.New("no value entered", "type a value, or pass it explicitly on the command line")
	}
	return val, nil
}

func huhSelect(prompt string, items []string) (string, error) {
	var selected string
	opts := make([]huh.Option[string], 0, len(items))
	for _, it := range items {
		opts = append(opts, huh.NewOption(it, it))
	}
	err := huh.NewSelect[string]().Title(prompt).Options(opts...).Value(&selected).Run()
	if err != nil {
		return "", humane.Wrap(err, "selection cancelled", "pick an entry or set the value explicitly on the command line")
	}
	return selected, nil
}

func huhInput(prompt string) (string, error) {
	var val string
	err := huh.NewInput().Title(prompt).Value(&val).Run()
	if err != nil {
		return "", humane.Wrap(err, "input cancelled", "type a value, or pass it explicitly on the command line")
	}
	val = strings.TrimSpace(val)
	if val == "" {
		return "", humane.New("no value entered", "type a value, or pass it explicitly on the command line")
	}
	return val, nil
}
