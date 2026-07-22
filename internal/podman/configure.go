package podman

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/bartrosa/homelab-cli/internal/exec"
	"github.com/bartrosa/homelab-cli/internal/platform"
)

// DefaultRegistries used for unqualified short-name pulls.
// TODO: make Forgejo (self-hosted) registry the first search entry once available.
var DefaultRegistries = []string{"docker.io", "ghcr.io", "quay.io"}

// StepResult describes one configure step outcome.
type StepResult struct {
	Name    string
	Changed bool
	Detail  string
	Skipped bool
}

// ConfigureOpts controls post-install configuration.
type ConfigureOpts struct {
	Info         platform.Info
	DryRun       bool
	AllowUserNS  bool // explicit consent for Ubuntu AppArmor userns sysctl
	EnableSocket bool
	Registries   []string // unqualified-search-registries; defaults to DefaultRegistries
	Stdout       io.Writer
	Stderr       io.Writer

	// Injectable for tests (nil = real OS).
	Username  string
	HomeDir   string
	ReadFile  func(name string) ([]byte, error)
	WriteFile func(name string, data []byte, perm os.FileMode) error
	MkdirAll  func(path string, perm os.FileMode) error
	Stat      func(name string) (os.FileInfo, error)
}

func (o *ConfigureOpts) normalize() {
	if o.ReadFile == nil {
		o.ReadFile = os.ReadFile
	}
	if o.WriteFile == nil {
		o.WriteFile = os.WriteFile
	}
	if o.MkdirAll == nil {
		o.MkdirAll = os.MkdirAll
	}
	if o.Stat == nil {
		o.Stat = os.Stat
	}
	if len(o.Registries) == 0 {
		o.Registries = append([]string{}, DefaultRegistries...)
	}
	if o.Username == "" {
		if u, err := user.Current(); err == nil {
			o.Username = u.Username
		} else {
			o.Username = os.Getenv("USER")
		}
	}
	if o.HomeDir == "" {
		if h, err := os.UserHomeDir(); err == nil {
			o.HomeDir = h
		}
	}
	if o.Info.GOOS == "" {
		o.Info = platform.Detect()
	}
}

// Configure runs all post-install steps that apply to this OS.
// On macOS, Linux rootless/Quadlet steps are skipped (Podman runs in a VM).
func Configure(ctx context.Context, runner exec.Runner, opts ConfigureOpts) ([]StepResult, error) {
	opts.normalize()
	var results []StepResult

	if opts.Info.GOOS == platform.OSDarwin {
		results = append(results, StepResult{
			Name:    "linux-rootless",
			Skipped: true,
			Detail:  "macOS uses podman machine (VM); skipping subuid, linger, Quadlet, AppArmor",
		})
		return results, nil
	}
	if opts.Info.GOOS != platform.OSLinux {
		results = append(results, StepResult{
			Name:    "configure",
			Skipped: true,
			Detail:  fmt.Sprintf("unsupported OS %s", opts.Info.GOOS),
		})
		return results, nil
	}

	steps := []func(context.Context, exec.Runner, *ConfigureOpts) (StepResult, error){
		ensureSubIDs,
		ensureAppArmorUserNS,
		ensureLinger,
		ensureQuadletDirs,
		ensureRegistries,
		ensureSocket,
	}
	for _, step := range steps {
		r, err := step(ctx, runner, &opts)
		results = append(results, r)
		if err != nil {
			return results, err
		}
	}
	return results, nil
}

// ensureSubIDs adds /etc/subuid and /etc/subgid mappings for rootless Podman.
// Requires sudo (usermod writes system files).
func ensureSubIDs(ctx context.Context, runner exec.Runner, opts *ConfigureOpts) (StepResult, error) {
	name := "subuid/subgid"
	user := opts.Username
	if user == "" {
		return StepResult{Name: name, Detail: "cannot determine username"}, fmt.Errorf("username empty")
	}
	hasUID, err := hasSubIDEntry(opts.ReadFile, "/etc/subuid", user)
	if err != nil {
		return StepResult{Name: name, Detail: err.Error()}, err
	}
	hasGID, err := hasSubIDEntry(opts.ReadFile, "/etc/subgid", user)
	if err != nil {
		return StepResult{Name: name, Detail: err.Error()}, err
	}
	if hasUID && hasGID {
		return StepResult{Name: name, Changed: false, Detail: "already configured"}, nil
	}
	if opts.DryRun {
		return StepResult{Name: name, Changed: true, Detail: "would run: sudo usermod --add-subuids/--add-subgids"}, nil
	}
	if !hasUID {
		if err := runner.Run(ctx, "sudo", "usermod", "--add-subuids", "100000-165535", user); err != nil {
			return StepResult{Name: name, Detail: err.Error()}, err
		}
	}
	if !hasGID {
		if err := runner.Run(ctx, "sudo", "usermod", "--add-subgids", "100000-165535", user); err != nil {
			return StepResult{Name: name, Detail: err.Error()}, err
		}
	}
	return StepResult{Name: name, Changed: true, Detail: "configured subuid/subgid for " + user}, nil
}

// hasSubIDEntry reports whether user has a mapping line in a subuid/subgid file.
func hasSubIDEntry(readFile func(string) ([]byte, error), path, username string) (bool, error) {
	data, err := readFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	prefix := username + ":"
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), prefix) {
			return true, nil
		}
	}
	return false, nil
}

// ensureAppArmorUserNS addresses Ubuntu >= 23.10 blocking unprivileged user namespaces.
// Requires sudo when applying. Never automatic — needs AllowUserNS.
func ensureAppArmorUserNS(ctx context.Context, runner exec.Runner, opts *ConfigureOpts) (StepResult, error) {
	name := "apparmor-userns"
	if !needsAppArmorUserNSFix(opts.Info) {
		return StepResult{Name: name, Skipped: true, Detail: "not Ubuntu >= 23.10"}, nil
	}
	out, err := runner.RunWithOutput(ctx, "sysctl", "-n", "kernel.apparmor_restrict_unprivileged_userns")
	if err != nil {
		// sysctl key may be missing on some kernels
		return StepResult{Name: name, Skipped: true, Detail: "sysctl key unavailable: " + err.Error()}, nil
	}
	val := strings.TrimSpace(out)
	if val == "0" {
		return StepResult{Name: name, Changed: false, Detail: "already 0 (rootless userns allowed)"}, nil
	}
	if val != "1" {
		return StepResult{Name: name, Skipped: true, Detail: "unexpected value: " + val}, nil
	}

	warn := "Ubuntu restricts unprivileged user namespaces (apparmor_restrict_unprivileged_userns=1), which breaks rootless Podman. Setting it to 0 weakens isolation for unprivileged processes. Pass --allow-userns to apply."
	if !opts.AllowUserNS {
		return StepResult{Name: name, Changed: false, Detail: warn}, nil
	}
	if opts.DryRun {
		return StepResult{Name: name, Changed: true, Detail: "would write /etc/sysctl.d/99-podman-rootless.conf and sysctl --system"}, nil
	}
	conf := "kernel.apparmor_restrict_unprivileged_userns=0\n"
	// Write via sudo tee (system path).
	script := fmt.Sprintf(`printf '%s' | sudo tee /etc/sysctl.d/99-podman-rootless.conf >/dev/null && sudo sysctl --system`, conf)
	if err := runner.Run(ctx, "bash", "-c", script); err != nil {
		return StepResult{Name: name, Detail: err.Error()}, err
	}
	return StepResult{Name: name, Changed: true, Detail: "set apparmor_restrict_unprivileged_userns=0"}, nil
}

// readOSReleaseKey is platform.ReadOSReleaseKey; overridden in tests.
var readOSReleaseKey = platform.ReadOSReleaseKey

// needsAppArmorUserNSFix is true for Ubuntu 23.10+.
func needsAppArmorUserNSFix(info platform.Info) bool {
	if info.GOOS != platform.OSLinux || info.Packager != platform.PackagerAPT {
		return false
	}
	id, ok := readOSReleaseKey("ID")
	if !ok || id != "ubuntu" {
		return false
	}
	ver, ok := readOSReleaseKey("VERSION_ID")
	if !ok {
		return false
	}
	return ubuntuAtLeast(ver, 23, 10)
}

// ubuntuAtLeast parses VERSION_ID like "24.04" and compares to major.minor.
func ubuntuAtLeast(versionID string, major, minor int) bool {
	parts := strings.SplitN(strings.TrimSpace(versionID), ".", 3)
	if len(parts) < 2 {
		return false
	}
	var maj, mino int
	if _, err := fmt.Sscanf(parts[0], "%d", &maj); err != nil {
		return false
	}
	if _, err := fmt.Sscanf(parts[1], "%d", &mino); err != nil {
		return false
	}
	if maj > major {
		return true
	}
	if maj < major {
		return false
	}
	return mino >= minor
}

// ensureLinger enables systemd user lingering so Quadlet units start at boot
// without an interactive login (power-loss recovery). Requires sudo.
func ensureLinger(ctx context.Context, runner exec.Runner, opts *ConfigureOpts) (StepResult, error) {
	name := "linger"
	user := opts.Username
	out, err := runner.RunWithOutput(ctx, "loginctl", "show-user", user, "--property=Linger")
	if err != nil {
		return StepResult{Name: name, Detail: err.Error()}, err
	}
	if parseLinger(out) {
		return StepResult{Name: name, Changed: false, Detail: "already enabled"}, nil
	}
	if opts.DryRun {
		return StepResult{Name: name, Changed: true, Detail: "would run: sudo loginctl enable-linger " + user}, nil
	}
	if err := runner.Run(ctx, "sudo", "loginctl", "enable-linger", user); err != nil {
		return StepResult{Name: name, Detail: err.Error()}, err
	}
	return StepResult{Name: name, Changed: true, Detail: "enabled linger for " + user}, nil
}

// parseLinger returns true when loginctl show-user output indicates Linger=yes.
func parseLinger(out string) bool {
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if strings.EqualFold(line, "Linger=yes") {
			return true
		}
	}
	return false
}

// ensureQuadletDirs creates rootless and documents root Quadlet unit directories.
func ensureQuadletDirs(_ context.Context, _ exec.Runner, opts *ConfigureOpts) (StepResult, error) {
	name := "quadlet-dirs"
	rootless := filepath.Join(opts.HomeDir, ".config", "containers", "systemd")
	_, err := opts.Stat(rootless)
	exists := err == nil
	if exists {
		return StepResult{Name: name, Changed: false, Detail: "already exists: " + rootless}, nil
	}
	if opts.DryRun {
		return StepResult{Name: name, Changed: true, Detail: "would mkdir " + rootless}, nil
	}
	if err := opts.MkdirAll(rootless, 0o755); err != nil {
		return StepResult{Name: name, Detail: err.Error()}, err
	}
	return StepResult{Name: name, Changed: true, Detail: "created " + rootless + " (root Quadlet path: /etc/containers/systemd/)"}, nil
}

// ensureRegistries writes rootless registries.conf with unqualified-search-registries.
func ensureRegistries(_ context.Context, _ exec.Runner, opts *ConfigureOpts) (StepResult, error) {
	name := "registries"
	dir := filepath.Join(opts.HomeDir, ".config", "containers")
	path := filepath.Join(dir, "registries.conf")
	want := registriesConfContent(opts.Registries)

	if data, err := opts.ReadFile(path); err == nil {
		if strings.Contains(string(data), "unqualified-search-registries") {
			return StepResult{Name: name, Changed: false, Detail: "already configured: " + path}, nil
		}
	}
	if opts.DryRun {
		return StepResult{Name: name, Changed: true, Detail: "would write " + path}, nil
	}
	if err := opts.MkdirAll(dir, 0o755); err != nil {
		return StepResult{Name: name, Detail: err.Error()}, err
	}
	if err := opts.WriteFile(path, []byte(want), 0o644); err != nil {
		return StepResult{Name: name, Detail: err.Error()}, err
	}
	return StepResult{Name: name, Changed: true, Detail: "wrote " + path}, nil
}

// registriesConfContent builds a containers registries.conf snippet.
func registriesConfContent(registries []string) string {
	quoted := make([]string, 0, len(registries))
	for _, r := range registries {
		quoted = append(quoted, fmt.Sprintf("%q", r))
	}
	// TODO: prepend self-hosted Forgejo registry once deployed.
	return fmt.Sprintf("# Managed by lab podman configure\nunqualified-search-registries = [%s]\n", strings.Join(quoted, ", "))
}

// ensureSocket optionally enables the user podman.socket for Docker API clients.
func ensureSocket(ctx context.Context, runner exec.Runner, opts *ConfigureOpts) (StepResult, error) {
	name := "podman.socket"
	if !opts.EnableSocket {
		return StepResult{Name: name, Skipped: true, Detail: "skipped (pass --enable-socket)"}, nil
	}
	if opts.DryRun {
		return StepResult{Name: name, Changed: true, Detail: "would run: systemctl --user enable --now podman.socket"}, nil
	}
	if err := runner.Run(ctx, "systemctl", "--user", "enable", "--now", "podman.socket"); err != nil {
		return StepResult{Name: name, Detail: err.Error()}, err
	}
	return StepResult{Name: name, Changed: true, Detail: "enabled and started podman.socket"}, nil
}

// PurgeUserData removes rootless container storage and config (destructive).
func PurgeUserData(opts ConfigureOpts) error {
	opts.normalize()
	paths := []string{
		filepath.Join(opts.HomeDir, ".local", "share", "containers"),
		filepath.Join(opts.HomeDir, ".config", "containers"),
	}
	if opts.DryRun {
		if opts.Stdout != nil {
			for _, p := range paths {
				_, _ = fmt.Fprintln(opts.Stdout, "would remove:", p)
			}
		}
		return nil
	}
	for _, p := range paths {
		if err := os.RemoveAll(p); err != nil {
			return fmt.Errorf("remove %s: %w", p, err)
		}
	}
	return nil
}
