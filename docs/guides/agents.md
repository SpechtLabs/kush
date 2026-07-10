---
title: Agent Plugins
createTime: 2026/07/10 10:00:00
---

kush ships a Claude Code plugin so agentic tools reach for `kush exec` (isolated, non-interactive) instead of the interactive subshell commands that would hang them. This page covers installing it; for why an agent should prefer `kush exec`, see [Run one command](./exec.md).

## Install the plugin

This repository is a Claude Code plugin marketplace. Add it, then install the `kush` plugin:

```text
/plugin marketplace add SpechtLabs/kush
/plugin install kush
```

## What it installs

The plugin bundles one skill, at `plugins/kush/skills/kush/SKILL.md`. It teaches Claude Code when `kush exec <context> -- <command>` applies and which interactive commands (`kush ctx`, `kush ns`) to avoid, so an agent runs against exactly one cluster without waiting on a shell that never exits.

The skill drives kush, it doesn't bundle it, so the `kush` binary still needs to be on `PATH`. See [Installation](../getting-started/installation.md).
