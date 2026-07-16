---
title: Configuration
createTime: 2026/07/10 10:00:00
---

kush reads an optional config file via [viper](https://github.com/spf13/viper). An absent config is not an error; kush falls back to sane defaults. A malformed one is: you get a clear error at startup instead of kush silently ignoring it.

## Config file location

kush looks for `config.yaml` in, in order:

| Location                 |
| ------------------------ |
| `~/.config/kush/`        |
| `$XDG_CONFIG_HOME/kush/` |
| `/etc/kush/`             |

## Config keys

| Key                              | Type                                  | Default                                | Env override  |
| -------------------------------- | ------------------------------------- | -------------------------------------- | ------------- |
| `context_lookup_locations`       | list of strings                       | `[$KUBECONFIG, ~/.kube/config]`        | none          |
| `picker`                         | string (`auto` \| `builtin` \| `fzf`) | `auto`                                 | `KUSH_PICKER` |
| `shell`                          | string                                | `""` (uses `$SHELL`, then `/bin/bash`) | `KUSH_SHELL`  |
| `pre_exec_hook`                  | list of strings                       | `[]`                                   | none          |
| `post_exec_hook`                 | list of strings                       | `[]`                                   | none          |
| `contexts.<name>.pre_exec_hook`  | list of strings                       | `[]`                                   | none          |
| `contexts.<name>.post_exec_hook` | list of strings                       | `[]`                                   | none          |

### `context_lookup_locations`

Ordered list of places kush looks for kubeconfigs. Each entry supports `$ENV`/`${ENV}` expansion, `~`, and globs (use `*`, not `*.yaml`, for files that have no extension).

When this key is set and non-empty, it **replaces** the default lookup entirely. If it's absent, empty, or matches nothing, kush falls back to the default: `$KUBECONFIG` (its entries, in order, if it contains multiple `:`-separated paths) plus `~/.kube/config`.

Contexts are merged first-wins; list order is precedence. If the same context name is defined in more than one file, kush emits a warning to stderr and uses the first file it saw:

```text
warning: context "prod" defined in 2 files; using ~/.kube/config
```

A file matched by a glob that isn't a valid kubeconfig is skipped, not fatal:

```text
warning: skipping ~/.kube/configs/notes.txt: not a valid kubeconfig
```

### `picker`

Controls which context picker `kush` (no argument) and `kush ctx` (no argument) open.

| Value     | Behavior                                                           |
| --------- | ------------------------------------------------------------------ |
| `auto`    | Use `fzf` if it's on `PATH`, otherwise the built-in charm/huh TUI. |
| `builtin` | Always use the built-in TUI, even if `fzf` is installed.           |
| `fzf`     | Always use `fzf`; errors clearly if it isn't installed.            |

`fzf`, when used, inherits your full environment; `FZF_DEFAULT_OPTS`, colors, and keybindings all apply as normal. kush passes it only `--prompt`, `--no-multi`, and `--print-query`.

### `shell`

The shell kush forks when entering a context. Empty string (the default) means "use `$SHELL`, falling back to `/bin/bash` if that's unset." Set this explicitly when your interactive shell differs from your login `$SHELL` (for example, you run `fish` day to day but `$SHELL` is still set to `/bin/zsh`), so that subshell command history and tools like atuin land where you actually expect them.

### `pre_exec_hook`

An ordered list of shell commands to run after kush knows the target context, but before it creates the isolated kubeconfig and starts the subshell or `kush exec` command. If a hook exits non-zero, kush aborts without running the remaining hooks.

Use this for context-specific authentication that must happen before entering the isolated environment:

```yaml
pre_exec_hook:
  - "tsh join $KUSH_CONTEXT"
  - "printf 'authenticated to %s\\n' $KUSH_CONTEXT"
```

Hooks run through your configured `shell`, then `$SHELL`, then `/bin/sh`. They inherit stdin/stdout/stderr, so interactive auth prompts work. kush also adds these environment variables for the hook:

| Variable         | Meaning                            |
| ---------------- | ---------------------------------- |
| `KUSH_CONTEXT`   | Target context name.               |
| `KUSH_NAMESPACE` | Target namespace, if one is known. |

Per-context hooks override the global list:

```yaml
pre_exec_hook:
  - "tsh join $KUSH_CONTEXT"

contexts:
  cluster-123:
    pre_exec_hook:
      - "tsh join cluster-123"
```

After a successful hook, kush reloads kubeconfig before extracting the isolated context, so hooks that refresh kubeconfig-backed auth state are reflected in the shell or command.

### `post_exec_hook`

An ordered list of shell commands to run after kush creates the isolated kubeconfig and starts the child shell. Post-exec hooks have `KUBECONFIG` and all `KUSH_*` state variables available. They run in the same shell initialization process and in the configured order, so exported environment variables and working-directory changes carry into the interactive shell:

```yaml
post_exec_hook:
  - "export CLUSTER_ENV=$KUSH_CONTEXT"
```

Post-exec hooks apply only to interactive shells opened by `kush ctx`, bare `kush`, or `kush ns` outside an existing kush shell. `kush exec` does not open a shell, so it does not run them. Per-context post-exec hooks override the global post-exec list independently of pre-exec hooks.

## State environment variables

These are not config; kush _sets_ them inside every subshell it forks (both `kush ctx` and `kush ns`), for the process consuming them to read:

| Variable          | Meaning                                                                                              |
| ----------------- | ---------------------------------------------------------------------------------------------------- |
| `KUSH_ACTIVE`     | `1` whenever you're inside a kush shell. Used by the nesting guard and by `kush init`'s prompt glue. |
| `KUSH_CONTEXT`    | The pinned context name.                                                                             |
| `KUSH_NAMESPACE`  | The pinned namespace.                                                                                |
| `KUSH_KUBECONFIG` | Path to the private temp kubeconfig backing this shell.                                              |

`KUBECONFIG` itself is also set, pointed at the same temp file; that's what makes kubectl, helm, and every other kube-aware tool honor the isolation without any kush-specific flag.

## Example config.yaml

```yaml
# ~/.config/kush/config.yaml

# Where kush looks for kubeconfigs, in precedence order.
# Omit this key entirely to use the default: $KUBECONFIG + ~/.kube/config.
context_lookup_locations:
  - $KUBECONFIG # splits on ':' if it has multiple paths
  - ~/.kube/config
  - ~/.kube/configs/* # glob, picks up every file in the directory
  - ${WORK_KUBECONFIGS}/* # env-var expansion works here too

# Which picker to open when you run `kush` or `kush ctx` with no argument.
# auto (default) = fzf if installed, else the built-in TUI.
picker: auto

# Shell kush forks for interactive subshells.
# Empty string (default) = $SHELL, falling back to /bin/bash.
shell: /usr/local/bin/fish

# Optional commands to run before entering any context.
pre_exec_hook:
  - "tsh join $KUSH_CONTEXT"

# Optional commands to initialize each interactive kush shell.
post_exec_hook:
  - "export CLUSTER_ENV=$KUSH_CONTEXT"

# Per-context settings override the corresponding global hook list.
contexts:
  cluster-123:
    pre_exec_hook:
      - "tsh join cluster-123"
    post_exec_hook:
      - "export CLUSTER_ENV=production"
```
