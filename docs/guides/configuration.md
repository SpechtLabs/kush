---
title: Config & discovery
createTime: 2026/07/10 10:00:00
---

kush has to find your kubeconfigs before it can pin one. Out of the box it reads `$KUBECONFIG` and `~/.kube/config`, the same as kubectl, so if that's where your clusters already live there's nothing to configure. This guide is for when they don't: a directory full of per-cluster files, several explicit paths, or a picker or shell you want to change.

All of it lives in an optional `config.yaml`, read from `~/.config/kush/`, `$XDG_CONFIG_HOME/kush/`, or `/etc/kush/`. No file means defaults. A malformed file is a startup error rather than a silent fallback, so you find out immediately. For the exhaustive key list see the [configuration reference](../reference/configuration.md); below are the setups you'll actually run into.

## "kush can't see all my clusters"

Almost always this is a kubeconfig layout kush doesn't look in by default. Point it at the extra locations with `context_lookup_locations`:

```yaml
# ~/.config/kush/config.yaml
context_lookup_locations:
  - $KUBECONFIG
  - ~/.kube/config
  - ~/.kube/configs/*
```

Two things will bite you if you don't know them:

- **The list replaces the default, it doesn't extend it.** The moment you set `context_lookup_locations`, kush looks at exactly what you wrote and nothing else. Keep `$KUBECONFIG` and `~/.kube/config` in the list if you still want them (or leave the key out entirely to keep the defaults).
- **Order is precedence.** Entries are read top to bottom, and when two files define the same context name the earlier one wins. kush says so on stderr rather than picking silently:

```text
warning: context "prod" defined in 2 files; using ~/.kube/config
```

## A directory of per-cluster files

It's common to keep one kubeconfig per cluster in a directory. Glob the directory and kush picks up every valid file in it:

```yaml
# ~/.config/kush/config.yaml
context_lookup_locations:
  - ~/.kube/config
  - ~/.kube/configs/*
```

Mind the extension. Exported kubeconfigs often have no `.yaml` suffix at all (they're named `prod-cluster`, `staging-eks`, and so on), so a glob like `~/.kube/configs/*.yaml` quietly matches nothing and kush finds zero extra contexts. Use a bare `*` unless you know every file ends in `.yaml`.

A bare `*` is safe even in a messy directory: anything that isn't a valid kubeconfig, a stray `.bak` or a README, is skipped with a warning instead of failing the run.

```text
warning: skipping ~/.kube/configs/notes.txt: not a valid kubeconfig
```

If you'd rather generate this layout from a single merged kubeconfig, `kush split` writes one self-contained file per context into a directory for you (see the [CLI reference](../reference/cli.md)).

## Choosing the picker

When you run `kush` or `kush ctx` with no context name, it opens a picker. `picker` decides which one, and `KUSH_PICKER` overrides it per invocation:

- `auto` (default): fzf if it's on your `PATH`, otherwise a built-in TUI.
- `builtin`: always the built-in TUI, even with fzf installed.
- `fzf`: always fzf, with a clear error if it isn't installed.

fzf runs with your full environment, so `FZF_DEFAULT_OPTS`, colors, and keybindings all carry over. kush only adds `--prompt`, `--no-multi`, and `--print-query`.

## Choosing the shell

kush forks `$SHELL` when you enter a context, falling back to `/bin/bash`. Set `shell` (or `KUSH_SHELL`) when your interactive shell isn't your login `$SHELL`:

```yaml
# ~/.config/kush/config.yaml
shell: /opt/homebrew/bin/fish
```

The case that catches people: you run fish all day, but `$SHELL` still says `/bin/zsh` because you changed shells without updating it. Left alone, kush forks zsh, and your history and tools like atuin write somewhere you're not looking. Pinning `shell` keeps the subshell the one you actually live in.

## A complete config.yaml

```yaml
# ~/.config/kush/config.yaml
context_lookup_locations:
  - $KUBECONFIG
  - ~/.kube/config
  - ~/.kube/configs/*

picker: fzf
shell: /opt/homebrew/bin/fish
```
