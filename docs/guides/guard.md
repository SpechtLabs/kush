---
title: Guard kubectl outside kush
createTime: 2026/07/10 10:00:00
---

kush contains the blast radius _inside_ a shell: a pinned shell physically can't reach another cluster. It does nothing about the `kubectl` you run in a plain terminal, against whatever `~/.kube/config` currently points at. If "wrong context" is the failure you're trying to kill, close that gap too by wrapping `kubectl` so it refuses to run outside a kush shell.

The mechanism is the `KUSH_ACTIVE` variable kush sets in every shell it forks. A thin wrapper checks it and bails on anything that operates on a cluster, while letting local meta commands (`config`, `completion`, `version`, and friends) through so your shell still behaves normally when you're not pinned.

::: tabs

@tab fish

## fish

Drop this in `~/.config/fish/conf.d/kush.fish`:

```fish
if type -q kush
    function kubectl --wraps kubectl
        # Meta/config commands that don't operate on workloads stay allowed.
        set -l safe config completion version options plugin

        if test (count $argv) -eq 0
            or contains -- "$argv[1]" $safe
            or contains -- -h $argv
            or contains -- --help $argv
            command kubectl $argv
            return
        end

        if not set -q KUSH_ACTIVE; or test "$KUSH_ACTIVE" != 1
            echo "kubectl blocked: enter a context first with 'kush <ctx>', or use 'kush exec'" >&2
            return 1
        end

        command kubectl $argv
    end
end
```

Now `kubectl get pods` in a bare shell tells you to enter kush first, while inside `kush prod` it runs as usual.

You can find a [full guard example] in [my dotfiles repo].

[my dotfiles repo]: https://github.com/cedi/dotfiles
[full guard example]: https://github.com/cedi/dotfiles/blob/main/.config/fish/conf.d/kubernetes.fish#L53-L69

@tab bash/zsh

## bash / zsh

Same shape as a function in `~/.bashrc` or `~/.zshrc`:

```bash
kubectl() {
  case "${1:-}" in
    config|completion|version|options|plugin|"" )
      command kubectl "$@"; return ;;
  esac
  case " $* " in
    *" -h "*|*" --help "* ) command kubectl "$@"; return ;;
  esac

  if [ "${KUSH_ACTIVE:-}" != "1" ]; then
    echo "kubectl blocked: enter a context first with 'kush <ctx>', or use 'kush exec'" >&2
    return 1
  fi
  command kubectl "$@"
}
```

:::

## `kush exec` is never blocked

`kush exec prod -- kubectl get pods` runs kubectl with `KUSH_ACTIVE=1` in its environment, and kush invokes the command directly rather than through your interactive shell. So the wrapper never sees it, and even if it did, the check passes. Scripts, CI, and agents that go through `kush exec` keep working with the guard installed.

## Choosing the safe-list

The safe-list is a judgment call, not a fixed rule: anything you allow runs against whatever ambient context is active when you're outside kush. `config`, `completion`, `version`, and `options` never contact a cluster, so they're uncontroversial. Read-only commands that _do_ reach the API server, like `cluster-info` and `api-resources`, are a gray area; add them only if you're fine with harmless reads hitting your default context. Everything that mutates or targets workloads (`apply`, `delete`, `exec`, `scale`, `rollout`, and the everyday `get`/`describe`/`logs`) should stay off the list. Blocking those is the whole point.

This wrapper is opt-in and complements kush rather than replacing it. kush guarantees isolation once you're in a shell; the guard makes sure you're in one before kubectl touches anything.
