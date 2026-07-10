---
title: Show your context in the prompt
createTime: 2026/07/10 10:00:00
---

Once you're living in kush shells, the question you'll ask constantly is "which cluster am I in right now?" `kush current` answers it on demand, but it's better to just see it. kush doesn't render a prompt of its own; it sets a handful of environment variables (and pins `KUBECONFIG`) so whatever prompt engine you already run can show the context itself.

| Variable          | Meaning                                      |
| ----------------- | -------------------------------------------- |
| `KUSH_ACTIVE`     | `1`, set whenever you're inside a kush shell |
| `KUSH_CONTEXT`    | the pinned context name                      |
| `KUSH_NAMESPACE`  | the pinned namespace                         |
| `KUSH_KUBECONFIG` | path to the private temp kubeconfig          |

Pick the section for your setup.

## starship

Nothing to do. starship's native `kubernetes` module reads `KUBECONFIG`, and kush points `KUBECONFIG` at the private single-context file for the life of the shell, so your existing kube prompt already reflects the pinned context and namespace.

## oh-my-posh

oh-my-posh's native `kubectl` segment also reads `KUBECONFIG` directly, so context and namespace show up with no kush-specific wiring. Use it as a segment or a tooltip:

```json
"tooltips": [
  {
    "type": "kubectl",
    "style": "powerline",
    "tips": ["kubectl", "helm", "kustomize", "k"],
    "template": " ⎈ {{ .Context }}{{ if .Namespace }}: {{ .Namespace }}{{ end }} "
  }
]
```

For an always-on marker that shows you're inside a kush shell before you even touch kubectl, gate a `text` segment on `KUSH_ACTIVE`:

```json
"blocks": [
  {
    "type": "prompt",
    "alignment": "left",
    "segments": [
      {
        "type": "text",
        "style": "powerline",
        "template": "{{ if eq .Env.KUSH_ACTIVE \"1\" }}⎈ {{ end }}❯ "
      }
    ]
  }
]
```

There's a [full oh-my-posh example] in [my dotfiles repo].

[my dotfiles repo]: https://github.com/cedi/dotfiles
[full oh-my-posh example]: https://github.com/cedi/dotfiles/blob/main/.config/oh-my-posh/themes/tron-cedi.omp.json

## Plain shells (no prompt engine)

Not running a prompt engine? `kush init` emits an opt-in fallback that prepends a `(kush:<ctx>)` marker to `PS1` (bash/zsh) or `fish_prompt` (fish), gated on `KUSH_ACTIVE` so it's invisible outside a kush shell:

```text
eval "$(kush init bash)"
eval "$(kush init zsh)"
kush init fish | source
```

Add the line for your shell to its startup file (`~/.bashrc`, `~/.zshrc`, or `~/.config/fish/config.fish`).

## `kush init` is not `kush completion`

Easy to mix up, since both are shell glue you `eval`:

- `kush init` sets up the prompt marker above. It's only for plain shells; skip it if you run starship or oh-my-posh.
- `kush completion` sets up tab-completion for context names, which every user should install. See [Tab-completion](./completion.md).

Neither is the safety [guard](./guard.md): the prompt tells you where you are, the guard stops you acting when you're nowhere.
