---
title: Installation
createTime: 2026/07/10 10:00:00
---

kush ships as a single Go binary. Install it with Krew, Homebrew, `go install`, or a prebuilt release binary. kush is unix-only in v1: no Windows binaries, no `cmd.exe`/PowerShell support.

:::: tabs

@tab Krew

Install kush as a kubectl plugin:

::: terminal Install with Krew

```shell
kubectl krew index add kush https://github.com/SpechtLabs/kush.git
kubectl krew install kush/kush
```

:::

Krew exposes the plugin as `kubectl kush`, so you can enter a context with `kubectl kush <context>`.

@tab Homebrew

The quickest way on macOS or Linux is the Specht Labs tap:

::: terminal Install with Homebrew

```shell
brew install spechtlabs/tap/kush
```

:::

That adds the tap and installs kush in one step, so `brew upgrade` keeps it current. The cask ships an unsigned binary and clears the macOS quarantine flag for you, so it runs without a Gatekeeper prompt.

@tab go install

Build it yourself straight from the module:

::: terminal Build from source

```shell
go install github.com/spechtlabs/kush/cmd/kush@latest
```

:::

@tab Release binary

Grab a prebuilt binary for your platform from the [GitHub releases page](https://github.com/spechtlabs/kush/releases), then move it onto your `PATH`.

@tab Nix (Flakes)

Run `kush` directly or build it using the provided Nix Flake:

::: terminal Run with Nix

```shell
nix run github:spechtlabs/kush -- version
```

:::

To add `kush` to your NixOS configuration, add the input:

```nix
inputs.kush = {
  url = "github:spechtlabs/kush";
  inputs.nixpkgs.follows = "nixpkgs";
};
```

`nixpkgs.follows` is optional but recommended — without it, kush pulls in
its own separate nixpkgs evaluation instead of reusing the one your
system already has, which costs extra build time and store space.

And install the package:

```nix
environment.systemPackages = [
  inputs.kush.packages.${pkgs.stdenv.hostPlatform.system}.default
];
```

::::

## Optional: fzf

kush ships a built-in picker (charm/huh TUI) that works out of the box with no extra dependencies. If you have [`fzf`](https://github.com/junegunn/fzf) on your `PATH`, kush uses it instead for context selection; fzf inherits your full environment, so your `FZF_DEFAULT_OPTS`, colors, and keybindings all apply. You don't need to configure anything; kush detects fzf automatically.

## Verify

::: terminal Verify the install

```shell
$ kubectl kush version
# prints the installed version and exits 0
```

:::

If that prints a version instead of "unknown command" or "command not found," you're set up. If you installed with Homebrew, `go install`, or a release binary, `kush version` works too. Head to the [quick start](./quick.md) to enter your first context.
