// Package stack defines installable developer-environment components and orchestration.
package stack

import (
	"context"
	"log/slog"
)

// Category groups installable developer components.
type Category string

// Component categories group installable developer stack blocks.
const (
	CategoryLanguage         Category = "language"
	CategoryBuildTool        Category = "build-tool"
	CategoryContainer        Category = "container"
	CategoryGPU              Category = "gpu"
	CategoryPackageMgr       Category = "package-manager"
	CategoryVCS              Category = "vcs"
	CategoryDatabaseEmbedded Category = "database-embedded"
)

// Component is an installable developer-environment block.
type Component interface {
	ID() string
	DisplayName() string
	Category() Category
	Description() string
	DefaultVersion() string
	Requires() []string
	IsInstalled(ctx context.Context, env *Env) (bool, string, error)
	Install(ctx context.Context, env *Env, opts InstallOptions) error
	PathEntries() []PathEntry
}

// InstallOptions configures a component install.
type InstallOptions struct {
	Version        string
	Force          bool
	NonInteractive bool
	DryRun         bool
	SkipPath       bool
	Logger         *slog.Logger
	Extra          map[string]string // e.g. rust-toolchain, cmake-source
}

// PathEntry is a shell rc snippet for PATH and env vars.
type PathEntry struct {
	Shell   string // bash, zsh, fish, all
	Content string
	Marker  string
}

// PlanStep describes one orchestrator step.
type PlanStep struct {
	ID       string
	Action   string // install, skip
	Reason   string
	Version  string
	Requires []string
}
