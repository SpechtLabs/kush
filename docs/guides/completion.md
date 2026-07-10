---
title: Completion & prompt
createTime: 2026/07/10 10:00:00
---

kush has two separate opt-in integrations with your shell: tab-completion for context names, and a prompt indicator so you can see which context you're in. They come from two different commands — don't confuse them.

## Tab-completion: `kush completion`

```
kush completion bash
kush completion zsh
kush completion fish
```

This is a standard cobra shell-completion script, generated for your shell and loaded the usual way:

::: terminal Install completion (bash)
```shell
$ echo 'source <(kush completion bash)' >> ~/.bashrc
$ source ~/.bashrc
```
:::

::: terminal Install completion (zsh)
```shell
$ echo 'source <(kush completion zsh)' >> ~/.zshrc
$ source ~/.zshrc
```
:::

::: terminal Install completion (fish)
```shell
$ kush completion fish > ~/.config/fish/completions/kush.fish
```
:::

Once it's loaded, context names complete wherever kush expects one:

::: terminal Tab-complete a context
```shell
$ kush ctx <TAB>
dev  prod  staging

$ kush <TAB>
dev  prod  staging

$ kush exec <TAB>
dev  prod  staging
```
:::

`kush ns` stays free-text — completing namespaces would mean kush making a call to the cluster's API, and kush never does that.

## Prompt integration

kush sets four environment variables for the lifetime of a subshell, so any prompt engine that reads them (or reads `KUBECONFIG`, which is also pinned) can show your context without kush having to render anything itself:

| Variable | Meaning |
| --- | --- |
| `KUSH_ACTIVE` | `1` — set whenever you're inside a kush shell |
| `KUSH_CONTEXT` | the pinned context name |
| `KUSH_NAMESPACE` | the pinned namespace |
| `KUSH_KUBECONFIG` | path to the private temp kubeconfig |

### starship

starship's native `kubernetes` module reads `KUBECONFIG` and renders context and namespace from it directly. Since kush points `KUBECONFIG` at the private, single-context temp file for the lifetime of the subshell, this works out of the box — no starship config changes needed.

### oh-my-posh

oh-my-posh doesn't read `KUBECONFIG` the same way starship does, so add a segment that reads `KUSH_CONTEXT` and `KUSH_NAMESPACE` directly instead.

### Plain shells (no prompt engine)

If you're not running starship or oh-my-posh, `kush init` emits an opt-in fallback that prepends a `(kush:<ctx>)` marker to `PS1` (bash/zsh) or `fish_prompt` (fish), gated on `KUSH_ACTIVE` so it's invisible outside a kush shell:

```
eval "$(kush init bash)"
eval "$(kush init zsh)"
kush init fish | source
```

Add the appropriate line to your shell's startup file (`~/.bashrc`, `~/.zshrc`, or `~/.config/fish/config.fish`).

## `kush init` vs `kush completion` — not the same thing

They solve different problems:

- `kush completion` — tab-completion for arguments on the `kush` command itself. Every user should install this.
- `kush init` — an opt-in prompt-marker fallback for plain shells with no prompt engine. If you already use starship or oh-my-posh, you don't need it.
