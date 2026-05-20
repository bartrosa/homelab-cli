// Package homelabroot resolves the personal homelab repository path for script delegation.
package homelabroot

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const envVar = "LAB_HOMELAB_ROOT"

// Resolve returns an absolute path to the homelab repo (config override, env, or common locations).
func Resolve(configRoot string) (string, error) {
	var candidates []string
	if strings.TrimSpace(configRoot) != "" {
		candidates = append(candidates, configRoot)
	} else {
		if v := os.Getenv(envVar); v != "" {
			candidates = append(candidates, v)
		}
		home, err := os.UserHomeDir()
		if err == nil {
			candidates = append(candidates,
				filepath.Join(home, "Projects", "PERSONAL", "homelab"),
				filepath.Join(home, "homelab"),
			)
		}
		cwd, _ := os.Getwd()
		if cwd != "" {
			candidates = append(candidates, cwd, filepath.Join(cwd, ".."))
		}
	}

	seen := map[string]struct{}{}
	for _, c := range candidates {
		if c == "" {
			continue
		}
		abs, err := filepath.Abs(c)
		if err != nil {
			continue
		}
		if _, ok := seen[abs]; ok {
			continue
		}
		seen[abs] = struct{}{}
		if isHomelabRepo(abs) {
			return abs, nil
		}
	}
	return "", fmt.Errorf("homelab repo not found (set %s or homelab.root in config)", envVar)
}

func isHomelabRepo(dir string) bool {
	for _, marker := range []string{
		filepath.Join(dir, "ml-stack", "podman-compose.yml"),
		filepath.Join(dir, "project-initiators", "golang"),
		filepath.Join(dir, "postgres", "config", "instances.yaml"),
	} {
		if _, err := os.Stat(marker); err == nil {
			return true
		}
	}
	return false
}
