// Package media provides local media conversion helpers.
package media

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/bartrosa/homelab-cli/internal/executil"
)

// HEICOptions configures HEIC → JPEG conversion.
type HEICOptions struct {
	Dir     string // default: current directory
	Quality int    // 1–100, passed to heif-convert -q
	Force   bool   // reconvert even when .jpg exists
	DryRun  bool
}

// ConvertHEIC converts .HEIC/.heic files in Dir to .jpg via heif-convert.
func ConvertHEIC(ctx context.Context, stdout, stderr io.Writer, opt HEICOptions) (int, error) {
	dir := strings.TrimSpace(opt.Dir)
	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return 0, err
		}
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return 0, err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return 0, fmt.Errorf("directory %s: %w", abs, err)
	}
	if !info.IsDir() {
		return 0, fmt.Errorf("%s is not a directory", abs)
	}

	if opt.Quality <= 0 || opt.Quality > 100 {
		opt.Quality = 100
	}

	entries, err := os.ReadDir(abs)
	if err != nil {
		return 0, err
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if ext != ".heic" {
			continue
		}
		files = append(files, filepath.Join(abs, name))
	}

	if len(files) == 0 {
		return 0, fmt.Errorf("no .HEIC/.heic files in %s", abs)
	}

	heif, err := executil.LookPath("heif-convert")
	if err != nil && !opt.DryRun {
		return 0, fmt.Errorf("heif-convert not found (macOS: brew install libheif; Ubuntu: apt install libheif-examples): %w", err)
	}
	if heif == "" {
		heif = "heif-convert"
	}

	run := executil.NewRunner(stdout, stderr)
	run.DryRun = opt.DryRun
	n := 0
	for _, src := range files {
		out := src + ".jpg"
		if !opt.Force {
			if _, err := os.Stat(out); err == nil {
				_, _ = fmt.Fprintf(stderr, "skip (exists): %s\n", filepath.Base(out))
				continue
			}
		}
		if opt.DryRun {
			_, _ = fmt.Fprintf(stderr, "[dry-run] %s -q %d %s %s\n", heif, opt.Quality, filepath.Base(src), filepath.Base(out))
			n++
			continue
		}
		if err := run.Run(ctx, heif, "-q", fmt.Sprintf("%d", opt.Quality), src, out); err != nil {
			return n, fmt.Errorf("%s: %w", filepath.Base(src), err)
		}
		_, _ = fmt.Fprintf(stdout, "%s → %s\n", filepath.Base(src), filepath.Base(out))
		n++
	}
	return n, nil
}
