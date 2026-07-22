package podman

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/bartrosa/homelab-cli/internal/exec"
	"github.com/bartrosa/homelab-cli/internal/platform"
)

// Install method identifiers.
const (
	MethodAPT        = "apt"
	MethodDNF        = "dnf"
	MethodSilverblue = "silverblue"
	MethodBrew       = "brew"
)

// Method describes how Podman will be installed on this host.
type Method struct {
	Kind   string
	Reason string
}

// InstallOpts controls Install / Upgrade / Remove behaviour.
type InstallOpts struct {
	Info   platform.Info
	DryRun bool
	Stdout io.Writer
	Stderr io.Writer
	// Purge also removes local container storage and user config (destructive).
	Purge bool
}

var aptPackages = []string{
	"podman", "podman-docker", "uidmap", "slirp4netns", "fuse-overlayfs", "passt",
}

var dnfPackages = []string{"podman", "podman-docker"}

// DetectMethod picks an install method from platform.Info.
func DetectMethod(info platform.Info) (Method, error) {
	switch info.GOOS {
	case platform.OSDarwin:
		if info.Packager == platform.PackagerBrew {
			return Method{
				Kind:   MethodBrew,
				Reason: "macOS with Homebrew (Podman runs in a VM; skip Linux rootless/Quadlet steps)",
			}, nil
		}
		return Method{}, fmt.Errorf("homebrew not found; install brew or Podman manually")

	case platform.OSLinux:
		if info.IsSilverblue {
			return Method{
				Kind:   MethodSilverblue,
				Reason: "Fedora Silverblue — Podman is in the base image; verify only (rpm-ostree for extras needs reboot)",
			}, nil
		}
		switch info.Packager {
		case platform.PackagerAPT:
			return Method{
				Kind:   MethodAPT,
				Reason: "Debian/Ubuntu with apt",
			}, nil
		case platform.PackagerDNF:
			return Method{
				Kind:   MethodDNF,
				Reason: "Fedora/RHEL with dnf",
			}, nil
		default:
			return Method{}, fmt.Errorf("no supported package manager on linux (need apt, dnf, or silverblue)")
		}

	default:
		return Method{}, fmt.Errorf("unsupported OS %q for Podman install", info.GOOS)
	}
}

// PlannedInstallCommands returns argv lists for Install.
func PlannedInstallCommands(method Method) [][]string {
	switch method.Kind {
	case MethodAPT:
		return [][]string{
			{"sudo", "apt-get", "update"},
			append([]string{"sudo", "apt-get", "install", "-y"}, aptPackages...),
		}
	case MethodDNF:
		return [][]string{
			append([]string{"sudo", "dnf", "install", "-y"}, dnfPackages...),
		}
	case MethodBrew:
		return [][]string{
			{"brew", "install", "podman"},
			{"podman", "machine", "init"},
			{"podman", "machine", "start"},
		}
	case MethodSilverblue:
		return nil // verify-only path
	default:
		return nil
	}
}

// PlannedUpgradeCommands returns argv lists for Upgrade.
func PlannedUpgradeCommands(method Method) [][]string {
	switch method.Kind {
	case MethodAPT:
		return [][]string{
			{"sudo", "apt-get", "update"},
			append([]string{"sudo", "apt-get", "install", "--only-upgrade", "-y"}, aptPackages...),
		}
	case MethodDNF:
		return [][]string{
			append([]string{"sudo", "dnf", "upgrade", "-y"}, dnfPackages...),
		}
	case MethodBrew:
		return [][]string{{"brew", "upgrade", "podman"}}
	case MethodSilverblue:
		return [][]string{{"rpm-ostree", "upgrade"}}
	default:
		return nil
	}
}

// PlannedRemoveCommands returns argv lists for Remove (packages only; purge is separate).
func PlannedRemoveCommands(method Method) [][]string {
	switch method.Kind {
	case MethodAPT:
		return [][]string{
			append([]string{"sudo", "apt-get", "remove", "-y"}, aptPackages...),
		}
	case MethodDNF:
		return [][]string{
			append([]string{"sudo", "dnf", "remove", "-y"}, dnfPackages...),
		}
	case MethodBrew:
		return [][]string{{"brew", "uninstall", "podman"}}
	case MethodSilverblue:
		return nil // do not layer-remove base Podman
	default:
		return nil
	}
}

// IsInstalled reports whether podman is on PATH and runnable.
func IsInstalled(ctx context.Context, runner exec.Runner) (bool, error) {
	_, err := runner.RunWithOutput(ctx, "podman", "--version")
	return err == nil, nil
}

// Version returns the output of `podman --version`.
func Version(ctx context.Context, runner exec.Runner) (string, error) {
	out, err := runner.RunWithOutput(ctx, "podman", "--version")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// Install installs Podman using the method for opts.Info (or platform.Detect).
// On Silverblue, verifies the binary is present and does not layer packages.
func Install(ctx context.Context, runner exec.Runner, opts InstallOpts) error {
	info := opts.Info
	if info.GOOS == "" {
		info = platform.Detect()
	}
	method, err := DetectMethod(info)
	if err != nil {
		return err
	}

	if method.Kind == MethodSilverblue {
		return installSilverblue(ctx, runner, opts, method)
	}

	cmds := PlannedInstallCommands(method)
	if opts.DryRun {
		writePlan(opts.Stdout, method, "install", cmds)
		return nil
	}
	for _, argv := range cmds {
		if len(argv) == 0 {
			continue
		}
		// macOS: machine init may fail if the VM already exists — continue to start.
		if method.Kind == MethodBrew && len(argv) >= 3 && argv[0] == "podman" && argv[1] == "machine" && argv[2] == "init" {
			_ = runner.Run(ctx, argv[0], argv[1:]...)
			continue
		}
		if err := runner.Run(ctx, argv[0], argv[1:]...); err != nil {
			return err
		}
	}
	return nil
}

func installSilverblue(ctx context.Context, runner exec.Runner, opts InstallOpts, method Method) error {
	if opts.DryRun {
		writePlan(opts.Stdout, method, "verify (silverblue)", nil)
		if opts.Stdout != nil {
			_, _ = fmt.Fprintln(opts.Stdout, "  # podman is expected in the Silverblue base image")
			_, _ = fmt.Fprintln(opts.Stdout, "  podman --version")
		}
		return nil
	}
	ok, err := IsInstalled(ctx, runner)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("podman not found on Silverblue; layer with: rpm-ostree install podman (requires reboot)")
	}
	if opts.Stdout != nil {
		ver, _ := Version(ctx, runner)
		_, _ = fmt.Fprintf(opts.Stdout, "Podman already present on Silverblue (%s) — skipping package install\n", ver)
	}
	return nil
}

// Upgrade upgrades Podman via the OS package manager.
func Upgrade(ctx context.Context, runner exec.Runner, opts InstallOpts) error {
	info := opts.Info
	if info.GOOS == "" {
		info = platform.Detect()
	}
	method, err := DetectMethod(info)
	if err != nil {
		return err
	}
	cmds := PlannedUpgradeCommands(method)
	if opts.DryRun {
		writePlan(opts.Stdout, method, "upgrade", cmds)
		if method.Kind == MethodSilverblue && opts.Stdout != nil {
			_, _ = fmt.Fprintln(opts.Stdout, "  # rpm-ostree upgrade typically requires a reboot")
		}
		return nil
	}
	for _, argv := range cmds {
		if len(argv) == 0 {
			continue
		}
		if err := runner.Run(ctx, argv[0], argv[1:]...); err != nil {
			return err
		}
	}
	return nil
}

// Remove uninstalls Podman packages. With Purge, also removes user container storage
// and config under ~/.local/share/containers and ~/.config/containers.
func Remove(ctx context.Context, runner exec.Runner, opts InstallOpts) error {
	info := opts.Info
	if info.GOOS == "" {
		info = platform.Detect()
	}
	method, err := DetectMethod(info)
	if err != nil {
		return err
	}
	if method.Kind == MethodSilverblue {
		return fmt.Errorf("refusing to remove Podman from Silverblue base image")
	}
	cmds := PlannedRemoveCommands(method)
	if opts.DryRun {
		writePlan(opts.Stdout, method, "remove", cmds)
		if opts.Purge && opts.Stdout != nil {
			_, _ = fmt.Fprintln(opts.Stdout, "  # --purge: remove ~/.local/share/containers and ~/.config/containers")
		}
		return nil
	}
	for _, argv := range cmds {
		if len(argv) == 0 {
			continue
		}
		if err := runner.Run(ctx, argv[0], argv[1:]...); err != nil {
			return err
		}
	}
	return nil
}

func writePlan(w io.Writer, method Method, action string, cmds [][]string) {
	if w == nil {
		return
	}
	_, _ = fmt.Fprintf(w, "method: %s (%s)\n", method.Kind, method.Reason)
	_, _ = fmt.Fprintf(w, "action: %s\n", action)
	for _, argv := range cmds {
		_, _ = fmt.Fprintf(w, "  %s\n", strings.Join(argv, " "))
	}
}
