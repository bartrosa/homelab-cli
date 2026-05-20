// Package services manages compose stacks from the homelab repository.
package services

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/bartrosa/homelab-cli/internal/executil"
	"github.com/bartrosa/homelab-cli/internal/homelabroot"
)

// Stack names supported out of the box.
const (
	StackML = "ml-stack"
)

var stackCompose = map[string]string{
	StackML: "ml-stack/podman-compose.yml",
}

// Runner controls compose stacks.
type Runner struct {
	Stdout      io.Writer
	Stderr      io.Writer
	HomelabRoot string
	Runtime     string // podman-compose | docker compose
	DryRun      bool
}

// NewRunner creates a stack runner; runtime defaults to podman-compose.
func NewRunner(stdout, stderr io.Writer, homelabRoot, runtime string, dryRun bool) *Runner {
	if runtime == "" {
		runtime = "podman-compose"
	}
	return &Runner{
		Stdout: stdout, Stderr: stderr,
		HomelabRoot: homelabRoot, Runtime: runtime, DryRun: dryRun,
	}
}

// Up starts one or more stacks.
func (r *Runner) Up(ctx context.Context, names ...string) error {
	for _, name := range names {
		if err := r.runCompose(ctx, name, "up", "-d"); err != nil {
			return err
		}
	}
	return nil
}

// Down stops stacks.
func (r *Runner) Down(ctx context.Context, names ...string) error {
	for _, name := range names {
		if err := r.runCompose(ctx, name, "down"); err != nil {
			return err
		}
	}
	return nil
}

// Logs tails a stack (follow).
func (r *Runner) Logs(ctx context.Context, name string) error {
	return r.runCompose(ctx, name, "logs", "-f")
}

// List returns known stack names.
func (r *Runner) List() []StackInfo {
	var out []StackInfo
	for name, rel := range stackCompose {
		out = append(out, StackInfo{Name: name, ComposeFile: rel})
	}
	return out
}

// StackInfo describes a compose stack.
type StackInfo struct {
	Name        string
	ComposeFile string
}

func (r *Runner) runCompose(ctx context.Context, stack, subcmd string, extra ...string) error {
	rel, ok := stackCompose[strings.TrimSpace(stack)]
	if !ok {
		return fmt.Errorf("unknown stack %q (known: %s)", stack, strings.Join(r.knownNames(), ", "))
	}
	root, err := homelabroot.Resolve(r.HomelabRoot)
	if err != nil {
		return err
	}
	composePath := filepath.Join(root, rel)
	dir := filepath.Dir(composePath)
	file := filepath.Base(composePath)

	ex := executil.NewRunner(r.Stdout, r.Stderr)
	ex.DryRun = r.DryRun
	ex.WorkDir = dir

	args := []string{"-f", file, subcmd}
	args = append(args, extra...)

	switch {
	case strings.HasPrefix(r.Runtime, "podman"):
		return ex.Run(ctx, "podman-compose", args...)
	case strings.HasPrefix(r.Runtime, "docker"):
		dargs := append([]string{"compose", "-f", file, subcmd}, extra...)
		return ex.Run(ctx, "docker", dargs...)
	default:
		return fmt.Errorf("unsupported runtime %q", r.Runtime)
	}
}

func (r *Runner) knownNames() []string {
	var n []string
	for k := range stackCompose {
		n = append(n, k)
	}
	return n
}
