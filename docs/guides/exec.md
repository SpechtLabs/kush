---
title: Run one command
createTime: 2026/07/10 10:00:00
---

`kush exec` runs a single command against an isolated context, non-interactively — no subshell, no picker, nothing waiting on a TTY. It's the form to reach for from scripts, CI, and agents.

## The command

```
kush exec <context> [-n namespace] -- <command> [args...]
```

Everything after `--` is the command to run. kush pins `KUBECONFIG` to a private one-context temp file for the duration of that command, forwards stdin/stdout/stderr, propagates the child's exit code, and deletes the temp file when the command finishes:

::: terminal Run a single command
```shell
$ kush exec prod -- kubectl get pods
NAME                     READY   STATUS    RESTARTS   AGE
web-7d9f8b6c5-x2z9k      1/1     Running   0          3h

$ echo $?
0
```
:::

## Targeting a namespace

Pass `-n`/`--namespace` to pin the namespace for that one command:

::: terminal Run against a specific namespace
```shell
$ kush exec prod -n billing -- kubectl get pods
NAME                READY   STATUS    RESTARTS   AGE
worker-6f9c7d-abcde  1/1     Running   0          2h
```
:::

## Exit codes and stdio are forwarded

`kush exec` is transparent to the command it runs: stdout and stderr pass straight through, and the process exits with the same code the wrapped command did. A failing command still fails the `kush exec` call:

::: terminal Exit code propagation
```shell
$ kush exec staging -- kubectl get pod does-not-exist
Error from server (NotFound): pods "does-not-exist" not found

$ echo $?
1
```
:::

That makes it safe to chain in a script with `&&` or check in CI without any extra plumbing.

## Works with any tool

`kush exec` isn't kubectl-specific — it works with anything that reads `KUBECONFIG`:

::: terminal helm through kush exec
```shell
$ kush exec staging -n billing -- helm upgrade --install billing ./chart
Release "billing" has been upgraded. Happy Helming!
```
:::

## Use in scripts and CI

Because each invocation is a single non-interactive process with a temp kubeconfig cleaned up on exit, `kush exec` composes cleanly in scripts:

::: terminal In a CI step
```shell
$ kush exec prod -n billing -- kubectl rollout status deploy/billing
deployment "billing" successfully rolled out

$ kush exec prod -- kubectl get nodes -o name | wc -l
5
```
:::

## The agent-safe alternative

`kush ctx` and `kush ns` are interactive — they wait for a shell to exit, which will hang an agent or a non-interactive CI runner that doesn't know to send an exit. `kush exec` never spawns a shell and never waits on anything beyond the one command you gave it, which makes it the form to use whenever the caller isn't a human at a terminal.

It also beats `kubectl --context` for that use case: it's uniform across every kube tool, not just kubectl, and it gives real blast-radius containment — the command physically cannot see or touch any context other than the one you named, because no other context exists in the temp kubeconfig it's handed.
