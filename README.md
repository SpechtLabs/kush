# kush

[![Documentation](https://github.com/spechtlabs/kush/actions/workflows/docs-website.yaml/badge.svg)](https://github.com/spechtlabs/kush/actions/workflows/docs-website.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/spechtlabs/kush)](https://goreportcard.com/report/github.com/spechtlabs/kush)
[![Go Doc](https://godoc.org/github.com/spechtlabs/kush?status.svg)](https://godoc.org/github.com/spechtlabs/kush)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](./LICENSE)

> Ephemeral, isolated kube-context subshells. One shell, one context, deleted on exit.

Run `kush prod` and you drop into a normal shell pinned to the `prod` context through a private, throwaway kubeconfig. Exit that shell and the kubeconfig is deleted; nothing about the session persists and `~/.kube/config` is never touched. Open a `prod` shell in one terminal and a `dev` shell in another, and neither can bleed into the other or into your default kubectl setup.

**[Full documentation & getting started →](https://kush.specht-labs.de)**

## The problem it solves

`kubectl config use-context` mutates a single shared file. Switch context in one terminal and every other terminal, script, and background tool reading `~/.kube/config` switches with it. That's how you end up running a command against production because some other pane changed context five minutes ago.

kush sidesteps this by never touching that file. Each shell gets its own minimal kubeconfig scoped to exactly one context, so there's no global state to corrupt and no wrong-cluster surprise waiting for you.

It's also auth-agnostic: kush copies the user block from your existing kubeconfig verbatim, so exec plugins, OIDC, and cloud auth keep working unchanged. Ships as a single Go binary with no runtime dependencies.

## Install

kush is unix-only in v1 (no Windows, no `cmd.exe`/PowerShell). Two ways to get it:

```shell
# Build from source
go install github.com/spechtlabs/kush/cmd/kush@latest
```

Or grab a prebuilt binary from the [releases page](https://github.com/spechtlabs/kush/releases).

Verify:

```shell
$ kush version
```

If you have [`fzf`](https://github.com/junegunn/fzf) on your `PATH`, kush uses it for the context picker automatically; otherwise it falls back to a built-in TUI. Nothing to configure either way.

## Usage

The whole workflow is enter, work, exit:

```shell
$ kush prod                       # subshell pinned to prod, via a private kubeconfig
$ kubectl get pods                # kubectl, helm, k9s, flux: all see only prod
$ kush ns kube-system             # re-pin the namespace in place, same shell
$ exit                            # temp kubeconfig deleted, back to your normal shell
```

Run `kush` with no argument to pick a context interactively. To switch context, you exit and enter another; kush deliberately won't change context in place, since that's the silent state change that gets people into trouble.

For scripts, CI, and agents that need a single command without an interactive shell, use `kush exec`. It pins `KUBECONFIG` for just that one process, forwards stdin/stdout/stderr, propagates the exit code, and cleans up when the command finishes:

```shell
$ kush exec prod -- kubectl get pods
$ kush exec prod -n kube-system -- kubectl get pods
```

## Commands

| Command | What it does |
| --- | --- |
| `kush [ctx]` / `kush ctx [name]` | Enter an isolated subshell for a context. No argument opens the picker. `-l` lists contexts and exits. |
| `kush ns [name]` | Re-pin the namespace of the current kush shell in place (or spawn a shell for your current context, if run outside one). |
| `kush exec <ctx> -- <cmd>` | Run one command against an isolated context, non-interactively. `-n` sets the namespace. |
| `kush current` | Print the active `context/namespace`. Empty and exit `0` outside a kush shell. |
| `kush lint` | Check every discovered kubeconfig for missing clusters/users, empty current-context, unreachable files. Exits non-zero on errors. |
| `kush split [-o dir]` | Write one self-contained kubeconfig per context to disk. Never mutates the source. |
| `kush init <bash\|zsh\|fish>` | Emit optional prompt glue that shows a `(kush:<ctx>)` marker. |
| `kush completion <bash\|zsh\|fish>` | Standard shell completion for context names. |
| `kush version` | Print the version. |

Full flags and examples live in the [CLI reference](./docs/reference/cli.md).

## Configuration

Config is optional; absent means sane defaults. kush reads `config.yaml` from `~/.config/kush/`, `$XDG_CONFIG_HOME/kush/`, or `/etc/kush/`.

```yaml
# ~/.config/kush/config.yaml

# Where kush looks for kubeconfigs, in precedence order.
# Omit to use the default: $KUBECONFIG + ~/.kube/config.
context_lookup_locations:
  - $KUBECONFIG
  - ~/.kube/config
  - ~/.kube/configs/*        # globs and $ENV/~ expansion both work

picker: auto                 # auto | builtin | fzf   (env: KUSH_PICKER)
shell: ""                    # empty = $SHELL, then /bin/bash   (env: KUSH_SHELL)
```

Inside every subshell, kush sets `KUSH_ACTIVE`, `KUSH_CONTEXT`, `KUSH_NAMESPACE`, and `KUSH_KUBECONFIG`, plus `KUBECONFIG` itself pointed at the temp file. That last one is what makes every kube-aware tool honor the isolation without any kush-specific flag. See the [configuration reference](./docs/reference/configuration.md) for the details.

## How isolation works

When you enter a context, kush extracts just that context's cluster and user into a fresh temp kubeconfig, points `KUBECONFIG` at it, and forks your shell. On exit (normal, signal, or crash) the temp file is deleted, and a stale-file sweep reaps anything a `kill -9` might have left behind. Your real kubeconfig is only ever read, never written. The [isolation writeup](./docs/understanding/isolation.md) covers the mechanics.

## License

Apache 2.0. See [LICENSE](./LICENSE).
