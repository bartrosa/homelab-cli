// Package toolchain wraps mise for language runtime installation.
package toolchain

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/bartrosa/homelab-cli/internal/executil"
	"github.com/bartrosa/homelab-cli/internal/platform"
)

// Known language aliases for mise.
var knownLangs = map[string]string{
	"go": "go", "golang": "go",
	"node": "node", "nodejs": "node", "typescript": "node",
	"bun": "bun", "deno": "deno",
	"python": "python", "py": "python",
	"rust":   "rust",
	"ruby":   "ruby",
	"java":   "java",
	"zig":    "zig",
	"erlang": "erlang", "elixir": "elixir",
	"lua": "lua",
}

// Runner manages toolchains via mise.
type Runner struct {
	Runner *executil.Runner
	Info   platform.Info
}

// New creates a mise runner.
func New(stdout, stderr io.Writer, dryRun bool) *Runner {
	r := executil.NewRunner(stdout, stderr)
	r.DryRun = dryRun
	return &Runner{Runner: r, Info: platform.Detect()}
}

// Install installs one or more languages via mise.
func (r *Runner) Install(ctx context.Context, langs ...string) error {
	if !r.Info.SupportsMise() {
		return fmt.Errorf("mise toolchains not supported on %s", r.Info.GOOS)
	}
	if err := r.ensureMise(ctx); err != nil {
		return err
	}
	for _, lang := range langs {
		canonical, ok := knownLangs[strings.ToLower(strings.TrimSpace(lang))]
		if !ok {
			return fmt.Errorf("unknown language %q (supported: go, node, python, rust, bun, deno, zig, java, ruby, erlang, elixir, lua)", lang)
		}
		if err := r.Runner.Run(ctx, "mise", "use", "-g", canonical+"@latest"); err != nil {
			return fmt.Errorf("mise install %s: %w", canonical, err)
		}
	}
	return nil
}

// List prints installed toolchains.
func (r *Runner) List(ctx context.Context) error {
	if err := r.ensureMise(ctx); err != nil {
		return err
	}
	return r.Runner.Run(ctx, "mise", "list")
}

// Use activates a specific version globally.
func (r *Runner) Use(ctx context.Context, lang, version string) error {
	if err := r.ensureMise(ctx); err != nil {
		return err
	}
	canonical, ok := knownLangs[strings.ToLower(strings.TrimSpace(lang))]
	if !ok {
		return fmt.Errorf("unknown language %q", lang)
	}
	spec := canonical + "@" + strings.TrimSpace(version)
	return r.Runner.Run(ctx, "mise", "use", "-g", spec)
}

func (r *Runner) ensureMise(ctx context.Context) error {
	if executil.CommandExists("mise") {
		return nil
	}
	// Try installing mise via curl installer (matches upstream docs).
	if r.Runner.DryRun {
		return r.Runner.Run(ctx, "sh", "-c", "curl https://mise.jdx.dev/install.sh | sh")
	}
	if !executil.CommandExists("curl") {
		return fmt.Errorf("mise not found and curl missing; install mise: https://mise.jdx.dev")
	}
	return r.Runner.Run(ctx, "sh", "-c", "curl https://mise.jdx.dev/install.sh | sh")
}

// SupportedLanguages returns canonical language keys.
func SupportedLanguages() []string {
	seen := make(map[string]struct{})
	var out []string
	for _, v := range knownLangs {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}
