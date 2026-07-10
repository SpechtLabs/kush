---
title: CLI Reference
createTime: 2026/07/10 10:00:00
---

Every kush subcommand, its flags, and what it does. For a narrative walkthrough of any of these, see the [Guides](../guides/enter-context.md).

## `kush [context]` / `kush ctx [name]`

**Synopsis**

```text
kush [context] [-l|--list]
kush ctx [context] [-l|--list]
```

**Description**

Enters an ephemeral subshell pinned to `<context>`, isolated via a private one-context kubeconfig (see [How isolation works](../understanding/isolation.md)). Both forms are equivalent. Called with no argument, kush opens the interactive picker over every discovered context instead of requiring you to know the name.

**Flags**

| Flag | Description |
| --- | --- |
| `-l`, `--list` | Print every discovered context, marking the current one, and exit. No subshell is spawned. |

::: terminal Enter a context

```shell
$ kush prod
# you're now in a subshell pinned to prod

$ kubectl config current-context
prod

$ exit
# back in your normal shell
```

:::

::: terminal List contexts

```shell
$ kush ctx --list
  dev
  staging
* prod
```

:::

## `kush ns [name]`

**Synopsis**

```text
kush ns [name]
```

**Description**

Behavior depends on whether you're already inside a kush shell:

- **Inside** a kush shell: re-pins the namespace of the *current* shell in place by editing its temp kubeconfig. Immediate, no new shell spawned, and exempt from the nesting guard.
- **Outside** a kush shell: spawns a new subshell for your *current* kubectl context, pinned to the given namespace.

Called with no argument, prompts for a namespace as free text (namespace listing isn't offered; kush never makes a cluster API call).

**Flags**

None.

::: terminal Re-pin namespace in place

```shell
$ kush prod
$ kush ns kube-system
# same shell, now pinned to prod/kube-system

$ kubectl get pods
# lists pods in kube-system
```

:::

## `kush exec <context> [-n namespace] -- <command> [args...]`

**Synopsis**

```text
kush exec <context> [-n|--namespace <namespace>] -- <command> [args...]
```

**Description**

Runs a single command against an isolated context, non-interactively, with no subshell and no picker. Everything after `--` is passed through as the command and its arguments. `KUBECONFIG` is pinned to a private one-context file for the duration of the call, stdin/stdout/stderr are forwarded, the child's exit code is propagated, and the temp kubeconfig is deleted when the command finishes. Built for scripts, CI, and AI agents that need to touch exactly one cluster without an interactive shell to hang on.

**Flags**

| Flag | Description |
| --- | --- |
| `-n`, `--namespace` | Namespace to pin for this invocation. |

::: terminal Run one command

```shell
$ kush exec prod -- kubectl get pods
NAME                    READY   STATUS    RESTARTS   AGE
web-7d9f8c9b5f-4x2kq    1/1     Running   0          3d

$ echo $?
0
```

:::

::: terminal With a namespace

```shell
$ kush exec prod -n kube-system -- kubectl get pods
NAME                       READY   STATUS    RESTARTS   AGE
coredns-5d78c9869d-2f8j9   1/1     Running   0          12d
```

:::

## `kush current`

**Synopsis**

```text
kush current
```

**Description**

Prints the active context and namespace. Outside a kush shell, prints nothing and exits `0`, so it's safe to call unconditionally from a script or prompt.

**Flags**

None.

::: terminal Check active state

```shell
$ kush current
prod/default

$ exit
$ kush current
# (empty, not in a kush shell)
```

:::

## `kush lint`

**Synopsis**

```text
kush lint
```

**Description**

Checks every discovered kubeconfig for common problems: contexts referencing a missing cluster or user, an empty `current-context`, and unreachable file entries. Read-only (it never modifies anything), and exits non-zero if it finds errors, so it's suitable for a CI check.

**Flags**

None.

::: terminal Lint kubeconfigs

```shell
$ kush lint
error: context "staging" references missing cluster "staging-eks"
warning: skipping ~/.kube/configs/broken.yaml: not a valid kubeconfig

$ echo $?
1
```

:::

## `kush split [-o dir]`

**Synopsis**

```text
kush split [-o|--out <dir>]
```

**Description**

Writes one self-contained kubeconfig per discovered context into a target directory, each file containing only that context's cluster and user: the same extract-one shape kush uses internally, but persisted to disk instead of a temp file. Never mutates the source kubeconfig(s). Colliding sanitized filenames are disambiguated automatically.

**Flags**

| Flag | Description |
| --- | --- |
| `-o`, `--out` | Output directory. Defaults to `~/.kube/kush`. |

::: terminal Split into per-context files

```shell
$ kush split
wrote ~/.kube/kush/dev.yaml
wrote ~/.kube/kush/staging.yaml
wrote ~/.kube/kush/prod.yaml
```

:::

::: terminal Custom output directory

```shell
$ kush split -o ./kubeconfigs
wrote kubeconfigs/dev.yaml
wrote kubeconfigs/staging.yaml
wrote kubeconfigs/prod.yaml
```

:::

## `kush init <bash|zsh|fish>`

**Synopsis**

```text
kush init <bash|zsh|fish>
```

**Description**

Emits opt-in prompt-fallback shell glue that prepends a `(kush:<ctx>)` marker to your prompt, gated on `KUSH_ACTIVE` so it's a no-op outside a kush shell. Only needed if your prompt engine doesn't already read `KUBECONFIG` or the `KUSH_*` env vars directly; starship's native `kubernetes` module, for example, needs no glue at all.

**Flags**

None.

::: terminal Wire up bash or zsh

```shell
$ eval "$(kush init bash)"
# add the same line to ~/.bashrc or ~/.zshrc to make it permanent
```

:::

::: terminal Wire up fish

```shell
$ kush init fish | source
# add the same line to config.fish to make it permanent
```

:::

## `kush completion <bash|zsh|fish>`

**Synopsis**

```text
kush completion <bash|zsh|fish>
```

**Description**

Standard cobra shell-completion script. Once installed, `kush ctx <TAB>`, bare `kush <TAB>`, and `kush exec <TAB>` complete detected context names. `kush ns` stays free-text; completing namespaces would require a cluster API call, which kush never makes.

**Flags**

None.

::: terminal Install completion (bash)

```shell
$ kush completion bash > /etc/bash_completion.d/kush
# restart your shell, then:
$ kush ctx <TAB>
dev      staging  prod
```

:::

## `kush version`

**Synopsis**

```text
kush version
kush --version
```

**Description**

Prints the installed kush version and exits `0`.

**Flags**

None.

::: terminal Check the version

```shell
$ kush version
kush dev
```

:::
