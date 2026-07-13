---
pageLayout: home
externalLinkIcon: false

config:
  - type: doc-hero
    hero:
      name: One shell, one context. Prod in one terminal, dev in another, zero bleed.
      text: Ephemeral, isolated kube-context subshells
      tagline: A private, throwaway kubeconfig per shell, deleted on exit. Auth-agnostic. No global state to corrupt.
      image: /logo.png
      actions:
        - text: Get Started →
          link: /getting-started/quick
          theme: brand
          icon: mdi:rocket-launch
        - text: View Documentation →
          link: /getting-started/overview
          theme: alt
          icon: mdi:book-open-page-variant

  - type: features
    title: Why kush?
    description: Isolation is the whole point.
    features:
      - title: One shell = one context
        icon: mdi:layers
        details: Enter a context, get a shell locked to it via a private kubeconfig. Every kube tool in that shell sees only that one context.
      - title: Zero global mutation
        icon: mdi:shield-lock-outline
        details: Never touches ~/.kube/config or your active context. Run prod and dev side by side in different terminals with no leakage between them.
      - title: Exit = gone
        icon: mdi:delete-clock
        details: The temp kubeconfig is deleted on exit, whether normal, signal, or crash (a stale-file sweep reaps anything left by a kill -9).
      - title: Auth-agnostic
        icon: mdi:key-variant
        details: Assumes your contexts are already authenticated. exec/OIDC/cloud plugins keep working; kush copies the user block verbatim.
      - title: Works with any tool
        icon: mdi:kubernetes
        details: kubectl, helm, k9s, kustomize, flux, argocd, and anything else that reads KUBECONFIG is pinned to the one context.
      - title: Built for agents too
        icon: mdi:robot-happy
        details: kush exec runs one command in an isolated context non-interactively, with a Claude Code plugin so agentic tools use it correctly.

  - type: VPListCompareCustom
    title: "Manual juggling vs. kush"
    description: "How you switch contexts today, and how kush does it"
    left:
      title: "The manual way"
      description: "Fragile, global, easy to get wrong"
      items:
        - title: "kubectl config use-context"
          description: "Mutates global state shared by every shell and tool"
        - title: "Per-tool --context flags"
          description: "Different spelling per tool; forget one and you hit the wrong cluster"
        - title: "KUBECONFIG juggling"
          description: "Hand-built single-context files, easy to misconfigure"
        - title: "Leaky by default"
          description: "The full kubeconfig stays visible to every command"
        - title: "Manual cleanup"
          description: "Temp files and switched contexts linger"

    right:
      title: "The kush way"
      description: "Isolated, ephemeral, blast-radius-contained"
      items:
        - title: "kush <ctx>"
          description: "A subshell pinned to one context; exit and it's gone"
        - title: "One uniform form"
          description: "Every tool in the shell sees exactly one context"
        - title: "Private throwaway kubeconfig"
          description: "Built for you, 0600, in $XDG_RUNTIME_DIR"
        - title: "Cannot touch other clusters"
          description: "The command physically only sees the pinned context"
        - title: "Auto-cleanup + stale sweep"
          description: "Deleted on exit; crashes reaped on next run"

  - type: custom

  - type: VPReleases
    repo: SpechtLabs/kush

  - type: VPContributors
    repo: SpechtLabs/kush
---

::: cast src=/casts/kush-home.cast title="Enter a real kind-backed prod context" rows=14
:::

That's the whole workflow: enter, work, exit. Nothing you did in there touched `~/.kube/config` or your default context; open a second terminal and run `kush dev` alongside it, zero bleed between the two.
