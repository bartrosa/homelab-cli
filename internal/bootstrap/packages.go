package bootstrap

import (
	"strings"
)

// packageMap maps generic package names to manager-specific names.
var packageMap = map[string]map[string]string{
	"git":             {"apt": "git", "rpm-ostree": "git"},
	"curl":            {"apt": "curl", "rpm-ostree": "curl"},
	"wget":            {"apt": "wget", "rpm-ostree": "wget"},
	"ca-certificates": {"apt": "ca-certificates", "rpm-ostree": "ca-certificates"},
	"gnupg":           {"apt": "gnupg", "rpm-ostree": "gnupg"},
	"tmux":            {"apt": "tmux", "rpm-ostree": "tmux"},
	"htop":            {"apt": "htop", "rpm-ostree": "htop"},
	"jq":              {"apt": "jq", "rpm-ostree": "jq"},
	"unzip":           {"apt": "unzip", "rpm-ostree": "unzip"},
	"ripgrep":         {"apt": "ripgrep", "rpm-ostree": "ripgrep"},
	"fd":              {"apt": "fd-find", "rpm-ostree": "fd-find"},
	"fzf":             {"apt": "fzf", "rpm-ostree": "fzf"},
	"bat":             {"apt": "bat", "rpm-ostree": "bat"},
	"yq":              {"apt": "yq", "rpm-ostree": "yq"},
	"zsh":             {"apt": "zsh", "rpm-ostree": "zsh"},
	"build":           {"apt": "build-essential", "rpm-ostree": "@development-tools"},
	"podman-compose":  {"apt": "podman-compose", "rpm-ostree": "podman-compose"},
	"distrobox":       {"apt": "distrobox", "rpm-ostree": "distrobox"},
	"flatpak":         {"apt": "flatpak", "rpm-ostree": "flatpak"},
}

// PackageFor returns the native package name for a generic name and manager.
func PackageFor(genericName, manager string) (string, bool) {
	manager = strings.ToLower(manager)
	if m, ok := packageMap[genericName]; ok {
		if pkg, ok := m[manager]; ok {
			return pkg, true
		}
	}
	return "", false
}

// CLIBasicPackages returns cross-target CLI packages.
func CLIBasicPackages() []string {
	return []string{"git", "curl", "wget", "ca-certificates", "gnupg", "tmux", "htop", "jq", "unzip"}
}

// ShellToolPackages returns shell/search utilities.
func ShellToolPackages() []string {
	return []string{"ripgrep", "fd", "fzf", "bat", "yq", "zsh"}
}

// AllEssentialSections returns default section order.
func AllEssentialSections() []string {
	return []string{
		"system-update",
		"cli-basics",
		"shell-tools",
		"build",
		"container-runtime",
		"mise",
		"distrobox",
		"flatpak-flathub",
	}
}

// FilterSections applies --only and --skip lists.
func FilterSections(all, only, skip []string) []string {
	set := map[string]struct{}{}
	if len(only) > 0 {
		for _, s := range only {
			set[strings.TrimSpace(s)] = struct{}{}
		}
	} else {
		for _, s := range all {
			set[s] = struct{}{}
		}
	}
	for _, s := range skip {
		delete(set, strings.TrimSpace(s))
	}
	var out []string
	for _, s := range all {
		if _, ok := set[s]; ok {
			out = append(out, s)
		}
	}
	return out
}
