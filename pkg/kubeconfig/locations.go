package kubeconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	humane "github.com/sierrasoftworks/humane-errors-go"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// Warning is a non-fatal diagnostic from resolving/merging kubeconfig locations.
type Warning struct{ Message string }

// ExpandLocations expands each entry — env vars ($VAR/${VAR}, with $KUBECONFIG
// and any other value split on os.PathListSeparator), a leading ~ for the home
// dir, and shell globs — into an ordered, de-duplicated list of existing regular
// files. Unset env, non-matching globs, missing paths and directories all
// contribute nothing (never an error).
func ExpandLocations(locations []string) ([]string, error) {
	var out []string
	seen := make(map[string]bool)
	add := func(p string) {
		if p == "" || seen[p] {
			return
		}
		info, err := os.Stat(p)
		if err != nil || info.IsDir() {
			return
		}
		seen[p] = true
		out = append(out, p)
	}

	for _, loc := range locations {
		expanded := os.Expand(loc, os.Getenv)
		for _, entry := range filepath.SplitList(expanded) {
			entry = expandTilde(strings.TrimSpace(entry))
			if entry == "" {
				continue
			}
			matches, err := filepath.Glob(entry)
			if err != nil {
				return nil, humane.Wrap(err, "invalid path pattern "+entry, "check your context_lookup_locations entry")
			}
			for _, m := range matches {
				add(m)
			}
		}
	}
	return out, nil
}

func expandTilde(p string) string {
	if p != "~" && !strings.HasPrefix(p, "~/") {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return p
	}
	if p == "~" {
		return home
	}
	return filepath.Join(home, p[2:])
}

// LoadResolved merges kubeconfigs from the given locations. nil/empty locations
// — or configured locations that match no files — fall back to clientcmd's
// default loading rules. Otherwise the expanded files become the merge
// precedence (first-wins); duplicate context names across files yield warnings
// naming the winning file.
func LoadResolved(locations []string) (*api.Config, []Warning, error) {
	if len(locations) == 0 {
		cfg, err := Load()
		return cfg, nil, err
	}
	files, err := ExpandLocations(locations)
	if err != nil {
		return nil, nil, err
	}
	if len(files) == 0 {
		cfg, loadErr := Load()
		return cfg, nil, loadErr
	}
	// Parse each candidate once. Skip (with a warning) any file that isn't a
	// valid kubeconfig, so globbing over a directory of mixed files — old
	// formats, notes, empty files — doesn't break the whole load.
	valid := make([]string, 0, len(files))
	parsed := make(map[string]*api.Config, len(files))
	var warnings []Warning
	for _, f := range files {
		c, parseErr := clientcmd.LoadFromFile(f)
		if parseErr != nil {
			warnings = append(warnings, Warning{Message: fmt.Sprintf("skipping %s: not a valid kubeconfig (%v)", f, parseErr)})
			continue
		}
		valid = append(valid, f)
		parsed[f] = c
	}
	if len(valid) == 0 {
		cfg, loadErr := Load()
		return cfg, warnings, loadErr
	}

	cfg, err := LoadFrom(&clientcmd.ClientConfigLoadingRules{Precedence: valid})
	if err != nil {
		return nil, warnings, err
	}
	return cfg, append(warnings, duplicateWarnings(valid, parsed)...), nil
}

// duplicateWarnings reports context names present in more than one file, using
// the already-parsed configs. The winner is the first file in list order
// (matching clientcmd's first-wins merge). Names are sorted for determinism.
func duplicateWarnings(files []string, parsed map[string]*api.Config) []Warning {
	count := make(map[string]int)
	firstFile := make(map[string]string)
	for _, f := range files {
		c := parsed[f]
		if c == nil {
			continue
		}
		for name := range c.Contexts {
			count[name]++
			if _, ok := firstFile[name]; !ok {
				firstFile[name] = f
			}
		}
	}
	var names []string
	for n, c := range count {
		if c > 1 {
			names = append(names, n)
		}
	}
	sort.Strings(names)
	ws := make([]Warning, 0, len(names))
	for _, n := range names {
		ws = append(ws, Warning{Message: fmt.Sprintf("context %q defined in %d files; using %s", n, count[n], firstFile[n])})
	}
	return ws
}
