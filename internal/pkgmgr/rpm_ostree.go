package pkgmgr

import (
	"context"
	"fmt"
	"strings"

	"github.com/bartrosa/homelab-cli/internal/exec"
)

// RPMOstree implements Fedora Silverblue layering.
type RPMOstree struct {
	Runner exec.Runner
}

// Name returns the manager identifier.
func (r *RPMOstree) Name() string { return "rpm-ostree" }

// Available reports whether rpm-ostree is usable.
func (r *RPMOstree) Available() bool { return true }

// UpdateCache is a no-op for rpm-ostree.
func (r *RPMOstree) UpdateCache(_ context.Context) error {
	// rpm-ostree has no separate cache refresh for installs
	return nil
}

// Install layers packages via rpm-ostree install --idempotent.
func (r *RPMOstree) Install(ctx context.Context, packages ...string) error {
	if len(packages) == 0 {
		return nil
	}
	args := append([]string{"install", "--idempotent", "-y"}, packages...)
	return r.Runner.Run(ctx, "rpm-ostree", args...)
}

// IsInstalled checks rpm -q exit status.
func (r *RPMOstree) IsInstalled(ctx context.Context, pkg string) (bool, error) {
	err := r.Runner.Run(ctx, "rpm", "-q", pkg)
	return err == nil, nil
}

// BuildRPMOstreeCommand returns argv for install (for tests).
func BuildRPMOstreeCommand(packages ...string) []string {
	args := []string{"install", "--idempotent", "-y"}
	return append(args, packages...)
}

// NeedsRebootHint is printed after rpm-ostree layers packages.
const NeedsRebootHint = "rpm-ostree changes require a reboot before new packages are available"

// FormatRebootWarning returns a user-facing reboot notice.
func FormatRebootWarning(section string) string {
	return fmt.Sprintf("[%s] %s", section, NeedsRebootHint)
}

// StripGroup keeps @group names intact for rpm.
func StripGroup(pkg string) string {
	return strings.TrimSpace(pkg)
}

// NewRPMOstree returns a Silverblue manager.
func NewRPMOstree(runner exec.Runner) *RPMOstree {
	return &RPMOstree{Runner: runner}
}
