// Package executil runs external commands with consistent logging and dry-run support.
package executil

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// Runner executes shell commands.
type Runner struct {
	Stdout   io.Writer
	Stderr   io.Writer
	DryRun   bool
	Env      []string
	WorkDir  string
	Inherit  bool // append os.Environ when true
}

// NewRunner returns a runner writing to stdout/stderr.
func NewRunner(stdout, stderr io.Writer) *Runner {
	return &Runner{Stdout: stdout, Stderr: stderr, Inherit: true}
}

// Run executes name with args. Returns combined error from start/wait.
func (r *Runner) Run(ctx context.Context, name string, args ...string) error {
	if r.DryRun {
		_, _ = fmt.Fprintf(r.Stderr, "[dry-run] %s %s\n", name, strings.Join(args, " "))
		return nil
	}
	cmd := exec.CommandContext(ctx, name, args...)
	if r.Inherit {
		cmd.Env = append(os.Environ(), r.Env...)
	} else if len(r.Env) > 0 {
		cmd.Env = r.Env
	}
	if r.WorkDir != "" {
		cmd.Dir = r.WorkDir
	}
	cmd.Stdout = r.Stdout
	cmd.Stderr = r.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %s: %w", name, strings.Join(args, " "), err)
	}
	return nil
}

// LookPath wraps exec.LookPath.
func LookPath(name string) (string, error) {
	return exec.LookPath(name)
}

// CommandExists reports whether binary is on PATH.
func CommandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// Output runs a command and returns stdout bytes.
func (r *Runner) Output(ctx context.Context, name string, args ...string) ([]byte, error) {
	if r.DryRun {
		return nil, nil
	}
	cmd := exec.CommandContext(ctx, name, args...)
	if r.Inherit {
		cmd.Env = append(os.Environ(), r.Env...)
	}
	if r.WorkDir != "" {
		cmd.Dir = r.WorkDir
	}
	cmd.Stderr = r.Stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w", name, strings.Join(args, " "), err)
	}
	return out, nil
}

// RunQuiet runs a command and discards stdout/stderr (for presence checks).
func (r *Runner) RunQuiet(ctx context.Context, name string, args ...string) error {
	if r.DryRun {
		return nil
	}
	cmd := exec.CommandContext(ctx, name, args...)
	if r.Inherit {
		cmd.Env = append(os.Environ(), r.Env...)
	}
	if r.WorkDir != "" {
		cmd.Dir = r.WorkDir
	}
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %s: %w", name, strings.Join(args, " "), err)
	}
	return nil
}
