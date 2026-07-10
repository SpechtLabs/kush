//go:build unix

package tempkube

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestSweepStale(t *testing.T) {
	dir := t.TempDir()

	// A file owned by a dead PID (PID 2^30 is not running).
	deadPID := 1 << 30
	dead := filepath.Join(dir, "prod-"+strconv.Itoa(deadPID)+"-aaaa.yaml")
	if err := os.WriteFile(dead, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	// A file owned by us (alive).
	alive := filepath.Join(dir, "dev-"+strconv.Itoa(os.Getpid())+"-bbbb.yaml")
	if err := os.WriteFile(alive, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	// A non-kush file: left untouched.
	other := filepath.Join(dir, "notes.txt")
	if err := os.WriteFile(other, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	SweepStale(dir)

	if _, err := os.Stat(dead); !os.IsNotExist(err) {
		t.Fatal("dead-PID temp file was not swept")
	}
	if _, err := os.Stat(alive); err != nil {
		t.Fatal("alive-PID temp file was wrongly swept")
	}
	if _, err := os.Stat(other); err != nil {
		t.Fatal("non-kush file was wrongly swept")
	}
}
