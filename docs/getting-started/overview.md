---
title: Overview
createTime: 2026/07/10 10:00:00
---

kush gives you ephemeral, isolated kube-context subshells. Run `kush prod` and you get a normal shell, pinned to the `prod` context through a private, throwaway kubeconfig. Exit that shell and the kubeconfig is deleted — nothing about the session persists.

## The mental model

One shell, one context. To switch context you exit the current shell and enter another; kush deliberately doesn't offer a way to change context in place, because that's exactly the kind of silent state change that gets you into trouble. The one thing you *can* change without leaving is the namespace: `kush ns foo` re-pins the current shell's kubeconfig on the spot, no new shell required.

When you exit, it's gone. The temp kubeconfig kush wrote for that shell is deleted, and nothing was ever written back to `~/.kube/config` or your regular active context. That's the whole point — you can have a `prod` shell open in one terminal and a `dev` shell open in another, and neither can bleed into the other or into your default kubectl setup.

## Versus `kubectl config use-context`

`use-context` mutates a single shared file. Switch context in one terminal and every other terminal, script, and background tool reading `~/.kube/config` switches with it. It's easy to run a command against the wrong cluster because some other pane changed context five minutes ago. kush sidesteps this by never touching that file — each shell gets its own private, minimal kubeconfig scoped to exactly one context.

## Versus `tka shell` / `kubie`

kush is a generic successor to both. `tka shell` is tied to a specific auth flow; kubie is unmaintained. kush is auth-agnostic — it copies the user block from your existing kubeconfig verbatim, so exec plugins, OIDC, and cloud auth all keep working — and it's a single Go binary with no extra runtime dependencies.

## When to use it

Reach for kush whenever you're about to work against a cluster that isn't your safe default: production, a customer environment, anything where "oops, wrong context" is expensive. For scripts, CI, and AI agents that need to run one-off commands without an interactive shell, see `kush exec`.
