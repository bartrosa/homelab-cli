package podman

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bartrosa/homelab-cli/internal/exec"
	"github.com/bartrosa/homelab-cli/internal/platform"
)

// Check statuses for doctor output.
const (
	StatusOK   = "OK"
	StatusWarn = "WARN"
	StatusFail = "FAIL"
)

// Check is one doctor verification item.
type Check struct {
	Name   string
	Status string
	Detail string
	Fix    string
}

// Report is the full doctor result set.
type Report struct {
	Checks []Check
}

// DoctorOpts controls doctor checks.
type DoctorOpts struct {
	Info     platform.Info
	HomeDir  string
	Username string
	ReadFile func(name string) ([]byte, error)
	Stat     func(name string) (os.FileInfo, error)
}

func (o *DoctorOpts) normalize() {
	if o.ReadFile == nil {
		o.ReadFile = os.ReadFile
	}
	if o.Stat == nil {
		o.Stat = os.Stat
	}
	if o.HomeDir == "" {
		if h, err := os.UserHomeDir(); err == nil {
			o.HomeDir = h
		}
	}
	if o.Username == "" {
		o.Username = os.Getenv("USER")
	}
	if o.Info.GOOS == "" {
		o.Info = platform.Detect()
	}
}

// Doctor verifies that rootless Podman is usable and Quadlet prerequisites are in place.
func Doctor(ctx context.Context, runner exec.Runner, opts DoctorOpts) (Report, error) {
	opts.normalize()
	var checks []Check

	checks = append(checks, checkVersion(ctx, runner))
	checks = append(checks, checkInfo(ctx, runner))

	switch opts.Info.GOOS {
	case platform.OSLinux:
		checks = append(checks, checkUIDMap(ctx, runner))
		checks = append(checks, checkNetworkBackend(ctx, runner))
		checks = append(checks, checkLinger(ctx, runner, opts.Username))
		checks = append(checks, checkQuadletDir(opts))
		checks = append(checks, checkRegistries(opts))
		checks = append(checks, checkAppArmor(ctx, runner, opts.Info))
		checks = append(checks, checkSocket(ctx, runner))
	case platform.OSDarwin:
		checks = append(checks, Check{
			Name:   "rootless-linux",
			Status: StatusWarn,
			Detail: "macOS uses podman machine; Linux rootless/Quadlet checks skipped",
		})
	}

	return Report{Checks: checks}, nil
}

func checkVersion(ctx context.Context, runner exec.Runner) Check {
	out, err := runner.RunWithOutput(ctx, "podman", "--version")
	if err != nil {
		return Check{Name: "podman --version", Status: StatusFail, Detail: err.Error(), Fix: "lab podman install"}
	}
	return Check{Name: "podman --version", Status: StatusOK, Detail: strings.TrimSpace(out)}
}

func checkInfo(ctx context.Context, runner exec.Runner) Check {
	// Podman 4.x uses .Store.*; 5.x often exposes .Host.Storage.*.
	for _, tmpl := range []string{
		"{{.Store.GraphDriverName}}",
		"{{.Host.Storage.GraphDriverName}}",
	} {
		out, err := runner.RunWithOutput(ctx, "podman", "info", "--format", tmpl)
		if err == nil && strings.TrimSpace(out) != "" {
			driver := strings.TrimSpace(out)
			detail := "storage driver: " + driver
			status := StatusOK
			if !strings.Contains(strings.ToLower(driver), "overlay") {
				status = StatusWarn
				detail += " (expected overlay)"
			}
			return Check{Name: "storage driver", Status: status, Detail: detail}
		}
	}
	if _, err := runner.RunWithOutput(ctx, "podman", "info"); err != nil {
		return Check{Name: "podman info", Status: StatusFail, Detail: err.Error(), Fix: "lab podman configure && lab podman doctor"}
	}
	return Check{Name: "podman info", Status: StatusWarn, Detail: "info ok but could not parse storage driver"}
}

func checkUIDMap(ctx context.Context, runner exec.Runner) Check {
	out, err := runner.RunWithOutput(ctx, "podman", "unshare", "cat", "/proc/self/uid_map")
	if err != nil {
		return Check{
			Name:   "uid_map",
			Status: StatusFail,
			Detail: err.Error(),
			Fix:    "lab podman configure (subuid/subgid); on Ubuntu 24.04 also --allow-userns",
		}
	}
	if strings.TrimSpace(out) == "" {
		return Check{Name: "uid_map", Status: StatusFail, Detail: "empty uid_map", Fix: "lab podman configure"}
	}
	return Check{Name: "uid_map", Status: StatusOK, Detail: summarizeUIDMap(out)}
}

// summarizeUIDMap collapses multiline uid_map to a short detail string.
func summarizeUIDMap(out string) string {
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) == 0 {
		return "empty"
	}
	return strings.TrimSpace(lines[0])
}

func checkNetworkBackend(ctx context.Context, runner exec.Runner) Check {
	// Prefer pasta (passt); fall back to slirp4netns presence.
	pasta, _ := runner.RunWithOutput(ctx, "bash", "-c", "command -v pasta || command -v passt || true")
	slirp, _ := runner.RunWithOutput(ctx, "bash", "-c", "command -v slirp4netns || true")
	pasta = strings.TrimSpace(pasta)
	slirp = strings.TrimSpace(slirp)
	switch {
	case pasta != "":
		return Check{Name: "network backend", Status: StatusOK, Detail: "pasta/passt present: " + pasta}
	case slirp != "":
		return Check{Name: "network backend", Status: StatusWarn, Detail: "slirp4netns present (pasta preferred): " + slirp}
	default:
		return Check{
			Name:   "network backend",
			Status: StatusFail,
			Detail: "neither pasta/passt nor slirp4netns found",
			Fix:    "lab podman install (installs passt + slirp4netns on apt)",
		}
	}
}

func checkLinger(ctx context.Context, runner exec.Runner, username string) Check {
	if username == "" {
		return Check{Name: "linger", Status: StatusWarn, Detail: "username unknown"}
	}
	out, err := runner.RunWithOutput(ctx, "loginctl", "show-user", username, "--property=Linger")
	if err != nil {
		return Check{Name: "linger", Status: StatusFail, Detail: err.Error(), Fix: "lab podman configure"}
	}
	if parseLinger(out) {
		return Check{Name: "linger", Status: StatusOK, Detail: "Linger=yes"}
	}
	return Check{
		Name:   "linger",
		Status: StatusFail,
		Detail: "Linger not enabled — Quadlet units will not start at boot",
		Fix:    "lab podman configure",
	}
}

func checkQuadletDir(opts DoctorOpts) Check {
	path := filepath.Join(opts.HomeDir, ".config", "containers", "systemd")
	if _, err := opts.Stat(path); err != nil {
		return Check{
			Name:   "quadlet dir",
			Status: StatusFail,
			Detail: path + " missing",
			Fix:    "lab podman configure",
		}
	}
	return Check{Name: "quadlet dir", Status: StatusOK, Detail: path}
}

func checkRegistries(opts DoctorOpts) Check {
	path := filepath.Join(opts.HomeDir, ".config", "containers", "registries.conf")
	data, err := opts.ReadFile(path)
	if err != nil {
		return Check{
			Name:   "registries",
			Status: StatusFail,
			Detail: path + " missing",
			Fix:    "lab podman configure",
		}
	}
	if !strings.Contains(string(data), "unqualified-search-registries") {
		return Check{
			Name:   "registries",
			Status: StatusWarn,
			Detail: "file exists but unqualified-search-registries not set",
			Fix:    "lab podman configure",
		}
	}
	return Check{Name: "registries", Status: StatusOK, Detail: path}
}

func checkAppArmor(ctx context.Context, runner exec.Runner, info platform.Info) Check {
	if !needsAppArmorUserNSFix(info) {
		return Check{Name: "apparmor-userns", Status: StatusOK, Detail: "not applicable"}
	}
	out, err := runner.RunWithOutput(ctx, "sysctl", "-n", "kernel.apparmor_restrict_unprivileged_userns")
	if err != nil {
		return Check{Name: "apparmor-userns", Status: StatusWarn, Detail: "sysctl unavailable"}
	}
	val := strings.TrimSpace(out)
	if val == "0" {
		return Check{Name: "apparmor-userns", Status: StatusOK, Detail: "restrict_unprivileged_userns=0"}
	}
	return Check{
		Name:   "apparmor-userns",
		Status: StatusFail,
		Detail: fmt.Sprintf("restrict_unprivileged_userns=%s (breaks rootless)", val),
		Fix:    "lab podman configure --allow-userns",
	}
}

func checkSocket(ctx context.Context, runner exec.Runner) Check {
	out, err := runner.RunWithOutput(ctx, "systemctl", "--user", "is-active", "podman.socket")
	if err != nil {
		return Check{
			Name:   "podman.socket",
			Status: StatusWarn,
			Detail: "not active (optional; enable with lab podman configure --enable-socket)",
		}
	}
	state := strings.TrimSpace(out)
	if state == "active" {
		return Check{Name: "podman.socket", Status: StatusOK, Detail: "active"}
	}
	return Check{Name: "podman.socket", Status: StatusWarn, Detail: state}
}
