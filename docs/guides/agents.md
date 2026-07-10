---
title: Agents & the plugin
createTime: 2026/07/10 10:00:00
---

kush's interactive commands assume a human is on the other end. An agentic tool like Claude Code isn't — so it needs a different entry point.

## Use `kush exec`, never the interactive commands

`kush ctx <name>` (and the bare `kush <name>` shorthand) forks an interactive subshell and waits there until you type `exit`. `kush ns` with no argument does the same thing via a free-text prompt. From an agent, both of those just hang: there's no human to answer the prompt or type `exit`, so the command never returns.

`kush exec` is built for exactly this case:

```
kush exec <context> [-n namespace] -- <command> [args...]
```

It runs one command against an isolated context, non-interactively: pins `KUBECONFIG` to a private one-context temp file, forwards stdin/stdout/stderr, propagates the child's exit code, and deletes the temp file when the command exits. No subshell, no prompt, no hang.

::: terminal Run a command against an isolated context
```shell
$ kush exec prod -- kubectl get pods -n kube-system
NAME                       READY   STATUS    RESTARTS   AGE
coredns-abc123             1/1     Running   0          4d
```
:::

## Why `kush exec` beats `kubectl --context`

An agent could just pass `--context` on every kubectl invocation instead. `kush exec` is still the better choice, for two reasons:

- **Uniform across tools.** `kubectl --context`, `helm --kube-context`, and whatever flag k9s or argocd use are all spelled differently. `kush exec <ctx> -- <command>` is one form that works for any kube tool, so an agent doesn't need tool-specific logic for context scoping.
- **Blast-radius containment.** A `--context` flag is a hint layered on top of a kubeconfig that still contains every other cluster — one dropped flag on a later command in the same session silently hits the wrong one. `kush exec`'s private kubeconfig contains only the target context; the command physically cannot see or touch anything else, dropped flag or not.

## The Claude Code plugin

This repository is also a Claude Code plugin marketplace. It exposes a `kush` plugin bundling a skill that teaches agentic tools to reach for `kush exec` and avoid the interactive commands.

::: terminal Install the plugin
```shell
$ /plugin marketplace add SpechtLabs/kush
$ /plugin install kush
```
:::

The bundled skill lives at `plugins/kush/skills/kush/SKILL.md` in this repo — it's what Claude Code loads to know when `kush exec` applies and which commands to avoid.
