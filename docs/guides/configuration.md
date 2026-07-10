---
title: Config & discovery
createTime: 2026/07/10 10:00:00
---

kush works with zero configuration — it reads `$KUBECONFIG` and `~/.kube/config`, just like kubectl. An optional `config.yaml` lets you change where kush looks for contexts, which picker it uses, and which shell it forks.

## Where the config file lives

kush reads `config.yaml` from:

- `~/.config/kush/`
- `$XDG_CONFIG_HOME/kush/` (if `XDG_CONFIG_HOME` is set)
- `/etc/kush/`

An absent config file is not an error — kush just falls back to its defaults. A malformed one (invalid YAML) produces a clear error at startup instead of silently ignoring it.

## `context_lookup_locations`

This is the list of places kush looks for kubeconfig files. Each entry supports `$ENV`/`${ENV}` expansion, `~` for your home directory, and globs.

```yaml
context_lookup_locations:
  - $KUBECONFIG
  - ~/.kube/config
  - ~/.kube/configs/*
```

A few things worth knowing:

- **Order is precedence.** Entries are checked top to bottom.
- **`$KUBECONFIG` splits on `:`.** If it holds multiple paths (`~/.kube/config:~/.kube/prod.yaml`), each one is tried in order, same as kubectl's own merge rule.
- **Setting this replaces the default, it doesn't extend it.** If `context_lookup_locations` is present and non-empty, kush uses exactly what you listed — nothing else. If it's absent, empty, or none of its entries resolve to anything, kush falls back to the default: `$KUBECONFIG` plus `~/.kube/config`.

## Merge order and duplicate contexts

Contexts are merged first-wins: whichever file earlier in `context_lookup_locations` defines a given context name wins. If the same context name shows up in more than one file, kush doesn't silently pick one — it warns on stderr:

```
warning: context "prod" defined in 2 files; using ~/.kube/config
```

## Bad files in a glob

If a glob entry like `~/.kube/configs/*` matches a file that isn't a valid kubeconfig — a stray `.bak`, a README, whatever — kush doesn't error out. It skips that file with a warning and keeps going:

```
warning: skipping ~/.kube/configs/notes.txt: not a valid kubeconfig
```

So a messy directory of configs just works; you don't have to keep it pristine.

## The `*` vs `*.yaml` gotcha

Kubeconfig files frequently have no file extension at all — that's how most cloud CLIs export them. If you glob for `~/.kube/configs/*.yaml` and your files are actually named `prod-cluster`, `staging-cluster`, and so on, the glob matches nothing and kush silently finds zero extra contexts.

Use `~/.kube/configs/*` instead, unless you know your files genuinely end in `.yaml`. The "skip invalid files" behavior above means a bare `*` is safe even if the directory has other, non-kubeconfig files mixed in.

## `picker`

Controls which context picker `kush ctx` opens with no argument. Env override: `KUSH_PICKER`.

- `auto` (default) — fzf if it's on `PATH`, otherwise the built-in charm/huh TUI.
- `builtin` — always the TUI, even if fzf is installed.
- `fzf` — always fzf; errors clearly if fzf isn't installed.

fzf inherits your full environment when kush shells out to it, so `FZF_DEFAULT_OPTS`, your colors, and your keybindings all apply as normal. kush itself only ever passes `--prompt`, `--no-multi`, and `--print-query`.

## `shell`

The shell kush forks when you enter a context. Default is `$SHELL`, falling back to `/bin/bash` if that's unset. Env override: `KUSH_SHELL`.

Set this explicitly if your interactive shell differs from your login `$SHELL` — the classic case is running fish day to day while `$SHELL` is still set to `/bin/zsh` (common if you switched shells without updating it). Left unconfigured, kush would fork zsh instead of fish, and your subshell's command history and atuin integration would land somewhere you're not looking.

## Complete example

```yaml
context_lookup_locations:
  - $KUBECONFIG
  - ~/.kube/config
  - ~/.kube/configs/*

picker: fzf

shell: /opt/homebrew/bin/fish
```
