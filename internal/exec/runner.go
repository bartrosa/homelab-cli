// Package exec provides a testable interface for running external commands.
package exec

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	osexec "os/exec"
	"strings"
)

// Runner executes external programs.
type Runner interface {
	Run(ctx context.Context, name string, args ...string) error
	RunWithOutput(ctx context.Context, name string, args ...string) (string, error)
}

// OSRunner runs commands via os/exec.
type OSRunner struct {
	Stdout  io.Writer
	Stderr  io.Writer
	WorkDir string
	Env     []string
}

// NewOSRunner returns a runner writing to stdout/stderr.
func NewOSRunner(stdout, stderr io.Writer) *OSRunner {
	return &OSRunner{Stdout: stdout, Stderr: stderr}
}

// Run executes name with args.
func (r *OSRunner) Run(ctx context.Context, name string, args ...string) error {
	cmd := osexec.CommandContext(ctx, name, args...)
	cmd.Stdout = r.Stdout
	cmd.Stderr = r.Stderr
	cmd.Dir = r.WorkDir
	if len(r.Env) > 0 {
		cmd.Env = append(os.Environ(), r.Env...)
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %s: %w", name, strings.Join(args, " "), err)
	}
	return nil
}

// RunWithOutput runs a command and returns combined stdout.
func (r *OSRunner) RunWithOutput(ctx context.Context, name string, args ...string) (string, error) {
	cmd := osexec.CommandContext(ctx, name, args...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	if r.Stderr != nil {
		cmd.Stderr = r.Stderr
	}
	cmd.Dir = r.WorkDir
	if len(r.Env) > 0 {
		cmd.Env = append(os.Environ(), r.Env...)
	}
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s %s: %w", name, strings.Join(args, " "), err)
	}
	return strings.TrimSpace(buf.String()), nil
}
