//go:build unix

package tempkube

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

// SweepStale removes temp kubeconfigs in dir whose owning PID (encoded in the
// filename as <ctx>-<pid>-<rand>.yaml) is no longer alive. Best effort: errors
// are ignored so a normal invocation is never blocked by cleanup.
func SweepStale(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		pid, ok := pidFromName(e.Name())
		if !ok {
			continue
		}
		if !processAlive(pid) {
			_ = os.Remove(filepath.Join(dir, e.Name()))
		}
	}
}

// pidFromName parses the PID out of <ctx>-<pid>-<rand>.yaml. The PID is the
// second-to-last '-'-separated field of the base name (rand is last).
func pidFromName(name string) (int, bool) {
	base := strings.TrimSuffix(name, ".yaml")
	if base == name {
		return 0, false // not a .yaml file
	}
	parts := strings.Split(base, "-")
	if len(parts) < 3 {
		return 0, false
	}
	pid, err := strconv.Atoi(parts[len(parts)-2])
	if err != nil || pid <= 0 {
		return 0, false
	}
	return pid, true
}

// processAlive reports whether a process with the given PID exists. Signal 0
// performs error checking without delivering a signal; EPERM means the process
// exists but is owned by another user (still alive).
func processAlive(pid int) bool {
	err := syscall.Kill(pid, 0)
	return err == nil || err == syscall.EPERM
}
