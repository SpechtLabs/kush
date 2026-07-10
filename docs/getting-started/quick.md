---
title: Quick Start
createTime: 2026/07/10 10:00:00
---

Two minutes, four commands.

## List your contexts

::: terminal See what's available
```shell
$ kush ctx --list
# prints every discovered context, marks the current one, no subshell
```
:::

## Enter one

::: terminal Enter a context
```shell
$ kush prod
# you're now in a subshell pinned to prod

$ kubectl get pods -n default
# kubectl, helm, k9s, flux: every kube tool in this shell sees only prod

$ exit
# temp kubeconfig deleted, back in your normal shell
```
:::

That's the whole workflow: enter, work, exit. Nothing you did in there touched `~/.kube/config` or your default context; open a second terminal and run `kush dev` alongside it, zero bleed between the two.

## Or skip the shell entirely

For a single command (scripts, CI, agents), use `kush exec` instead of a subshell. It pins KUBECONFIG for just that one process, forwards stdin/stdout/stderr, and propagates the exit code:

::: terminal One-off command against an isolated context
```shell
$ kush exec prod -- kubectl get pods -n kube-system
```
:::

## Where contexts come from

kush discovers contexts from `$KUBECONFIG` and `~/.kube/config` by default. If you keep kubeconfigs in more than one place (a directory of per-cluster files, say), point kush at them with `context_lookup_locations` in `~/.config/kush/config.yaml`. That's covered in the configuration reference; for now, `kush ctx --list` will always tell you exactly what kush can see.
