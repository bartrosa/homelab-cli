// Package packager installs system packages via native backends (brew, apt, dnf).
package packager

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/bartrosa/homelab-cli/internal/executil"
	"github.com/bartrosa/homelab-cli/internal/platform"
)

// Manager installs packages for the detected platform.
type Manager struct {
	Info   platform.Info
	Runner *executil.Runner
}

// New creates a package manager for the current host.
func New(stdout, stderr io.Writer, dryRun bool) *Manager {
	r := executil.NewRunner(stdout, stderr)
	r.DryRun = dryRun
	return &Manager{Info: platform.Detect(), Runner: r}
}

// Ensure installs the package if missing (idempotent).
func (m *Manager) Ensure(ctx context.Context, name string) error {
	installed, err := m.IsInstalled(ctx, name)
	if err != nil {
		return err
	}
	if installed {
		return nil
	}
	return m.Install(ctx, name)
}

// Install installs a single package.
func (m *Manager) Install(ctx context.Context, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("empty package name")
	}
	switch m.Info.Packager {
	case platform.PackagerBrew:
		return m.Runner.Run(ctx, "brew", "install", name)
	case platform.PackagerAPT:
		return m.Runner.Run(ctx, "sudo", "apt-get", "install", "-y", name)
	case platform.PackagerDNF:
		if m.Info.IsSilverblue {
			return m.Runner.Run(ctx, "rpm-ostree", "install", "-y", "--allow-inactive", name)
		}
		return m.Runner.Run(ctx, "sudo", "dnf", "install", "-y", name)
	default:
		return fmt.Errorf("no supported package manager on %s (install %q manually)", m.Info.GOOS, name)
	}
}

// IsInstalled checks whether the package appears installed (best-effort per backend).
func (m *Manager) IsInstalled(ctx context.Context, name string) (bool, error) {
	switch m.Info.Packager {
	case platform.PackagerBrew:
		err := m.Runner.RunQuiet(ctx, "brew", "list", "--formula", name)
		return err == nil, nil
	case platform.PackagerAPT:
		err := m.Runner.RunQuiet(ctx, "dpkg", "-s", name)
		return err == nil, nil
	case platform.PackagerDNF:
		err := m.Runner.RunQuiet(ctx, "rpm", "-q", name)
		return err == nil, nil
	default:
		return false, nil
	}
}

// ListTracked returns names the manager would use for common lab deps (informational).
func (m *Manager) ListTracked() []string {
	return []string{"ripgrep", "jq", "git", "podman"}
}
