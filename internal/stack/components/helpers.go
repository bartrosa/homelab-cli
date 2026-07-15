package components

import (
	"context"
	"fmt"
	"strings"

	"github.com/bartrosa/homelab-cli/internal/stack"
)

func cmdExists(ctx context.Context, env *stack.Env, name string) bool {
	_, err := env.Runner.RunWithOutput(ctx, "sh", "-c", "command -v "+name)
	return err == nil
}

func versionOf(ctx context.Context, env *stack.Env, cmd string, args ...string) string {
	out, err := env.Runner.RunWithOutput(ctx, cmd, args...)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(strings.Split(out, "\n")[0])
}

func miseSpec(id, version string) string {
	if version == "" || version == "latest" || version == "lts" {
		return id + "@latest"
	}
	return id + "@" + version
}

func ensureMise(ctx context.Context, env *stack.Env, dryRun bool) error {
	if cmdExists(ctx, env, "mise") {
		return nil
	}
	if dryRun {
		return env.Runner.Run(ctx, "sh", "-c", "curl https://mise.jdx.dev/install.sh | sh")
	}
	return env.Runner.Run(ctx, "sh", "-c", "curl https://mise.jdx.dev/install.sh | sh")
}

func miseInstalled(ctx context.Context, env *stack.Env, id string) (bool, string) {
	if !cmdExists(ctx, env, "mise") {
		return false, ""
	}
	out, err := env.Runner.RunWithOutput(ctx, "mise", "ls", "-g", "--json")
	if err != nil {
		out, _ = env.Runner.RunWithOutput(ctx, "mise", "ls", "-g")
	}
	if strings.Contains(out, id) {
		return true, versionOf(ctx, env, "mise", "current", id)
	}
	return false, ""
}

func pkgInstalled(ctx context.Context, env *stack.Env, pkg string) (bool, error) {
	if env.PkgMgr == nil {
		return cmdExists(ctx, env, pkg), nil
	}
	return env.PkgMgr.IsInstalled(ctx, pkg)
}

func installPkg(ctx context.Context, env *stack.Env, pkgs ...string) error {
	if env.PkgMgr == nil {
		return fmt.Errorf("no package manager available")
	}
	return env.PkgMgr.Install(ctx, pkgs...)
}
