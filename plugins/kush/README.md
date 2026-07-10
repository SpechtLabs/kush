# kush plugin

A Claude Code and Codex plugin that teaches agentic tools to use [`kush`](https://github.com/SpechtLabs/kush) — a CLI for ephemeral, isolated kube-context subshells.

It bundles one skill (`skills/kush/SKILL.md`) which steers agents toward
`kush exec <context> -- <command>` for isolated, non-interactive commands
against a named Kubernetes context, and away from the interactive subshell
commands (`kush ctx`, `kush ns`) that would hang a non-interactive agent.

Requires the `kush` binary on `PATH`. See the [main repo](https://github.com/SpechtLabs/kush) for installation.

## Claude Code install

```text
/plugin marketplace add SpechtLabs/kush
/plugin install kush
```

## Codex install

```text
codex plugin marketplace add .agents/plugins
codex plugin install kush
```
