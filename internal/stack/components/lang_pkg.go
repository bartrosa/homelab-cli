package components

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/bartrosa/homelab-cli/internal/stack"
)

type rustComponent struct{}

func (r *rustComponent) ID() string               { return "rust" }
func (r *rustComponent) DisplayName() string      { return "Rust" }
func (r *rustComponent) Category() stack.Category { return stack.CategoryLanguage }
func (r *rustComponent) Description() string {
	return "Rust via rustup (stable + clippy/rustfmt/rust-analyzer)"
}
func (r *rustComponent) DefaultVersion() string { return "stable" }
func (r *rustComponent) Requires() []string     { return nil }

func (r *rustComponent) PathEntries() []stack.PathEntry {
	return []stack.PathEntry{{Shell: "all", Marker: "rust", Content: `[ -f "$HOME/.cargo/env" ] && . "$HOME/.cargo/env"`}}
}

func (r *rustComponent) IsInstalled(ctx context.Context, env *stack.Env) (bool, string, error) {
	if !cmdExists(ctx, env, "rustup") {
		return false, "", nil
	}
	ver := versionOf(ctx, env, "rustc", "--version")
	return ver != "", ver, nil
}

func (r *rustComponent) Install(ctx context.Context, env *stack.Env, opts stack.InstallOptions) error {
	toolchain := "stable"
	if opts.Extra != nil && opts.Extra["rust-toolchain"] != "" {
		toolchain = opts.Extra["rust-toolchain"]
	}
	script := fmt.Sprintf(`curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y --default-toolchain %s --profile default --component clippy --component rustfmt --component rust-analyzer`, toolchain)
	if opts.DryRun {
		return env.Runner.Run(ctx, "sh", "-c", script)
	}
	return env.Runner.Run(ctx, "sh", "-c", script)
}

type scalaComponent struct{}

func (s *scalaComponent) ID() string               { return "scala" }
func (s *scalaComponent) DisplayName() string      { return "Scala" }
func (s *scalaComponent) Category() stack.Category { return stack.CategoryLanguage }
func (s *scalaComponent) Description() string {
	return "Scala via Coursier (cs setup)"
}
func (s *scalaComponent) DefaultVersion() string { return "latest" }
func (s *scalaComponent) Requires() []string     { return []string{"java"} }

func (s *scalaComponent) PathEntries() []stack.PathEntry {
	return []stack.PathEntry{
		{Shell: "all", Marker: "user-local-bin", Content: `export PATH="$HOME/.local/bin:$PATH"`},
		{Shell: "all", Marker: "coursier", Content: `export PATH="$HOME/.local/share/coursier/bin:$PATH"`},
	}
}

func (s *scalaComponent) IsInstalled(ctx context.Context, env *stack.Env) (bool, string, error) {
	if !cmdExists(ctx, env, "cs") || !cmdExists(ctx, env, "scala") {
		return false, "", nil
	}
	return true, versionOf(ctx, env, "scala", "-version"), nil
}

func (s *scalaComponent) Install(ctx context.Context, env *stack.Env, opts stack.InstallOptions) error {
	home, _ := os.UserHomeDir()
	bin := filepath.Join(home, ".local", "bin")
	_ = os.MkdirAll(bin, 0o750)
	arch := coursierArch()
	url := fmt.Sprintf("https://github.com/coursier/coursier/releases/latest/download/cs-%s.gz", arch)
	script := fmt.Sprintf(`curl -fL %q | gzip -d > %q/cs && chmod +x %q/cs && %q/cs setup --yes`, url, bin, bin, bin)
	if opts.DryRun {
		return env.Runner.Run(ctx, "sh", "-c", script)
	}
	return env.Runner.Run(ctx, "sh", "-c", script)
}

func coursierArch() string {
	switch runtime.GOOS {
	case "darwin":
		if runtime.GOARCH == "arm64" {
			return "aarch64-apple-darwin"
		}
		return "x86_64-apple-darwin"
	default:
		if runtime.GOARCH == "arm64" {
			return "aarch64-pc-linux"
		}
		return "x86_64-pc-linux"
	}
}

type uvComponent struct{}

func (u *uvComponent) ID() string               { return "uv" }
func (u *uvComponent) DisplayName() string      { return "uv" }
func (u *uvComponent) Category() stack.Category { return stack.CategoryPackageMgr }
func (u *uvComponent) Description() string      { return "Python package manager (uv)" }
func (u *uvComponent) DefaultVersion() string   { return "latest" }
func (u *uvComponent) Requires() []string       { return nil }

func (u *uvComponent) PathEntries() []stack.PathEntry {
	return []stack.PathEntry{{Shell: "all", Marker: "user-local-bin", Content: `export PATH="$HOME/.local/bin:$PATH"`}}
}

func (u *uvComponent) IsInstalled(ctx context.Context, env *stack.Env) (bool, string, error) {
	ok := cmdExists(ctx, env, "uv")
	return ok, versionOf(ctx, env, "uv", "--version"), nil
}

func (u *uvComponent) Install(ctx context.Context, env *stack.Env, opts stack.InstallOptions) error {
	script := "curl -LsSf https://astral.sh/uv/install.sh | sh"
	if opts.DryRun {
		return env.Runner.Run(ctx, "sh", "-c", script)
	}
	return env.Runner.Run(ctx, "sh", "-c", script)
}

type yarnComponent struct{}

func (y *yarnComponent) ID() string               { return "yarn" }
func (y *yarnComponent) DisplayName() string      { return "Yarn" }
func (y *yarnComponent) Category() stack.Category { return stack.CategoryPackageMgr }
func (y *yarnComponent) Description() string      { return "Yarn via corepack" }
func (y *yarnComponent) DefaultVersion() string   { return "stable" }
func (y *yarnComponent) Requires() []string       { return []string{"node"} }

func (y *yarnComponent) PathEntries() []stack.PathEntry { return nil }

func (y *yarnComponent) IsInstalled(ctx context.Context, env *stack.Env) (bool, string, error) {
	ok := cmdExists(ctx, env, "yarn")
	return ok, versionOf(ctx, env, "yarn", "--version"), nil
}

func (y *yarnComponent) Install(ctx context.Context, env *stack.Env, _ stack.InstallOptions) error {
	return env.Runner.Run(ctx, "sh", "-c", "corepack enable && corepack prepare yarn@stable --activate")
}

type pnpmComponent struct{}

func (p *pnpmComponent) ID() string               { return "pnpm" }
func (p *pnpmComponent) DisplayName() string      { return "pnpm" }
func (p *pnpmComponent) Category() stack.Category { return stack.CategoryPackageMgr }
func (p *pnpmComponent) Description() string      { return "pnpm via corepack" }
func (p *pnpmComponent) DefaultVersion() string   { return "latest" }
func (p *pnpmComponent) Requires() []string       { return []string{"node"} }

func (p *pnpmComponent) PathEntries() []stack.PathEntry { return nil }

func (p *pnpmComponent) IsInstalled(ctx context.Context, env *stack.Env) (bool, string, error) {
	ok := cmdExists(ctx, env, "pnpm")
	return ok, versionOf(ctx, env, "pnpm", "--version"), nil
}

func (p *pnpmComponent) Install(ctx context.Context, env *stack.Env, _ stack.InstallOptions) error {
	return env.Runner.Run(ctx, "sh", "-c", "corepack enable && corepack prepare pnpm@latest --activate")
}

func registerLangAndPkg() {
	stack.Register(&rustComponent{})
	stack.Register(&scalaComponent{})
	stack.Register(&uvComponent{})
	stack.Register(&yarnComponent{})
	stack.Register(&pnpmComponent{})
}

func init() { registerMiseComponents(); registerLangAndPkg() }
