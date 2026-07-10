---
title: Switch namespaces
createTime: 2026/07/10 10:00:00
---

`kush ns` changes the active namespace. What it actually does depends on whether you're already inside a kush shell.

## Inside a kush shell: re-pin in place

If you're inside a kush shell, `kush ns <name>` edits the private temp kubeconfig for that shell directly. No new shell is spawned — you stay exactly where you are, and the change takes effect immediately:

::: terminal Re-pin the namespace
```shell
$ kush prod
$ kubectl config current-context
prod

$ kush ns billing
$ kubectl get pods
# now querying the billing namespace, same shell, same context

$ kush current
prod/billing
```
:::

Because this is an in-place edit rather than a new subshell, `kush ns` is exempt from the nesting guard described in [Enter a context](./enter-context.md) — it never triggers the "already in a kush shell" error.

## Outside a kush shell: spawn a subshell

If you run `kush ns <name>` from a normal shell — not inside any kush pin — kush spawns a new subshell for whatever your *current* kube-context is, pinned to that namespace:

::: terminal Namespace subshell from outside
```shell
$ kubectl config current-context
staging

$ kush ns billing
# spawns a subshell pinned to context "staging", namespace "billing"

$ kush current
staging/billing
```
:::

## No argument: free-text prompt

Leave off the name and kush asks for one:

::: terminal Prompt for a namespace
```shell
$ kush ns
namespace: billing
# same behavior as `kush ns billing` from here
```
:::

This is a free-text prompt, not a picker — kush never calls the cluster API to list namespaces, so there's nothing to autocomplete against.

## Why kubectl and starship just pick it up

`kush ns` edits the `namespace:` field of the context entry in the temp kubeconfig that `KUBECONFIG` already points at. kubectl reads that file on every invocation, so the new namespace applies immediately with no extra flag. starship's native `kubernetes` module reads the same file, so your prompt updates too, without any kush-specific prompt integration.
