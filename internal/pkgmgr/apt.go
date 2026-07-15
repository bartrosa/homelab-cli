package pkgmgr

import (
	"context"
	"fmt"
	"strings"

	"github.com/bartrosa/homelab-cli/internal/exec"
)

// APT implements apt-get on Debian/Ubuntu.
type APT struct {
	Runner exec.Runner
	Sudo   bool
}

// Name returns the manager identifier.
func (a *APT) Name() string { return "apt" }

// Available reports whether apt is usable.
func (a *APT) Available() bool { return true }

// UpdateCache runs apt-get update.
func (a *APT) UpdateCache(ctx context.Context) error {
	return a.run(ctx, "apt-get", "update")
}

// Install installs packages via apt-get.
func (a *APT) Install(ctx context.Context, packages ...string) error {
	if len(packages) == 0 {
		return nil
	}
	args := []string{"apt-get", "install", "-y", "--no-install-recommends"}
	args = append(args, packages...)
	return a.run(ctx, args[0], args[1:]...)
}

// IsInstalled checks dpkg -s exit status.
func (a *APT) IsInstalled(ctx context.Context, pkg string) (bool, error) {
	err := a.Runner.Run(ctx, "dpkg", "-s", pkg)
	return err == nil, nil
}

func (a *APT) run(ctx context.Context, name string, args ...string) error {
	if a.Sudo {
		return a.Runner.Run(ctx, "sudo", append([]string{name}, args...)...)
	}
	return a.Runner.Run(ctx, name, args...)
}

// BuildAPTCommand returns the argv for apt install (for tests).
func BuildAPTCommand(packages ...string) []string {
	args := []string{"apt-get", "install", "-y", "--no-install-recommends"}
	return append(args, packages...)
}

// ParseDPKGInstalled interprets dpkg -s exit as installed state.
func ParseDPKGInstalled(err error) bool {
	return err == nil
}

// JoinPackages joins package names for logging.
func JoinPackages(packages ...string) string {
	return strings.Join(packages, " ")
}

// RequireRunner ensures runner is set.
func RequireRunner(r exec.Runner) exec.Runner {
	if r == nil {
		panic("nil runner")
	}
	return r
}

// ErrUnavailable indicates no supported manager was found.
var ErrUnavailable = fmt.Errorf("no supported package manager detected")
