---
title: OpenShift
createTime: 2026/07/11 10:00:00
---

kush works with OpenShift out of the box. There is nothing OpenShift-specific to install, configure, or flag: because kush copies your kubeconfig's `cluster`, `context`, and `user` blocks verbatim into the throwaway kubeconfig, whatever authenticates you against an OpenShift cluster today keeps working inside a kush shell — OAuth bearer tokens, the cluster CA, `oc`'s exec credential plugins, and OIDC all carry over untouched. And because kush only pins `KUBECONFIG` and forks your shell, `oc` honors it exactly like `kubectl` does.

This guide covers the everyday loop, the one real gotcha (token expiry), and how to make long sessions completely seamless.

## The loop: log in, enter, work, exit

Authenticate once the way you always do — `oc login` writes the context into `~/.kube/config` — then hand that context to kush:

::: terminal Log in, then enter

```shell
$ oc login https://api.mycluster.example.com:6443
# populates ~/.kube/config as usual

$ kush default/api-mycluster-example-com:6443/kube:admin
# subshell pinned to that context, via a private kubeconfig

$ oc get pods
$ oc get routes

$ exit
# temp kubeconfig deleted, ~/.kube/config never touched
```

:::

Every `oc` command between entering and `exit` runs against that one cluster and nothing else. As with kubectl, you never pass a `--context` flag to anything — `oc`, `odo`, `helm`, and `k9s` all read the same pinned `KUBECONFIG`.

OpenShift's default context names are long — `default/api-mycluster-example-com:6443/kube:admin` — so most people run `kush` with no argument and pick from the list instead of typing it:

::: terminal Pick instead of typing

```shell
$ kush
# opens the picker (fzf if installed, otherwise the built-in TUI)
```

:::

## Projects are namespaces

An OpenShift *project* is a Kubernetes namespace with extra metadata, so kush's namespace handling applies directly. Inside a kush shell you can re-pin with either tool:

::: terminal Switch project in place

```shell
kush ns billing        # kush's own re-pin
oc project billing     # oc's equivalent — both edit the same temp kubeconfig
```

:::

Both edit the current context in the temp kubeconfig, so the change is isolated to this shell and thrown away on exit — exactly like a namespace switch. Your `(kush:…)` prompt marker reads the namespace live from the kubeconfig file, so it updates whichever command you use. See [Switch namespaces](./namespaces.md) for the full picture.

## The one gotcha: token expiry in long sessions

If you authenticate with a username and password (`oc login -u … -p …`), OpenShift issues a static OAuth bearer token — typically valid for about 24 hours — and bakes it into your kubeconfig's `user` block. kush snapshots that token into the temp kubeconfig **when you enter the shell**. Two things follow from that:

- A kush shell you leave open long enough will start getting `401 Unauthorized` once the token expires.
- Refreshing the token in *another* terminal (a fresh `oc login` there) does not reach back into this shell's snapshot.

The fix is simple: `oc login` again from *inside* the kush shell. `oc` respects the pinned `KUBECONFIG`, so it rewrites the temp file in place — your session refreshes and stays just as isolated and disposable as before.

::: terminal Refresh without leaving the shell

```shell
$ oc get pods
error: You must be logged in to the server (Unauthorized)

$ oc login https://api.mycluster.example.com:6443
# rewrites the temp kubeconfig for THIS shell only

$ oc get pods
NAME                     READY   STATUS    RESTARTS   AGE
web-7d9f8b6c5-x2z9k      1/1     Running   0          3h
```

:::

## Making long sessions seamless: use a refreshing credential

To sidestep expiry entirely, authenticate with a credential that refreshes on demand instead of a static token. Configure OIDC against your cluster's OAuth server through an `oc`/`kubectl` [exec credential plugin](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#client-go-credential-plugins) (for example `kubelogin` / `oc-oidc`). Your `user` block then becomes an `exec:` stanza rather than a hard-coded token:

```yaml
users:
  - name: oidc/api-mycluster-example-com:6443
    user:
      exec:
        apiVersion: client.authentication.k8s.io/v1
        command: kubectl
        args: ["oidc-login", "get-token", "--oidc-issuer-url=…", "--oidc-client-id=…"]
```

kush copies exec stanzas verbatim, so **every** kush shell you open gets fresh, auto-refreshing tokens with no re-login step. This is a cluster/kubeconfig setup choice — it needs no kush configuration — and it's the recommended setup if you keep pinned shells open for hours.

Because the exec plugin config lives in the temp kubeconfig with the same `0600`/`0700` permissions kush applies to everything else, your credentials are no more exposed than in your real kubeconfig (see [How isolation works](../understanding/isolation.md)).

## In short

- **Nothing to install or configure** — `oc login`, then `kush <context>`, then use `oc` normally.
- **Projects behave like namespaces** — `kush ns` and `oc project` are interchangeable and both stay isolated.
- **Static tokens expire** — re-run `oc login` inside the shell to refresh in place.
- **For friction-free long sessions** — use an OIDC exec credential; kush carries it into every shell automatically.
