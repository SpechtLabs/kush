---
title: How isolation works
createTime: 2026/07/10 10:00:00
---

kush's whole value is one guarantee: a shell pinned to `prod` cannot see `dev`, `staging`, or anything else in your kubeconfig. Not "won't by default" — cannot, because the file backing that shell's `KUBECONFIG` only contains `prod`. This page walks through how that file gets built, where it lives, and how it gets cleaned up.

## Extract-one

When you run `kush prod`, kush doesn't hand your existing kubeconfig to a subshell with a context flag bolted on. It builds a brand-new, minimal kubeconfig from scratch:

1. **Load and merge** every kubeconfig kush finds, using the same precedence rules `kubectl` itself uses (see [Configuration](../reference/configuration.md) for the search locations).
2. **Look up** the context you named (`prod`) in that merged view.
3. **Resolve** the cluster and user entries that context points to.
4. **Build a new kubeconfig** containing exactly one context (`prod`), its one cluster, and its one user — with `current-context: prod` already set. Nothing else from the merged view is copied in.

That last step is the whole trick. A subshell reading this file has no way to address `dev` even if it wanted to, because `dev`'s cluster and user were never copied over. There's no context to switch to and no server URL to point a raw `kubectl --context` flag at.

```mermaid
flowchart LR
    A["kush prod"] --> B["Load + merge\nkubeconfigs"]
    B --> C["Look up 'prod',\nresolve cluster + user"]
    C --> D["Build minimal kubeconfig:\nprod + its cluster + its user"]
    D --> E["Write to private temp file\n0600 / dir 0700"]
    E --> F["Fork $SHELL with\nKUBECONFIG=<temp file>"]
    F --> G["you work"]
    G --> H["exit"]
    H --> I["delete temp file"]
```

## The private temp kubeconfig

The generated file is written to:

```
$XDG_RUNTIME_DIR/kush/<ctx>-<pid>-<rand>.yaml
```

falling back to `$TMPDIR` when `XDG_RUNTIME_DIR` isn't set. `<ctx>` is the context name, `<pid>` is kush's process ID, and `<rand>` is a random suffix — together they make the filename unique per invocation, so two shells entering the same context at once don't collide.

Permissions are locked down at creation: the containing directory (`.../kush/`) is `0700`, and the file itself is `0600`. Only your user can read it, which matters because it holds the same auth material — tokens, exec plugin config, client certs — as your real kubeconfig.

## Auth plugins keep working

kush copies the resolved user block into the new file **verbatim**. Whatever authenticates you today — a static token, a client cert, an `exec:` plugin calling out to `aws eks get-token` or an OIDC helper, a cloud provider's credential plugin — is copied as-is, not reinterpreted or wrapped. If it worked in your original kubeconfig, it works in the temp one, because from the plugin's point of view nothing changed. kush doesn't touch credentials; it just narrows which context and cluster they're allowed to reach in that shell.

## Forking the shell

Once the file is written, kush forks `$SHELL` (or the `shell` config override — see [Configuration](../reference/configuration.md)) with `KUBECONFIG` pointed at the temp file, plus the `KUSH_*` state variables (`KUSH_ACTIVE`, `KUSH_CONTEXT`, `KUSH_NAMESPACE`, `KUSH_KUBECONFIG`) set for the lifetime of that process. Because `KUBECONFIG` is what every kube-aware tool reads — kubectl, helm, k9s, kustomize, flux, argocd, whatever else you have on `PATH` — the isolation isn't a kush feature you have to opt into per-tool. Every tool in that shell sees the same single-context file.

## Deletion on exit

The temp file is meant to outlive nothing. kush registers a deferred cleanup for the normal exit path and a signal handler for interrupts (`SIGINT`, `SIGTERM`), so whether you type `exit`, close the terminal, or hit Ctrl-C, the temp kubeconfig is removed before the process actually goes away.

That covers voluntary exits. A hard crash — the machine loses power, the process gets `SIGKILL`, something panics past the signal handler — can leave an orphaned file behind. kush handles that with a **stale-file PID sweep**: on startup, before writing its own temp file, kush scans `$XDG_RUNTIME_DIR/kush/` (or the `$TMPDIR` fallback) for leftover files, extracts the `<pid>` embedded in each filename, and checks whether that process is still alive. If it isn't, the stale file is deleted. This keeps the temp directory from accumulating orphaned kubeconfigs across crashes without needing a background daemon or cron job to do it.

## The nesting guard

Because each kush shell sets `KUSH_ACTIVE=1`, kush can tell at startup whether it's already running inside one of its own shells. If it is, entering another context is refused outright:

```
already in a kush shell (prod); exit first to switch context
```

This is a single environment-variable check, not depth tracking — kush doesn't need to know how many shells deep you are, only whether you're already pinned to something. The one exception is `kush ns`, which re-pins the namespace of the shell you're already in by editing the existing temp kubeconfig in place; it doesn't spawn a new shell, so there's nothing to nest and the guard doesn't apply.

## Why not just `kubectl --context`

`--context` still reads the full merged kubeconfig — every cluster and user you've ever configured — and trusts you (or your script, or your agent) to pass the right flag every single time. One missed flag, one copy-pasted command from the wrong terminal tab, and you're running against the wrong cluster with valid credentials for it sitting right there. Extract-one removes that failure mode structurally: the file a shell can see is the file it's allowed to touch.
