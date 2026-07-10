Build `kush` — an ephemeral, isolated kube-context subshell tool (a generic,
auth-agnostic successor to `tka shell` and to the unmaintained `kubie`).
START HERE: read the full design spec before doing anything:
docs/superpowers/specs/2026-07-10-kush-design.md
It is approved and authoritative. Follow it; if you think something in it is
wrong, raise it with me before deviating.
Context you'll want:

- Module github.com/spechtlabs/kush, Go 1.26. Repo is empty except go.mod.
- Reference implementation for the subshell fork + signal handling:
  /Users/c.specht/src/gh/SpechtLabs/tka/cmd/cli/cmd_shell.go — kush's
  internal/shell is a near-direct lift MINUS the credential-revocation
  goroutine (kush has no creds to revoke). Match tka's cobra/viper structure.
- Stack: Go + spf13/cobra+viper + k8s.io/client-go (clientcmd) + charmbracelet.
- Follow the Go standards in ~/.claude/rules/go.md (wrap errors with %w, ctx
  first param, no bare panics in libs, table tests, etc.).
  Core semantics (see spec for full detail):
- One shell = one context. `kush <ctx>` extracts ONLY that context+cluster+user
  into a private temp kubeconfig ($XDG_RUNTIME_DIR/kush or $TMPDIR, 0600),
  forks $SHELL with KUBECONFIG pointed at it, deletes it on exit. Isolation is
  the whole point: prod in one terminal, dev in another, zero bleed.
- Nesting is BLOCKED via a KUSH_ACTIVE env guard (no depth tracking).
- Picker: fzf if on PATH (inherit os.Environ so the user's FZF_DEFAULT_OPTS /
  styling win — pass only --prompt/--no-multi/--ansi, zero styling flags),
  else a built-in charm/huh TUI.
- Prompt integration = export KUSH_ACTIVE/KUSH_CONTEXT/KUSH_NAMESPACE/
  KUSH_KUBECONFIG and let starship/oh-my-posh render them; `kush init <shell>`
  emits opt-in PS1/fish_prompt fallback for plain shells.
  Process:

1. Invoke the superpowers writing-plans skill to turn the spec into a PHASED
   implementation plan, in this order:
   Phase 1: core — `kush ctx` + extract-one isolation + subshell + cleanup +
   stale-file sweep + nesting guard. End-to-end working.
   Phase 2: `kush ns` (in-place inside a shell / subshell outside) + picker
   (fzf + fallback) + `kush init` prompt glue.
   Phase 3: `kush exec`.
   Phase 4: `kush lint` + `kush split`.
2. Then implement phase by phase (superpowers executing-plans / TDD).
   internal/kubeconfig is the pure, testable core — real table tests there
   (extract-one correctness, no cross-context bleed, ns edit, split round-trip).
3. Commit at the end of each phase with a clear message.
   Show me the plan for approval before you start writing code.
