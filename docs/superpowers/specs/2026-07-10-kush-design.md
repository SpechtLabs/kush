# kush — Design Spec

**Date:** 2026-07-10
**Module:** `github.com/spechtlabs/kush`
**Status:** Approved design, pre-implementation

## Summary

`kush` (kube-shell) is a single Go binary that drops you into an **ephemeral,
isolated subshell pinned to one Kubernetes context**. Unlike `kubectl config
use-context` — which mutates global state shared by every shell and tool — a
kush subshell owns a private, throwaway kubeconfig. You can run **prod in one
terminal and dev in another simultaneously**, with no leakage between them or
back into `~/.kube/config`.

It is a generic, auth-agnostic reimagining of [`tka
shell`](https://tka.specht-labs.de/guides/use-subshell). kush assumes contexts
already exist and are already authenticated in your `KUBECONFIG` (exec/OIDC
plugins keep working — kush copies the `user` block verbatim). It carries none
of tka's Tailscale/token machinery. It is also a spiritual successor to the
now-unmaintained [`kubie`](https://github.com/sbstp/kubie).

## Mental model

- **One shell = one context.** Enter a context, get a shell locked to it. To
  change context, `exit` and enter another.
- **Namespace is mutable within that shell** — `kush ns foo` re-pins the
  namespace in place, no new shell.
- **Exit = gone.** On exit the temp kubeconfig is deleted. Nothing persisted.

## Design decisions

| Decision | Choice | Consequence |
|---|---|---|
| Isolation | **Extract one** context into the temp kubeconfig | Self-contained, minimal, crisp "locked" shell. No merge logic. |
| Picker | **fzf if present, else built-in charm TUI** | Best UX; respects user's fzf config (see §5). |
| Nesting | **Blocked** | No depth tracking. One env-var guard. |
| Stack | **Go + cobra + viper + client-go + charm** | Mirrors tka; reuse known patterns. |
| Scope | context **and** namespace subshells; kubie parity | ctx / ns / exec / lint / split / init. |
| Platform | **unix only** (v1) | `//go:build unix`, like tka. Windows is v2+. |

## Command surface

```
kush [ctx]                 # enter subshell for <ctx>; no arg → picker
kush ctx [name]            # explicit form of the above
kush ns [name]             # inside a kush shell: re-pin namespace in place (no arg → picker)
                           # outside: enter a subshell for the CURRENT context at that namespace
kush exec <ctx> [-n ns] -- <cmd...>   # run one command in an isolated context, no interactive shell
kush current               # print active context/namespace (empty if not in a kush shell)
kush lint                  # check KUBECONFIG(s) for common problems
kush split [-o dir]        # split a monolithic kubeconfig into one file per context
kush init <shell>          # emit shell glue (bash|zsh|fish) for optional prompt fallback
kush --version / completion / help
```

Context/namespace resolution follows kubectl exactly: `KUBECONFIG`
(colon-joined list, merged) with `~/.kube/config` fallback, via
`clientcmd.NewDefaultClientConfigLoadingRules()`.

## Detailed design

### 1. Isolation mechanism (the core)

On `kush prod`:

1. Load + merge all kubeconfigs per the loading rules → in-memory `api.Config`.
2. Look up context `prod`; resolve its referenced `cluster` and `user`.
3. Build a **new minimal `api.Config`** containing only that one context + its
   cluster + user, with `current-context: prod` and the chosen namespace set on
   the context.
4. Write it to a private temp file:
   `$XDG_RUNTIME_DIR/kush/<ctx>-<pid>-<rand>.yaml` (Linux tmpfs) → fallback
   `$TMPDIR` / `os.TempDir()` (macOS), mode `0600`.
5. Fork `$SHELL` (fallback `/bin/bash`) with `KUBECONFIG=<tempfile>` and the
   state env vars (§4), inheriting stdio and forwarding signals (§6).
6. On exit / signal / crash-sweep → delete the temp file.

Because `KUBECONFIG` points at the private file, **every** kube tool in that
shell (kubectl, helm, k9s, kustomize, …) sees only `prod`, and any in-shell
mutation hits the throwaway copy. `kush ns foo` = `clientcmd` edit of that
file's `context.namespace` + refresh of `KUSH_NAMESPACE`.

**`kush ns` dual behavior:**
- **Inside** a kush shell: edit the current shell's temp kubeconfig namespace in
  place. Immediate, no subshell (so it does not count as nesting).
- **Outside**: spawn a subshell for the current context pinned to that
  namespace (consistent with `kush <ctx>`). Error if there is no current
  context.

### 2. `kush exec`

Same temp-kubeconfig construction as the interactive path, but instead of
forking `$SHELL` it runs the given command with `KUBECONFIG` set, forwards
stdio and exit code, then removes the temp file. Works with piped/non-interactive
stdin. Intended for scripts and CI.

### 3. lint / split

- **lint**: load KUBECONFIG(s) and report common problems — contexts referencing
  a missing cluster or user, duplicate names across merged files, unreachable
  file entries, empty current-context. Read-only.
- **split**: write one file per context (self-contained, extract-one semantics)
  into `-o <dir>` (default `~/.kube/kush/` or cwd). Never mutates the source.

### 4. Prompt & state integration

kush exports state; it does not fight the user's prompt:

```
KUSH_ACTIVE=1
KUSH_CONTEXT=prod
KUSH_NAMESPACE=default
KUSH_KUBECONFIG=/run/.../kush/prod-1234-ab12.yaml
```

- **starship**: native `kubernetes` module already renders context+namespace
  from `KUBECONFIG` — works out of the box. Document an optional `custom` module
  keyed on `KUSH_ACTIVE` for an explicit "ephemeral" marker.
- **oh-my-posh**: document a segment reading the same env vars.
- **Plain bash/zsh** (no engine): `kush init bash|zsh` emits an opt-in `PS1`/`PROMPT`
  wrapper prepending a `(kush:prod)` marker.
- **Plain fish** (no engine): `kush init fish | source` — fish ignores `PS1`, so
  this wraps `fish_prompt`.

`kush init` follows the `starship init` model: engine users can ignore it.

### 5. Picker (fzf + fallback), respecting user fzf config

When no context arg is given:

- If `fzf` is on `PATH`: pipe the context list to `fzf` on stdin, read the
  selected line back.
  - **Respect the user's fzf configuration by not getting in the way.** Inherit
    the full environment (`os.Environ()`) so `FZF_DEFAULT_OPTS`,
    `FZF_DEFAULT_OPTS_FILE`, colors, `--height`, layout, keybindings, preview,
    etc. all apply automatically.
  - Pass **only non-styling functional flags**: at most `--prompt='kush ctx> '`,
    `--no-multi`, and `--ansi` (only if we colorize the list). **No**
    color/height/border/layout/keybind flags — fzf lets command-line args
    override `FZF_DEFAULT_OPTS`, so every flag we pass is one we steal from the
    user. Zero styling flags = the user sees *their* fzf.
- Else: built-in filterable list via charmbracelet (`huh`/`bubbletea`), matching
  tka's aesthetic. No external dependency.

Same picker serves `kush ns` (namespace list for the active/current context).

### 6. Nesting guard

At startup, if `KUSH_ACTIVE=1`, refuse:
`already in a kush shell (prod); exit first to switch context`. Exit code 1.
One env-var check — no depth state. `kush ns` (in-place edit) is exempt.

### 7. Lifecycle & cleanup

- Primary: `defer os.Remove(tempfile)` around the subshell run.
- Signals: adapt tka's `internal/cli` signal handler — terminate child, then
  remove temp file. Drop tka's async sign-out goroutine (no creds to revoke).
- **Stale sweep:** temp filenames embed the owning PID; on any `kush`
  invocation, opportunistically delete files in the kush temp dir whose PID is no
  longer alive. Covers `kill -9`/crashes. Bounded: a dir scan at startup, not a
  daemon.

## Package layout

```
cmd/kush/main.go            # cobra root wiring (mirror tka's cmd/cli structure)
internal/cmd/*.go           # one file per subcommand (ctx, ns, exec, current, lint, split, init)
internal/kubeconfig/        # load/merge, extract-one, write temp, edit namespace, lint, split  ← testable core
internal/shell/             # fork subshell, env assembly, signal handling (//go:build unix, from tka)
internal/picker/            # fzf-detect + shell-out; charm/huh fallback TUI
internal/state/             # KUSH_* env read/write, nesting guard
```

`internal/kubeconfig` is pure data-in/data-out (no shell, no TTY) — trivially
unit-testable, and the one place a bug corrupts user data.

## Edge cases

- Context not found → error listing available contexts.
- No current-context (bare `kush ns` outside a shell) → actionable error.
- exec-plugin users (OIDC/cloud auth) → preserved (user block copied verbatim).
- `$SHELL` unset → `/bin/bash`.
- Piped/non-interactive stdin → `exec` works; interactive subshell errors clearly.
- Windows → out of scope v1 (`//go:build unix`).

## Out of scope (v2+)

- Windows support.
- Auth/login integration (tka's job; kush could later complement it).
- Context/namespace history & `-` (previous) toggle.
- Any Kubernetes API calls — kush only manipulates kubeconfig files, never talks
  to a cluster.

## Testing

- `internal/kubeconfig`: table tests — extract-one correctness (right
  cluster/user resolved, no cross-context bleed), namespace edit, lint detection,
  split round-trip. Real coverage; this is where a bug corrupts user data.
- `internal/state`: nesting-guard logic.
- `internal/shell` / `internal/picker`: thin, manually verified, plus one smoke
  test that `kush exec` runs a command with the correct `KUBECONFIG`.

## Deliberate simplifications

- Context history / `-` toggle — add when missed.
- Copy-all-contexts isolation — rejected in favor of extract-one.
- Depth tracking — unnecessary; nesting is blocked.
- Windows — unix build tag for v1.
