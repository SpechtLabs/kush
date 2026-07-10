---
title: Switch namespaces
createTime: 2026/07/10 10:00:00
---

Namespace is the one thing kush lets you change without leaving your shell. Context stays locked for the life of a pinned shell, that's the whole design, but you'll constantly want to move between namespaces in the same cluster, checking `billing` and then `kube-system` without re-entering anything. `kush ns` is how.

## Re-pin the namespace in place

Inside a kush shell, `kush ns <name>` rewrites the namespace in that shell's private kubeconfig where it sits. No new shell, nothing lost from your session, and it applies on the very next command:

::: terminal Re-pin the namespace

```shell
$ kush prod
$ kush ns billing
$ kubectl get pods
# querying prod/billing now, same shell

$ kush current
prod/billing
```

:::

Because it edits the existing kubeconfig instead of spawning a shell, `kush ns` is exempt from the [nesting guard](./enter-context.md). It's the one kush command you can run from inside a kush shell.

## From a normal shell, it opens one for you

Run `kush ns <name>` when you're *not* already pinned and it spawns a fresh shell for your current ambient context, set to that namespace. Read it as "give me an isolated shell on the cluster I'm already pointed at, in this namespace":

::: terminal Namespace subshell from outside

```shell
$ kubectl config current-context
staging

$ kush ns billing
# now in a subshell pinned to staging/billing
```

:::

## No name? kush asks

Leave the name off and kush prompts for one as free text:

::: terminal Prompt for a namespace

```shell
$ kush ns
namespace: billing
```

:::

It's a plain prompt, not a picker, and that's on purpose: listing namespaces would mean a call to the cluster's API, which kush never makes. You type the name you want and it re-pins.

## Why kubectl and your prompt just follow along

`kush ns` only edits the `namespace:` field in the temp kubeconfig that `KUBECONFIG` already points at. kubectl re-reads that file on every invocation, so the switch lands with no flag, and starship's `kubernetes` module reads the same file, so your prompt updates on its own. Nothing anywhere caches the old namespace.
