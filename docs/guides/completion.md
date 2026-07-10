---
title: Tab-completion
createTime: 2026/07/10 10:00:00
---

`kush completion` generates a shell-completion script so context names complete on TAB. Install it once and you stop typing (or misremembering) full context names.

```text
kush completion bash
kush completion zsh
kush completion fish
```

This is a standard cobra completion script, generated for your shell and loaded the usual way:

:::: tabs

@tab bash

::: terminal Install completion (bash)

```shell
echo 'source <(kush completion bash)' >> ~/.bashrc
source ~/.bashrc
```

:::

@tab zsh

::: terminal Install completion (zsh)

```shell
echo 'source <(kush completion zsh)' >> ~/.zshrc
source ~/.zshrc
```

:::

@tab fish

::: terminal Install completion (fish)

```shell
kush completion fish > ~/.config/fish/completions/kush.fish
```

:::

::::

Once it's loaded, context names complete wherever kush expects one:

::: terminal Tab-complete a context

```shell
$ kush ctx [TAB]
dev  prod  staging

$ kush [TAB]
dev  prod  staging

$ kush exec [TAB]
dev  prod  staging
```

:::

`kush ns` stays free-text; completing namespaces would mean kush making a call to the cluster's API, and kush never does that.

To show the active context in your prompt rather than complete it, see [Show your context in the prompt](./prompt.md).
