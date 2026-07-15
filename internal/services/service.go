// Package services manages local compose-backed homelab data services.
package services

import (
	"context"
	"io"

	"github.com/bartrosa/homelab-cli/internal/exec"
	"github.com/bartrosa/homelab-cli/internal/prompt"
)

// Category groups homelab services.
type Category string

// Service categories group compose-backed homelab services.
const (
	CategoryDatabase      Category = "database"
	CategoryCache         Category = "cache"
	CategoryMessageQueue  Category = "message-queue"
	CategoryVector        Category = "vector"
	CategoryObservability Category = "observability"
	CategoryStorage       Category = "storage"
)

// Service is a provisionable compose-backed homelab service.
type Service interface {
	ID() string
	DisplayName() string
	Category() Category
	Description() string
	Schema() ConfigSchema
	DependsOn() []string
	Init(ctx context.Context, opts InitOptions) error
	Status(ctx context.Context, opts InitOptions) (Status, error)
	Up(ctx context.Context, opts InitOptions) error
	Down(ctx context.Context, opts InitOptions) error
}

// InitOptions configures service lifecycle operations.
type InitOptions struct {
	Runner         exec.Runner
	Stdout         io.Writer
	Stderr         io.Writer
	Runtime        string // auto, docker, podman
	DryRun         bool
	Force          bool
	NonInteractive bool
	Values         map[string]any
	Prompter       prompt.Prompter
}

// Status describes runtime state for a service.
type Status struct {
	ID      string
	Running bool
	Detail  string
}

// ServiceMeta holds static metadata for a managed compose service.
type ServiceMeta struct {
	ID          string
	DisplayName string
	Category    Category
	Description string
}
