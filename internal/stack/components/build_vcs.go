// Package components registers stack installable components.
package components

import (
	"context"
	"fmt"

	"github.com/bartrosa/homelab-cli/internal/stack"
)

type pkgComponent struct {
	id, displayName, description string
	category                     stack.Category
	ubuntuPkgs, silverPkgs       []string
	checkCmd                     string
}

func (p *pkgComponent) ID() string                     { return p.id }
func (p *pkgComponent) DisplayName() string            { return p.displayName }
func (p *pkgComponent) Category() stack.Category       { return p.category }
func (p *pkgComponent) Description() string            { return p.description }
func (p *pkgComponent) DefaultVersion() string         { return "system" }
func (p *pkgComponent) Requires() []string             { return nil }
func (p *pkgComponent) PathEntries() []stack.PathEntry { return nil }

func (p *pkgComponent) IsInstalled(ctx context.Context, env *stack.Env) (bool, string, error) {
	if p.checkCmd != "" {
		ok := cmdExists(ctx, env, p.checkCmd)
		return ok, versionOf(ctx, env, p.checkCmd, "--version"), nil
	}
	if len(p.ubuntuPkgs) > 0 {
		ok, err := pkgInstalled(ctx, env, p.ubuntuPkgs[0])
		return ok, "", err
	}
	return false, "", nil
}

func (p *pkgComponent) Install(ctx context.Context, env *stack.Env, _ stack.InstallOptions) error {
	pkgs := p.ubuntuPkgs
	if env.Info.IsSilverblue {
		pkgs = p.silverPkgs
		if len(pkgs) > 0 {
			fmt.Fprintln(env.Stderr, "! rpm-ostree changes may require a reboot")
		}
	}
	if len(pkgs) == 0 {
		return fmt.Errorf("%s: no packages for this OS", p.id)
	}
	return installPkg(ctx, env, pkgs...)
}

func registerBuildAndVCS() {
	stack.Register(&pkgComponent{
		id: "cpp", displayName: "C/C++ toolchain", category: stack.CategoryBuildTool,
		description: "gcc, g++, clang",
		ubuntuPkgs:  []string{"build-essential", "gcc", "g++", "clang", "clang-tools", "clangd", "libc++-dev", "libc++abi-dev"},
		silverPkgs:  []string{"gcc", "gcc-c++", "clang", "clang-tools-extra", "libcxx-devel", "libcxxabi-devel"},
		checkCmd:    "g++",
	})
	stack.Register(&pkgComponent{
		id: "cmake", displayName: "CMake + Ninja", category: stack.CategoryBuildTool,
		description: "cmake and ninja-build",
		ubuntuPkgs:  []string{"cmake", "ninja-build", "pkg-config"},
		silverPkgs:  []string{"cmake", "ninja-build", "pkgconf-pkg-config"},
		checkCmd:    "cmake",
	})
	stack.Register(&pkgComponent{
		id: "make", displayName: "GNU Make", category: stack.CategoryBuildTool,
		description: "make build tool",
		ubuntuPkgs:  []string{"make"},
		silverPkgs:  []string{"make"},
		checkCmd:    "make",
	})
	stack.Register(&gitComponent{})
	stack.Register(&sqliteComponent{})
}

type gitComponent struct{}

func (g *gitComponent) ID() string                     { return "git" }
func (g *gitComponent) DisplayName() string            { return "Git" }
func (g *gitComponent) Category() stack.Category       { return stack.CategoryVCS }
func (g *gitComponent) Description() string            { return "Git and git-lfs" }
func (g *gitComponent) DefaultVersion() string         { return "system" }
func (g *gitComponent) Requires() []string             { return nil }
func (g *gitComponent) PathEntries() []stack.PathEntry { return nil }

func (g *gitComponent) IsInstalled(ctx context.Context, env *stack.Env) (bool, string, error) {
	ok := cmdExists(ctx, env, "git")
	return ok, versionOf(ctx, env, "git", "--version"), nil
}

func (g *gitComponent) Install(ctx context.Context, env *stack.Env, _ stack.InstallOptions) error {
	if env.Info.IsSilverblue {
		ok, _ := pkgInstalled(ctx, env, "git")
		if ok {
			fmt.Fprintln(env.Stdout, "git already present on Silverblue")
			return nil
		}
		return installPkg(ctx, env, "git", "git-lfs")
	}
	return installPkg(ctx, env, "git", "git-lfs")
}

type sqliteComponent struct{}

func (s *sqliteComponent) ID() string                     { return "sqlite" }
func (s *sqliteComponent) DisplayName() string            { return "SQLite" }
func (s *sqliteComponent) Category() stack.Category       { return stack.CategoryDatabaseEmbedded }
func (s *sqliteComponent) Description() string            { return "SQLite CLI and dev headers" }
func (s *sqliteComponent) DefaultVersion() string         { return "system" }
func (s *sqliteComponent) Requires() []string             { return nil }
func (s *sqliteComponent) PathEntries() []stack.PathEntry { return nil }

func (s *sqliteComponent) IsInstalled(ctx context.Context, env *stack.Env) (bool, string, error) {
	ok := cmdExists(ctx, env, "sqlite3")
	return ok, versionOf(ctx, env, "sqlite3", "--version"), nil
}

func (s *sqliteComponent) Install(ctx context.Context, env *stack.Env, _ stack.InstallOptions) error {
	if env.Info.IsSilverblue {
		fmt.Fprintln(env.Stderr, "! rpm-ostree changes may require a reboot")
		return installPkg(ctx, env, "sqlite", "sqlite-devel")
	}
	return installPkg(ctx, env, "sqlite3", "libsqlite3-dev")
}

func init() { registerBuildAndVCS() }
