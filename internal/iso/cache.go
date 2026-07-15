package iso

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// CachedImage is an ISO file in the local cache directory.
type CachedImage struct {
	Path      string
	Name      string
	Size      string
	SizeBytes int64
}

// DefaultCacheDir returns ~/.cache/homelab-cli/iso.
func DefaultCacheDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".cache", "homelab-cli", "iso"), nil
}

// ListCachedImages scans dir for .iso files (newest first).
func ListCachedImages(dir string) ([]CachedImage, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []CachedImage
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".iso") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		st, err := os.Stat(path)
		if err != nil || st.Size() == 0 {
			continue
		}
		out = append(out, CachedImage{
			Path:      path,
			Name:      e.Name(),
			Size:      formatBytes(st.Size()),
			SizeBytes: st.Size(),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// ResolveISORef resolves a distro id, filename fragment, or path to a cached/existing ISO.
func ResolveISORef(ref, cacheDir string) (string, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", fmt.Errorf("empty ISO reference")
	}

	if strings.Contains(ref, string(os.PathSeparator)) || strings.HasPrefix(ref, "~") {
		p, err := expandHome(ref)
		if err != nil {
			return "", err
		}
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p, nil
		}
	}

	if d, ok := LookupDistro(ref); ok {
		arch := "amd64"
		if len(d.Architectures) > 0 {
			arch = d.Architectures[0]
		}
		rel, err := d.Resolve(arch, "")
		if err == nil {
			candidate := filepath.Join(cacheDir, rel.ISOFilename)
			if st, err := os.Stat(candidate); err == nil && st.Size() > 0 {
				return candidate, nil
			}
			return "", fmt.Errorf("cached ISO for %q not found at %s (run: lab iso download %s)", ref, candidate, ref)
		}
	}

	images, err := ListCachedImages(cacheDir)
	if err != nil {
		return "", err
	}
	refLower := strings.ToLower(ref)
	var matches []CachedImage
	for _, img := range images {
		if strings.EqualFold(img.Name, ref) || strings.EqualFold(img.Name, ref+".iso") {
			return img.Path, nil
		}
		if strings.Contains(strings.ToLower(img.Name), refLower) {
			matches = append(matches, img)
		}
	}
	switch len(matches) {
	case 1:
		return matches[0].Path, nil
	case 0:
		return "", fmt.Errorf("no cached ISO matching %q (run: lab iso download …)", ref)
	default:
		names := make([]string, len(matches))
		for i, m := range matches {
			names[i] = m.Name
		}
		return "", fmt.Errorf("ambiguous ISO reference %q matches: %s", ref, strings.Join(names, ", "))
	}
}

func expandHome(p string) (string, error) {
	if p == "~" {
		return os.UserHomeDir()
	}
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, p[2:]), nil
	}
	return p, nil
}
