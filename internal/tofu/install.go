package tofu

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/bartrosa/homelab-cli/internal/exec"
	"github.com/bartrosa/homelab-cli/internal/executil"
	"github.com/bartrosa/homelab-cli/internal/platform"
)

// Install method identifiers.
const (
	MethodBrew   = "brew"
	MethodScript = "script"
	MethodSnap   = "snap"
	MethodWinget = "winget"
)

const (
	opentofuBrewPkg  = "opentofu"
	opentofuWingetID = "OpenTofu.OpenTofu"
	installScriptURL = "https://get.opentofu.org/install-opentofu.sh"
)

// Method describes how OpenTofu will be installed on this host.
type Method struct {
	Kind         string // brew | script | snap | winget
	Reason       string
	ScriptMethod string // deb | rpm when Kind is script (official installer flag; not "auto")
}

// InstallOpts controls Install / Upgrade behaviour.
type InstallOpts struct {
	Info   platform.Info
	DryRun bool
	Stdout io.Writer
	Stderr io.Writer
}

// DetectMethod picks an install method from platform.Info.
// Snap is a Linux fallback when no native packager is detected.
func DetectMethod(info platform.Info) (Method, error) {
	return DetectMethodWith(info, executil.CommandExists)
}

// DetectMethodWith is like DetectMethod but uses hasCmd for binary presence checks (tests).
func DetectMethodWith(info platform.Info, hasCmd func(string) bool) (Method, error) {
	return detectMethod(info, hasCmd)
}

func detectMethod(info platform.Info, hasCmd func(string) bool) (Method, error) {
	switch info.GOOS {
	case platform.OSDarwin:
		if info.Packager == platform.PackagerBrew || hasCmd("brew") {
			return Method{
				Kind:   MethodBrew,
				Reason: "macOS with Homebrew",
			}, nil
		}
		return Method{}, fmt.Errorf("Homebrew not found; install brew or OpenTofu manually")

	case platform.OSLinux:
		switch info.Packager {
		case platform.PackagerAPT:
			return Method{
				Kind:         MethodScript,
				ScriptMethod: "deb",
				Reason:       "Linux with " + info.PackagerLabel() + "; official install-opentofu.sh (--install-method deb)",
			}, nil
		case platform.PackagerDNF:
			return Method{
				Kind:         MethodScript,
				ScriptMethod: "rpm",
				Reason:       "Linux with " + info.PackagerLabel() + "; official install-opentofu.sh (--install-method rpm)",
			}, nil
		}
		if hasCmd("snap") {
			return Method{
				Kind:   MethodSnap,
				Reason: "Linux without apt/dnf; snap available",
			}, nil
		}
		return Method{}, fmt.Errorf("no supported install method (need apt, dnf, rpm-ostree, or snap)")

	case "windows":
		if hasCmd("winget") {
			return Method{
				Kind:   MethodWinget,
				Reason: "Windows with winget",
			}, nil
		}
		return Method{}, fmt.Errorf("winget not found; install OpenTofu manually")

	default:
		return Method{}, fmt.Errorf("unsupported OS %q for OpenTofu install", info.GOOS)
	}
}

// PlannedCommands returns the argv lists that Install/Upgrade would run for method.
func PlannedCommands(method Method, upgrade bool) [][]string {
	switch method.Kind {
	case MethodBrew:
		if upgrade {
			return [][]string{{"brew", "upgrade", opentofuBrewPkg}}
		}
		return [][]string{{"brew", "install", opentofuBrewPkg}}
	case MethodScript:
		sm := method.ScriptMethod
		if sm == "" {
			sm = "standalone"
		}
		return [][]string{{"bash", "-c", officialInstallScript(sm)}}
	case MethodSnap:
		if upgrade {
			return [][]string{{"sudo", "snap", "refresh", "opentofu"}}
		}
		return [][]string{{"sudo", "snap", "install", "--classic", "opentofu"}}
	case MethodWinget:
		if upgrade {
			return [][]string{{"winget", "upgrade", "--id", opentofuWingetID}}
		}
		return [][]string{{"winget", "install", "--id", opentofuWingetID}}
	default:
		return nil
	}
}

func officialInstallScript(installMethod string) string {
	// Official installer does not support --install-method auto; use deb/rpm/standalone/…
	return strings.TrimSpace(`
set -euo pipefail
tmp=$(mktemp)
trap 'rm -f "$tmp"' EXIT
curl --proto '=https' --tlsv1.2 -fsSL ` + installScriptURL + ` -o "$tmp"
chmod +x "$tmp"
sudo "$tmp" --install-method ` + installMethod + `
`)
}

// IsInstalled reports whether `tofu` is on PATH and runnable.
// On success, version is the trimmed output of `tofu version`.
func IsInstalled(ctx context.Context, runner exec.Runner) (bool, string, error) {
	out, err := runner.RunWithOutput(ctx, "tofu", "version")
	if err != nil {
		return false, "", nil
	}
	return true, strings.TrimSpace(out), nil
}

// Install installs the OpenTofu binary using the method for opts.Info
// (or platform.Detect when Info.GOOS is empty).
func Install(ctx context.Context, runner exec.Runner, opts InstallOpts) error {
	return installOrUpgrade(ctx, runner, opts, false)
}

// Upgrade upgrades an existing OpenTofu install (brew upgrade / snap refresh / winget upgrade;
// script method re-runs the official installer).
func Upgrade(ctx context.Context, runner exec.Runner, opts InstallOpts) error {
	return installOrUpgrade(ctx, runner, opts, true)
}

func installOrUpgrade(ctx context.Context, runner exec.Runner, opts InstallOpts, upgrade bool) error {
	info := opts.Info
	if info.GOOS == "" {
		info = platform.Detect()
	}
	method, err := DetectMethod(info)
	if err != nil {
		return err
	}

	cmds := PlannedCommands(method, upgrade)
	if opts.DryRun {
		writeInstallPlan(opts.Stdout, method, upgrade, cmds)
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

func writeInstallPlan(w io.Writer, method Method, upgrade bool, cmds [][]string) {
	if w == nil {
		return
	}
	action := "install"
	if upgrade {
		action = "upgrade"
	}
	_, _ = fmt.Fprintf(w, "method: %s (%s)\n", method.Kind, method.Reason)
	_, _ = fmt.Fprintf(w, "action: %s\n", action)
	for _, argv := range cmds {
		_, _ = fmt.Fprintf(w, "  %s\n", strings.Join(argv, " "))
	}
}
