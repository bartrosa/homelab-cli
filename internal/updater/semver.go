package updater

import (
	"strings"
)

// CompareVersions returns -1 if a < b, 0 if equal, 1 if a > b.
// Handles optional "v" prefix and "dev" as older than any release.
func CompareVersions(a, b string) int {
	va := normalizeVersion(a)
	vb := normalizeVersion(b)
	if va == "dev" && vb != "dev" {
		return -1
	}
	if vb == "dev" && va != "dev" {
		return 1
	}
	if va == vb {
		return 0
	}
	ap := parseParts(va)
	bp := parseParts(vb)
	for i := 0; i < 3; i++ {
		if ap[i] < bp[i] {
			return -1
		}
		if ap[i] > bp[i] {
			return 1
		}
	}
	// Compare pre-release suffix lexically (rc < final)
	as := prereleaseSuffix(va)
	bs := prereleaseSuffix(vb)
	if as == bs {
		return 0
	}
	if as == "" {
		return 1
	}
	if bs == "" {
		return -1
	}
	if as < bs {
		return -1
	}
	if as > bs {
		return 1
	}
	return 0
}

func normalizeVersion(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "v")
	if v == "" || v == "dev" {
		return "dev"
	}
	return v
}

func parseParts(v string) [3]int {
	core := v
	if idx := strings.IndexByte(v, '-'); idx >= 0 {
		core = v[:idx]
	}
	parts := strings.Split(core, ".")
	var out [3]int
	for i := 0; i < len(parts) && i < 3; i++ {
		out[i] = atoi(parts[i])
	}
	return out
}

func prereleaseSuffix(v string) string {
	if idx := strings.IndexByte(v, '-'); idx >= 0 {
		return v[idx+1:]
	}
	return ""
}

func atoi(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			break
		}
		n = n*10 + int(c-'0')
	}
	return n
}

// IsNewer reports whether remote is newer than current.
func IsNewer(current, remote string) bool {
	return CompareVersions(current, remote) < 0
}
