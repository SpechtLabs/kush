---
name: kush
description: Use when running kubectl, helm, k9s, or any Kubernetes tool against a specific named context or cluster non-interactively — especially when the user's active context must stay unchanged, when a command must be scoped to exactly one cluster, or when operating across multiple contexts (prod/dev/staging).
---

# kush — isolated kube-context commands

## Overview

`kush exec` runs one command against exactly one Kubernetes context using a **private, throwaway kubeconfig** that holds only that context and is deleted when the command exits. It never reads or writes the user's active context, and it works for **any** kube tool — kubectl, helm, k9s, kustomize, flux, argocd, custom scripts — not just kubectl.

If `kush` is installed, prefer it for context-scoped commands over `kubectl --context` / `helm --kube-context` / hand-set `KUBECONFIG`.

## Core command

```
kush exec <context> [-n <namespace>] -- <command> [args...]
```

- Everything after `--` runs with `KUBECONFIG` pinned to a one-context temp file.
- stdin/stdout/stderr are forwarded; the child's **exit code is propagated** (non-zero on failure — check it as usual).
- The temp kubeconfig is deleted on exit.

```
kush exec prod -- kubectl get pods -n kube-system
kush exec dev -n team-a -- helm list
kush exec staging -- kubectl apply -f manifest.yaml
```

## Find context names first

```
kush ctx --list        # prints every available context, marks the current one
```

Contexts come from `$KUBECONFIG` + `~/.kube/config`, or the locations configured in `~/.config/kush/config.yaml` (which can glob many files, e.g. `~/.kube/configs/*`).

## Why prefer `kush exec`

`kubectl --context X` and `helm --kube-context X` are valid, but:

| Approach | Limitation |
|---|---|
| `kubectl config use-context X` | Mutates global `~/.kube/config` — changes the context for **every** shell and tool. Never use it to "temporarily" switch. |
| `kubectl --context X` / `helm --kube-context X` | Per-tool flag with different spelling per tool; the full kubeconfig stays visible, so a forgotten flag on the next command silently hits the default cluster. |
| `KUBECONFIG=some/file cmd` | You must guarantee the file is a single, correct context; easy to get wrong. |
| `kush exec X -- cmd` | One uniform form for **every** tool; the command physically cannot see or touch any other cluster (blast-radius containment); auto-cleanup; safe to run for prod and dev **in parallel**. |

## Do NOT use the interactive commands from an agent

`kush ctx`, `kush <ctx>`, and `kush ns` (without args) fork an **interactive subshell** and will hang a non-interactive agent waiting on a prompt. For agentic/scripted use, always use `kush exec`. (`kush ctx --list` is fine — it just prints and exits.)

## Other commands

- `kush lint` — report common kubeconfig problems (contexts referencing a missing cluster/user, empty current-context); exits non-zero when it finds errors.
- `kush split [-o <dir>]` — write one self-contained kubeconfig per context.

## Common mistakes

- Using `kush ctx <name>` in a script expecting it to run a command — that opens a shell and blocks. Use `kush exec <name> -- <command>`.
- Guessing a context name — run `kush ctx --list` first; names must match exactly.
- Forgetting the `--` before the command in `kush exec`.
