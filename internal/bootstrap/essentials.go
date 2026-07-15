package bootstrap

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/bartrosa/homelab-cli/internal/exec"
	"github.com/bartrosa/homelab-cli/internal/pkgmgr"
)

// EssentialsOptions configures bootstrap essentials.
type EssentialsOptions struct {
	Target string
	Yes    bool
	Skip   []string
	Only   []string
	DryRun bool
	Stdout io.Writer
	Stderr io.Writer
	Runner exec.Runner
}

// RunEssentials executes selected bootstrap sections.
func RunEssentials(ctx context.Context, opts EssentialsOptions) error {
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}
	if opts.Stderr == nil {
		opts.Stderr = os.Stderr
	}
	if opts.Runner == nil {
		opts.Runner = exec.NewOSRunner(opts.Stdout, opts.Stderr)
	}

	mgr, osType, err := pkgmgr.DetectTarget(opts.Target, opts.Runner)
	if err != nil {
		return err
	}
	mgrName := mgr.Name()

	sections := FilterSections(AllEssentialSections(), opts.Only, opts.Skip)
	if len(sections) == 0 {
		return fmt.Errorf("no sections selected")
	}

	fmt.Fprintf(opts.Stdout, "Target: %s (%s)\n\n", osType, mgrName)

	for _, section := range sections {
		fmt.Fprintf(opts.Stdout, "== %s ==\n", section)
		if opts.DryRun {
			if err := planSection(opts.Stdout, section, mgrName, osType); err != nil {
				return err
			}
			fmt.Fprintf(opts.Stdout, "⏭️  dry-run\n\n")
			continue
		}
		reboot, err := runSection(ctx, opts, mgr, mgrName, osType, section)
		if err != nil {
			fmt.Fprintf(opts.Stderr, "❌ failed: %v\n", err)
			return err
		}
		fmt.Fprintf(opts.Stdout, "✅ done\n")
		if reboot {
			fmt.Fprintf(opts.Stdout, "ℹ️  %s\n", pkgmgr.NeedsRebootHint)
		}
		fmt.Fprintln(opts.Stdout)
	}
	return nil
}

func planSection(w io.Writer, section, mgrName string, osType pkgmgr.OSType) error {
	switch section {
	case "system-update":
		fmt.Fprintf(w, "[plan] %s update cache\n", mgrName)
	case "cli-basics":
		for _, p := range CLIBasicPackages() {
			pkg, _ := PackageFor(p, mgrName)
			fmt.Fprintf(w, "[plan] install %s\n", pkg)
		}
	case "shell-tools":
		for _, p := range ShellToolPackages() {
			pkg, _ := PackageFor(p, mgrName)
			fmt.Fprintf(w, "[plan] install %s\n", pkg)
		}
	case "build":
		pkg, _ := PackageFor("build", mgrName)
		fmt.Fprintf(w, "[plan] install %s\n", pkg)
	case "container-runtime":
		if mgrName == "apt" {
			fmt.Fprintln(w, "[plan] install Docker CE via apt repository")
		} else {
			fmt.Fprintln(w, "[plan] verify podman (preinstalled on Silverblue)")
		}
	case "mise":
		fmt.Fprintln(w, "[plan] curl https://mise.run | sh")
	case "distrobox":
		if osType == pkgmgr.OSSilverblue {
			fmt.Fprintln(w, "[plan] rpm-ostree install distrobox")
			fmt.Fprintln(w, "[plan] optional: distrobox create homelab-dev")
		} else {
			fmt.Fprintln(w, "[plan] skip (Silverblue only)")
		}
	case "flatpak-flathub":
		if osType == pkgmgr.OSSilverblue {
			fmt.Fprintln(w, "[plan] flatpak remote-add flathub")
		} else {
			fmt.Fprintln(w, "[plan] skip (Silverblue only)")
		}
	default:
		return fmt.Errorf("unknown section %q", section)
	}
	return nil
}

func runSection(ctx context.Context, opts EssentialsOptions, mgr pkgmgr.Manager, mgrName string, osType pkgmgr.OSType, section string) (bool, error) {
	switch section {
	case "system-update":
		return false, mgr.UpdateCache(ctx)
	case "cli-basics":
		return installGenericPackages(ctx, mgr, mgrName, CLIBasicPackages()...)
	case "shell-tools":
		return installGenericPackages(ctx, mgr, mgrName, ShellToolPackages()...)
	case "build":
		pkg, ok := PackageFor("build", mgrName)
		if !ok {
			return false, fmt.Errorf("no build package for %s", mgrName)
		}
		return mgrName == "rpm-ostree", mgr.Install(ctx, pkg)
	case "container-runtime":
		if mgrName == "apt" {
			return false, InstallDocker(ctx, opts.Runner, opts.Stdout, opts.Stderr)
		}
		return false, ensurePodmanSilverblue(ctx, opts.Runner)
	case "mise":
		return false, RunMiseInstall(ctx, opts.Runner, opts.Stdout)
	case "distrobox":
		if osType != pkgmgr.OSSilverblue {
			fmt.Fprintln(opts.Stdout, "⏭️  skipped (Silverblue only)")
			return false, nil
		}
		reboot, err := SetupDistrobox(ctx, opts)
		return reboot, err
	case "flatpak-flathub":
		if osType != pkgmgr.OSSilverblue {
			fmt.Fprintln(opts.Stdout, "⏭️  skipped (Silverblue only)")
			return false, nil
		}
		return false, EnsureFlathub(ctx, opts.Runner)
	default:
		return false, fmt.Errorf("unknown section %q", section)
	}
}

func installGenericPackages(ctx context.Context, mgr pkgmgr.Manager, mgrName string, names ...string) (bool, error) {
	var pkgs []string
	for _, n := range names {
		pkg, ok := PackageFor(n, mgrName)
		if !ok {
			return false, fmt.Errorf("no mapping for %q on %s", n, mgrName)
		}
		installed, err := mgr.IsInstalled(ctx, pkg)
		if err != nil {
			return false, err
		}
		if installed {
			continue
		}
		pkgs = append(pkgs, pkg)
	}
	if len(pkgs) == 0 {
		return false, nil
	}
	return mgrName == "rpm-ostree", mgr.Install(ctx, pkgs...)
}

func ensurePodmanSilverblue(ctx context.Context, runner exec.Runner) error {
	_, err := runner.RunWithOutput(ctx, "podman", "--version")
	if err != nil {
		return fmt.Errorf("podman not available: %w", err)
	}
	pkg, _ := PackageFor("podman-compose", "rpm-ostree")
	if err := runner.Run(ctx, "rpm", "-q", pkg); err != nil {
		return runner.Run(ctx, "rpm-ostree", "install", "--idempotent", "-y", pkg)
	}
	return nil
}

// SectionNamesForTest exposes section filtering for tests.
func SectionNamesForTest(only, skip []string) []string {
	return FilterSections(AllEssentialSections(), only, skip)
}

// ParseCSV splits comma-separated section names.
func ParseCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
