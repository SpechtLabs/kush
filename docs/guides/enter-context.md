---
title: Enter a context
createTime: 2026/07/10 10:00:00
---

The everyday way to work with kush is to enter a context, do your work, and leave. Entering doesn't flip a global setting; it opens a fresh shell that can only see the one cluster you named, backed by a private kubeconfig that's deleted the moment you exit. Nothing you do in there touches `~/.kube/config` or leaks into another terminal.

## The loop: enter, work, exit

::: terminal Enter, work, exit

```shell
$ kush prod
# you're now in a subshell pinned to prod

$ kubectl get pods
$ helm list

$ exit
# temp kubeconfig deleted, back where you started
```

:::

Everything between `kush prod` and `exit` runs against `prod` and nothing else. There's no "switch back" step to forget, because leaving the shell *is* the switch back. Even if you close the terminal or the process is killed instead of exiting cleanly, the kubeconfig still gets reaped (see [How isolation works](../understanding/isolation.md)).

Inside the shell, every kube-aware tool on your `PATH` (kubectl, helm, k9s, kustomize, flux, argocd) reads the same pinned `KUBECONFIG`, so they all see `prod` and only `prod`. You never pass a `--context` flag to anything.

## When you don't know the exact name

Run `kush` with no argument and pick from a list of every context it discovered:

::: terminal Pick from a list

```shell
$ kush
# opens a picker; select one and press enter
```

:::

`kush` and `kush ctx` are the same command; the bare form is just shorter. The picker is fzf if you have it, otherwise a built-in TUI (both configurable in [Config & discovery](./configuration.md)). To see what's there without entering anything, list it:

::: terminal List without entering

```shell
$ kush ctx --list
  dev
  staging
* prod
```

:::

The `*` marks your current ambient context. `--list` prints and exits; no shell is spawned.

## Switching context means exit, then re-enter

kush shells don't nest. Try to enter a second context from inside the first and it refuses:

::: terminal The nesting guard

```shell
$ kush prod
$ kush staging
already in a kush shell (prod); exit first to switch context
```

:::

This is deliberate. Switching context in place is exactly the silent state change kush exists to prevent, so there is no in-place switch: you exit `prod`, then enter `staging`. The one thing you *can* change without leaving is the namespace, with `kush ns` (see [Switch namespaces](./namespaces.md)).

## Working against two clusters at once

A pinned shell lives entirely in its own temp file, so you can keep as many open as you have terminals, each locked to a different context and none aware of the others. This is the safe way to compare environments, for example checking the same deployment in staging against prod:

::: terminal Two clusters at once

```shell
# terminal 1
$ kush prod
$ kubectl get deploy billing -o yaml

# terminal 2
$ kush staging
$ kubectl get deploy billing -o yaml
```

:::

Exit either shell and the other is untouched. Neither can act on the wrong cluster, because neither can see the other's.

## Knowing where you are

Inside a shell, `kush current` prints the active context and namespace; outside one it prints nothing and exits `0`, so it's safe to drop into a script or prompt:

::: terminal Where am I

```shell
$ kush current
prod/default
```

:::

For an always-on indicator, wire the context into your prompt with starship, oh-my-posh, or the plain-shell fallback, all covered in [Show your context in the prompt](./prompt.md). kush also exports `KUSH_CONTEXT`, `KUSH_NAMESPACE`, and a couple of siblings for anything else that wants to read them; the full list is in the [configuration reference](../reference/configuration.md).
