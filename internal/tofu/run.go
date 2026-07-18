package tofu

import (
	"context"
	"fmt"
	"strings"

	"github.com/bartrosa/homelab-cli/internal/exec"
)

// Stack is a directory containing OpenTofu (.tf) configuration.
type Stack struct {
	Dir string
}

// RunOpts controls orchestration of tofu against a stack.
//
// Env injects process environment variables for the tofu subprocess
// (e.g. TF_VAR_* / ROS_PASSWORD). Secrets come from op:// (1Password)
// injected by an `op run` wrapper — never from files on disk.
type RunOpts struct {
	Env         []string
	AutoApprove bool
	DryRun      bool
}

// Init runs `tofu -chdir=<dir> init`.
func Init(ctx context.Context, runner exec.Runner, stack Stack, opts RunOpts) error {
	return runTofu(ctx, runner, stack, opts, "init")
}

// Validate runs `tofu -chdir=<dir> validate`.
func Validate(ctx context.Context, runner exec.Runner, stack Stack, opts RunOpts) error {
	return runTofu(ctx, runner, stack, opts, "validate")
}

// Plan runs `tofu -chdir=<dir> plan`.
func Plan(ctx context.Context, runner exec.Runner, stack Stack, opts RunOpts) error {
	return runTofu(ctx, runner, stack, opts, "plan")
}

// Apply runs `tofu -chdir=<dir> apply`. When opts.AutoApprove is true,
// passes -auto-approve (CLI maps --yes to this). Default is interactive.
func Apply(ctx context.Context, runner exec.Runner, stack Stack, opts RunOpts) error {
	extra := []string{}
	if opts.AutoApprove {
		extra = append(extra, "-auto-approve")
	}
	return runTofu(ctx, runner, stack, opts, "apply", extra...)
}

// Fmt runs `tofu -chdir=<dir> fmt`.
func Fmt(ctx context.Context, runner exec.Runner, stack Stack, opts RunOpts) error {
	return runTofu(ctx, runner, stack, opts, "fmt")
}

// CommandLine returns the argv that would be passed to tofu for debugging / dry-run UI.
func CommandLine(stack Stack, subcommand string, extra ...string) []string {
	args := []string{"-chdir=" + stack.Dir, subcommand}
	args = append(args, extra...)
	return args
}

func runTofu(ctx context.Context, runner exec.Runner, stack Stack, opts RunOpts, subcommand string, extra ...string) error {
	dir := strings.TrimSpace(stack.Dir)
	if dir == "" {
		return fmt.Errorf("stack directory is empty")
	}
	args := CommandLine(Stack{Dir: dir}, subcommand, extra...)
	if opts.DryRun {
		return nil
	}
	restore := applyEnv(runner, opts.Env)
	defer restore()
	return runner.Run(ctx, "tofu", args...)
}

func applyEnv(runner exec.Runner, env []string) func() {
	if len(env) == 0 {
		return func() {}
	}
	or, ok := runner.(*exec.OSRunner)
	if !ok {
		return func() {}
	}
	prev := or.Env
	or.Env = append(append([]string{}, prev...), env...)
	return func() { or.Env = prev }
}
