---
title: Enter a context
createTime: 2026/07/10 10:00:00
---

Entering a context is the core kush workflow: you get a shell locked to one kube-context, with a private throwaway kubeconfig that disappears the moment you exit.

## The command

```
kush [context]
kush ctx [context]
```

Both forms are equivalent. Pass a context name and kush spawns a subshell pinned to it:

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

Inside that subshell, every kube tool — kubectl, helm, k9s, kustomize, flux, argocd, whatever you have on PATH — reads `KUBECONFIG` and sees only `prod`. There is no other context in that file for a tool to accidentally target.

## No argument: the picker

Run `kush` or `kush ctx` with no name and you get an interactive picker over every context kush discovered:

::: terminal Pick a context
```shell
$ kush
# opens a picker (fzf if it's on PATH, otherwise the built-in TUI)
# select "staging" and press enter

$ kubectl config current-context
staging
```
:::

Which picker you get is controlled by the `picker` setting in `~/.config/kush/config.yaml` (`auto`, `builtin`, or `fzf`), or the `KUSH_PICKER` env var.

## Listing contexts without entering one

`kush ctx --list` (short form `kush ctx -l`) prints every discovered context and marks the current one, then exits — no subshell is spawned:

::: terminal List contexts
```shell
$ kush ctx --list
  dev
  staging
* prod
```
:::

## The nesting guard

kush shells don't nest. If you're already inside one and try to enter another, kush refuses:

::: terminal Nesting guard
```shell
$ kush prod
$ kush staging
already in a kush shell (prod); exit first to switch context
```
:::

Exit the current shell first, then enter the new context. This is a single `KUSH_ACTIVE` check at startup, not depth tracking — kush doesn't need to count how deep you are, it just refuses to double-pin. (`kush ns`, which re-pins the namespace of the shell you're already in, is exempt from this guard — see [Switch namespaces](./namespaces.md).)

## What gets set inside the shell

Entering a context sets four environment variables for the lifetime of the subshell:

| Variable | Meaning |
| --- | --- |
| `KUSH_ACTIVE` | `1` — set whenever you're inside a kush shell |
| `KUSH_CONTEXT` | the pinned context name |
| `KUSH_NAMESPACE` | the pinned namespace |
| `KUSH_KUBECONFIG` | path to the private temp kubeconfig |

`KUBECONFIG` itself also points at that temp file, which is how kubectl, helm, and every other kube tool pick up the pin without any extra flags.

## Checking what's active

`kush current` prints the active context and namespace. Outside a kush shell it prints nothing and exits `0`:

::: terminal Check active state
```shell
$ kush current
prod/default

$ exit
$ kush current
# (empty — not in a kush shell)
```
:::

## Multiple contexts, multiple terminals

Because the pin lives in a private temp file and kush never touches `~/.kube/config` or your shell's ambient context, you can run `prod` in one terminal and `dev` in another at the same time with zero bleed between them:

::: terminal Two contexts side by side
```shell
# terminal 1
$ kush prod
$ kubectl config current-context
prod

# terminal 2
$ kush dev
$ kubectl config current-context
dev
```
:::

Exiting either shell deletes its temp kubeconfig and leaves the other untouched.
