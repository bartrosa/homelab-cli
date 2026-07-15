package stack

import (
	"sort"
)

// DefaultPresets are built-in stack install bundles.
var DefaultPresets = map[string][]string{
	"minimal":    {"git", "docker", "make"},
	"basic":      {"git", "docker", "make", "python", "node", "uv"},
	"backend":    {"git", "docker", "make", "python", "node", "uv", "go"},
	"frontend":   {"git", "docker", "make", "node", "bun", "yarn", "pnpm"},
	"systems":    {"git", "docker", "make", "rust", "zig", "cpp", "cmake"},
	"jvm":        {"git", "docker", "make", "java", "kotlin", "scala"},
	"ml":         {"git", "docker", "make", "python", "uv", "cmake"},
	"data":       {"git", "docker", "make", "python", "uv", "duckdb", "sqlite"},
	"gpu-nvidia": {"git", "docker", "make", "python", "uv", "cmake", "cuda"},
	"gpu-amd":    {"git", "docker", "make", "python", "uv", "cmake", "rocm"},
}

// PresetNames returns sorted preset keys merged with config overrides.
func PresetNames(custom map[string][]string) []string {
	merged := MergePresets(custom)
	names := make([]string, 0, len(merged))
	for k := range merged {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

// MergePresets overlays custom presets on defaults.
func MergePresets(custom map[string][]string) map[string][]string {
	out := make(map[string][]string, len(DefaultPresets)+len(custom))
	for k, v := range DefaultPresets {
		cp := append([]string(nil), v...)
		out[k] = cp
	}
	for k, v := range custom {
		out[k] = append([]string(nil), v...)
	}
	out["full"] = fullPreset()
	return out
}

// ResolvePreset returns component ids for a preset name.
func ResolvePreset(name string, custom map[string][]string) ([]string, error) {
	merged := MergePresets(custom)
	ids, ok := merged[name]
	if !ok {
		return nil, errUnknownPreset(name)
	}
	return append([]string(nil), ids...), nil
}

func fullPreset() []string {
	seen := map[string]struct{}{}
	var out []string
	for _, c := range All() {
		if c.Category() == CategoryGPU {
			continue
		}
		id := c.ID()
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}

func errUnknownPreset(name string) error {
	return &PresetError{Name: name}
}

// PresetError indicates an unknown preset.
type PresetError struct{ Name string }

func (e *PresetError) Error() string { return "unknown stack preset " + e.Name }
