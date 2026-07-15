package bootstrap

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/bartrosa/homelab-cli/internal/exec"
)

// SetupDistrobox installs distrobox and optionally creates homelab-dev container.
func SetupDistrobox(ctx context.Context, opts EssentialsOptions) (bool, error) {
	pkg, _ := PackageFor("distrobox", "rpm-ostree")
	installed, err := opts.Runner.RunWithOutput(ctx, "which", "distrobox")
	reboot := false
	if err != nil || strings.TrimSpace(installed) == "" {
		if err := opts.Runner.Run(ctx, "rpm-ostree", "install", "--idempotent", "-y", pkg); err != nil {
			return false, err
		}
		reboot = true
	}

	create := opts.Yes
	if !create && !opts.Yes {
		fmt.Fprint(opts.Stdout, "Create distrobox homelab-dev container? [y/N]: ")
		var ans string
		_, _ = fmt.Scanln(&ans)
		create = strings.EqualFold(strings.TrimSpace(ans), "y") || strings.EqualFold(strings.TrimSpace(ans), "yes")
	}
	if !create {
		return reboot, nil
	}

	if err := opts.Runner.Run(ctx, "distrobox", "create", "--name", "homelab-dev", "--image", "quay.io/fedora/fedora:41"); err != nil {
		return reboot, fmt.Errorf("distrobox create: %w", err)
	}
	_ = opts.Runner.Run(ctx, "distrobox", "enter", "homelab-dev", "--", "sudo", "dnf", "install", "-y", "@development-tools", "git", "curl")
	_ = opts.Runner.Run(ctx, "distrobox", "enter", "homelab-dev", "--", "sh", "-c", "curl https://mise.run | sh")
	home, _ := os.UserHomeDir()
	misePath := home + "/.local/share/mise/bin/mise"
	exportPath := home + "/.local/bin"
	_ = opts.Runner.Run(ctx, "distrobox-export", "--bin", misePath, "--export-path", exportPath)
	return reboot, nil
}

// EnsureFlathub adds Flathub remote if missing.
func EnsureFlathub(ctx context.Context, runner exec.Runner) error {
	out, err := runner.RunWithOutput(ctx, "flatpak", "remote-list")
	if err != nil {
		return err
	}
	sc := bufio.NewScanner(strings.NewReader(out))
	for sc.Scan() {
		if strings.Contains(sc.Text(), "flathub") {
			return nil
		}
	}
	return runner.Run(ctx, "flatpak", "remote-add", "--if-not-exists", "flathub", "https://flathub.org/repo/flathub.flatpakrepo")
}
