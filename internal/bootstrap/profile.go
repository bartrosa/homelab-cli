// Package bootstrap runs machine setup profiles (packages, toolchains, scripts).
package bootstrap

import (
	"embed"
	"fmt"
	"io"
	"io/fs"

	"gopkg.in/yaml.v3"
)

//go:embed profiles/*.yaml
var embeddedProfiles embed.FS

// StepType identifies a bootstrap step kind.
type StepType string

const (
	// StepPkg installs distribution packages.
	StepPkg StepType = "pkg"
	// StepToolchain installs mise runtimes.
	StepToolchain StepType = "toolchain"
	// StepScript runs a homelab shell script.
	StepScript StepType = "script"
)

// Step is one idempotent action in a profile.
type Step struct {
	Type      StepType `yaml:"type"`
	Packages  []string `yaml:"packages,omitempty"`
	Languages []string `yaml:"languages,omitempty"`
	Script    string   `yaml:"script,omitempty"` // path relative to homelab root
}

// Profile describes a bootstrap target.
type Profile struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Steps       []Step `yaml:"steps"`
}

// LoadEmbedded returns a built-in profile by name (e.g. laptop-macos).
func LoadEmbedded(name string) (*Profile, error) {
	path := fmt.Sprintf("profiles/%s.yaml", name)
	data, err := embeddedProfiles.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("unknown built-in profile %q: %w", name, err)
	}
	var p Profile
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("parse profile %s: %w", name, err)
	}
	if p.Name == "" {
		p.Name = name
	}
	return &p, nil
}

// ListEmbedded returns names of built-in profiles.
func ListEmbedded() ([]string, error) {
	entries, err := fs.ReadDir(embeddedProfiles, "profiles")
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		n := e.Name()
		if len(n) > 5 && n[len(n)-5:] == ".yaml" {
			names = append(names, n[:len(n)-5])
		}
	}
	return names, nil
}

// LoadFromYAML parses profile bytes (user config override).
func LoadFromYAML(data []byte) (*Profile, error) {
	var p Profile
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	if p.Name == "" {
		return nil, fmt.Errorf("profile missing name")
	}
	if len(p.Steps) == 0 {
		return nil, fmt.Errorf("profile %q has no steps", p.Name)
	}
	return &p, nil
}

// DecodeProfileMap unmarshals a config profiles map entry.
func DecodeProfileMap(raw any) (*Profile, error) {
	data, err := yaml.Marshal(raw)
	if err != nil {
		return nil, err
	}
	return LoadFromYAML(data)
}

// WriteProfileList prints profile summaries to w.
func WriteProfileList(w io.Writer, profiles []ProfileSummary) {
	for _, p := range profiles {
		_, _ = fmt.Fprintf(w, "  %s — %s\n", p.Name, p.Description)
	}
}

// ProfileSummary is a short profile listing entry.
type ProfileSummary struct {
	Name        string
	Description string
	Source      string // "builtin" or "config"
}
