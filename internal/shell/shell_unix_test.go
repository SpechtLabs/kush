//go:build unix

package shell

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestRunPostExecHooksShareEnvironmentWithSubshell(t *testing.T) {
	dir := t.TempDir()
	shellPath := filepath.Join(dir, "test-shell")
	outPath := filepath.Join(dir, "environment")

	// The first invocation handles -c like a shell. The exec at the end of
	// Run's hook script invokes this helper again without arguments, where it
	// records the environment inherited by the final subshell.
	helper := fmt.Sprintf(`#!/bin/sh
if [ "$1" = "-c" ]; then
  exec /bin/sh "$@"
fi
printf '%%s|%%s|%%s' "$POST_EXEC_VALUE" "$KUBECONFIG" "$KUSH_CONTEXT" > %q
`, outPath)
	if err := os.WriteFile(shellPath, []byte(helper), 0o700); err != nil {
		t.Fatal(err)
	}

	err := Run(context.Background(), shellPath, "/tmp/kubeconfig", []string{"KUSH_CONTEXT=prod"}, []string{
		"export POST_EXEC_VALUE=first",
		"export POST_EXEC_VALUE=$POST_EXEC_VALUE-second",
	})
	if err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "first-second|/tmp/kubeconfig|prod" {
		t.Fatalf("subshell environment = %q, want hook and kush environment", got)
	}
}
