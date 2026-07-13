---
title: Quick Start
createTime: 2026/07/10 10:00:00
---

Two minutes, four commands.

## List and enter

::: cast src=/casts/kush-home.cast title="List contexts, enter prod, run kubectl, exit" rows=14
:::

That's the whole workflow: enter, work, exit. Nothing you did in there touched `~/.kube/config` or your default context; open a second terminal and run `kush dev` alongside it, zero bleed between the two.

## Or skip the shell entirely

For a single command (scripts, CI, agents), use `kush exec` instead of a subshell. It pins KUBECONFIG for just that one process, forwards stdin/stdout/stderr, and propagates the exit code:

::: cast src=/casts/kush-exec.cast title="Run isolated one-off commands" rows=7
:::

## Where contexts come from

kush discovers contexts from `$KUBECONFIG` and `~/.kube/config` by default. If you keep kubeconfigs in more than one place (a directory of per-cluster files, say), point kush at them with `context_lookup_locations` in `~/.config/kush/config.yaml`. That's covered in the configuration reference; for now, `kush ctx --list` will always tell you exactly what kush can see.
