package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	humane "github.com/sierrasoftworks/humane-errors-go"
	"github.com/spechtlabs/kush/internal/config"
)

func runPreExecHook(ctx context.Context, ctxName, namespace string) error {
	hook := config.PreExecHook(ctxName)
	if hook == "" {
		return nil
	}

	shellPath := config.Shell()
	if shellPath == "" {
		shellPath = os.Getenv("SHELL")
	}
	if shellPath == "" {
		shellPath = "/bin/sh"
	}

	cmd := exec.CommandContext(ctx, shellPath, "-c", hook)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(),
		"KUSH_CONTEXT="+ctxName,
		"KUSH_NAMESPACE="+namespace,
	)

	if err := cmd.Run(); err != nil {
		return humane.Wrap(err, fmt.Sprintf("pre-exec hook failed for context %q", ctxName), "fix the configured pre_exec_hook or remove it from your kush config")
	}
	return nil
}
