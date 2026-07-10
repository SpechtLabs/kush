---
title: Installation
createTime: 2026/07/10 10:00:00
---

kush ships as a single Go binary. Two ways to get it:

- **Go install**: build it yourself with `go install`, pulling directly from the module.
- **Release binaries**: grab a prebuilt binary from the [GitHub releases page](https://github.com/spechtlabs/kush/releases).

kush is unix-only in v1: no Windows binaries, no `cmd.exe`/PowerShell support.

## Optional: fzf

kush ships a built-in picker (charm/huh TUI) that works out of the box with no extra dependencies. If you have [`fzf`](https://github.com/junegunn/fzf) on your `PATH`, kush uses it instead for context selection; fzf inherits your full environment, so your `FZF_DEFAULT_OPTS`, colors, and keybindings all apply. You don't need to configure anything; kush detects fzf automatically.

## Verify

::: terminal Verify the install
```shell
$ kush version
# prints the installed version and exits 0
```
:::

If that prints a version instead of "command not found," you're set up. Head to the [quick start](./quick.md) to enter your first context.
