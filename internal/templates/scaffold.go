// Package templates scaffolds new projects from homelab project-initiators.
package templates

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/bartrosa/homelab-cli/internal/homelabroot"
)

// Known template kinds.
var kinds = map[string]string{
	"go":         "golang",
	"golang":     "golang",
	"python":     "python",
	"rust":       "rust",
	"typescript": "typescript",
	"ts":         "typescript",
}

// NewProject copies a project-initiator template into destDir.
func NewProject(homelabRoot, kind, destDir string, stdout io.Writer) error {
	root, err := homelabroot.Resolve(homelabRoot)
	if err != nil {
		return err
	}
	canonical, ok := kinds[strings.ToLower(strings.TrimSpace(kind))]
	if !ok {
		return fmt.Errorf("unknown template %q (known: go, python, rust, typescript)", kind)
	}
	src := filepath.Join(root, "project-initiators", canonical)
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("template source %s: %w", src, err)
	}
	destDir = strings.TrimSpace(destDir)
	if destDir == "" {
		return fmt.Errorf("destination directory required")
	}
	if err := os.MkdirAll(destDir, 0o750); err != nil {
		return err
	}
	return copyDir(src, destDir, stdout)
}

func copyDir(src, dest string, stdout io.Writer) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		target := filepath.Join(dest, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o750)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o750); err != nil {
			return err
		}
		if err := os.WriteFile(target, data, 0o644); err != nil {
			return err
		}
		if stdout != nil {
			_, _ = fmt.Fprintf(stdout, "  %s\n", rel)
		}
		return nil
	})
}

// ListKinds returns supported template names.
func ListKinds() []string {
	return []string{"go", "python", "rust", "typescript"}
}
