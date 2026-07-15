package services

import "sort"

// DefaultPresets are built-in service bundles.
var DefaultPresets = map[string][]string{
	"observability":  {"prometheus", "loki", "tempo", "grafana"},
	"ml-stack":       {"postgres", "qdrant", "minio", "clickhouse"},
	"data-lakehouse": {"postgres", "clickhouse", "minio"},
	"microservices":  {"postgres", "redis", "rabbitmq"},
	"vector-search":  {"qdrant", "weaviate"},
	"full-obs":       {"prometheus", "grafana", "loki", "tempo", "minio"},
	"graphrag":       {"arcadedb", "qdrant", "minio", "postgres"},
	"graph-lab":      {"arcadedb", "nebulagraph"},
}

// PresetNames returns sorted preset keys.
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
		out[k] = append([]string(nil), v...)
	}
	for k, v := range custom {
		out[k] = append([]string(nil), v...)
	}
	return out
}

// ResolvePreset returns service ids for a preset name.
func ResolvePreset(name string, custom map[string][]string) ([]string, error) {
	merged := MergePresets(custom)
	ids, ok := merged[name]
	if !ok {
		return nil, &PresetError{Name: name}
	}
	return append([]string(nil), ids...), nil
}

// PresetError indicates an unknown preset.
type PresetError struct{ Name string }

func (e *PresetError) Error() string { return "unknown service preset " + e.Name }

// ExpandNames resolves preset names to service ids; passthrough for raw ids.
func ExpandNames(names []string, custom map[string][]string) ([]string, error) {
	var out []string
	seen := map[string]struct{}{}
	for _, name := range names {
		name = trim(name)
		if name == "" {
			continue
		}
		if ids, err := ResolvePreset(name, custom); err == nil {
			for _, id := range ids {
				if _, ok := seen[id]; ok {
					continue
				}
				seen[id] = struct{}{}
				out = append(out, id)
			}
			continue
		}
		if _, ok := Lookup(name); !ok {
			return nil, fmtUnknownService(name)
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	return out, nil
}

func trim(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t') {
		s = s[:len(s)-1]
	}
	return s
}

func fmtUnknownService(name string) error {
	return &UnknownServiceError{Name: name}
}

// UnknownServiceError indicates an unknown service id.
type UnknownServiceError struct{ Name string }

func (e *UnknownServiceError) Error() string { return "unknown service " + e.Name }
