package podman_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bartrosa/homelab-cli/internal/platform"
	"github.com/bartrosa/homelab-cli/internal/podman"
	"github.com/stretchr/testify/require"
)

func TestHasSubIDEntry_viaConfigureIdempotent(t *testing.T) {
	dir := t.TempDir()
	subuid := filepath.Join(dir, "subuid")
	subgid := filepath.Join(dir, "subgid")
	require.NoError(t, os.WriteFile(subuid, []byte("alice:100000:65536\n"), 0o644))
	require.NoError(t, os.WriteFile(subgid, []byte("alice:100000:65536\n"), 0o644))

	rec := &recordingRunner{
		outputs: map[string]string{
			"loginctl show-user alice --property=Linger": "Linger=yes",
		},
	}
	home := t.TempDir()
	results, err := podman.Configure(context.Background(), rec, podman.ConfigureOpts{
		Info:     platform.Info{GOOS: platform.OSLinux, Packager: platform.PackagerDNF},
		Username: "alice",
		HomeDir:  home,
		ReadFile: func(name string) ([]byte, error) {
			switch name {
			case "/etc/subuid":
				return os.ReadFile(subuid)
			case "/etc/subgid":
				return os.ReadFile(subgid)
			default:
				return os.ReadFile(name)
			}
		},
		WriteFile: os.WriteFile,
		MkdirAll:  os.MkdirAll,
		Stat:      os.Stat,
	})
	require.NoError(t, err)

	var sub StepFinder = results
	r := sub.find("subuid/subgid")
	require.False(t, r.Changed)
	require.Contains(t, r.Detail, "already configured")

	// No usermod calls when already mapped.
	for _, c := range rec.calls {
		require.NotContains(t, c, "usermod")
	}
}

type StepFinder []podman.StepResult

func (s StepFinder) find(name string) podman.StepResult {
	for _, r := range s {
		if r.Name == name {
			return r
		}
	}
	return podman.StepResult{}
}

func TestConfigure_dryRun_subuid(t *testing.T) {
	rec := &recordingRunner{
		outputs: map[string]string{
			"loginctl show-user bob --property=Linger": "Linger=no",
		},
	}
	home := t.TempDir()
	results, err := podman.Configure(context.Background(), rec, podman.ConfigureOpts{
		Info:     platform.Info{GOOS: platform.OSLinux, Packager: platform.PackagerAPT},
		DryRun:   true,
		Username: "bob",
		HomeDir:  home,
		ReadFile: func(name string) ([]byte, error) {
			if name == "/etc/subuid" || name == "/etc/subgid" {
				return []byte(""), nil
			}
			return nil, os.ErrNotExist
		},
		WriteFile: os.WriteFile,
		MkdirAll:  os.MkdirAll,
		Stat:      os.Stat,
	})
	require.NoError(t, err)
	// Dry-run may probe state (loginctl/sysctl) but must not mutate.
	for _, c := range rec.calls {
		require.NotContains(t, c, "usermod")
		require.NotContains(t, c, "enable-linger")
		require.NotContains(t, c, "sysctl --system")
	}
	r := StepFinder(results).find("subuid/subgid")
	require.True(t, r.Changed)
	require.Contains(t, r.Detail, "would run")
}

func TestConfigure_macos_skipsLinux(t *testing.T) {
	rec := &recordingRunner{}
	results, err := podman.Configure(context.Background(), rec, podman.ConfigureOpts{
		Info: platform.Info{GOOS: platform.OSDarwin, Packager: platform.PackagerBrew},
	})
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.True(t, results[0].Skipped)
	require.Empty(t, rec.calls)
}

func TestConfigure_registries_idempotent(t *testing.T) {
	home := t.TempDir()
	dir := filepath.Join(home, ".config", "containers")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	path := filepath.Join(dir, "registries.conf")
	require.NoError(t, os.WriteFile(path, []byte(`unqualified-search-registries = ["docker.io"]`+"\n"), 0o644))

	rec := &recordingRunner{
		outputs: map[string]string{
			"loginctl show-user alice --property=Linger": "Linger=yes",
		},
	}
	results, err := podman.Configure(context.Background(), rec, podman.ConfigureOpts{
		Info:     platform.Info{GOOS: platform.OSLinux, Packager: platform.PackagerDNF},
		Username: "alice",
		HomeDir:  home,
		ReadFile: func(name string) ([]byte, error) {
			if name == "/etc/subuid" || name == "/etc/subgid" {
				return []byte("alice:100000:65536\n"), nil
			}
			return os.ReadFile(name)
		},
		WriteFile: os.WriteFile,
		MkdirAll:  os.MkdirAll,
		Stat:      os.Stat,
	})
	require.NoError(t, err)
	r := StepFinder(results).find("registries")
	require.False(t, r.Changed)
}

func TestConfigure_apparmor_requiresFlag(t *testing.T) {
	// Force Ubuntu path via monkeying needsAppArmor — we test the gate when sysctl returns 1.
	// needsAppArmorUserNSFix reads real /etc/os-release; skip if not Ubuntu.
	id, ok := platform.ReadOSReleaseKey("ID")
	if !ok || id != "ubuntu" {
		t.Skip("apparmor flag test requires Ubuntu host for needsAppArmorUserNSFix")
	}

	rec := &recordingRunner{
		outputs: map[string]string{
			"sysctl -n kernel.apparmor_restrict_unprivileged_userns": "1",
			"loginctl show-user alice --property=Linger":             "Linger=yes",
		},
	}
	home := t.TempDir()
	results, err := podman.Configure(context.Background(), rec, podman.ConfigureOpts{
		Info:        platform.Info{GOOS: platform.OSLinux, Packager: platform.PackagerAPT},
		AllowUserNS: false,
		Username:    "alice",
		HomeDir:     home,
		ReadFile: func(name string) ([]byte, error) {
			if name == "/etc/subuid" || name == "/etc/subgid" {
				return []byte("alice:100000:65536\n"), nil
			}
			return nil, os.ErrNotExist
		},
		WriteFile: os.WriteFile,
		MkdirAll:  os.MkdirAll,
		Stat: func(string) (os.FileInfo, error) {
			return fakeFileInfo{}, nil
		},
	})
	require.NoError(t, err)
	r := StepFinder(results).find("apparmor-userns")
	require.False(t, r.Changed)
	require.Contains(t, r.Detail, "--allow-userns")
	for _, c := range rec.calls {
		require.NotContains(t, c, "sysctl --system")
	}
}

type fakeFileInfo struct{}

func (fakeFileInfo) Name() string       { return "systemd" }
func (fakeFileInfo) Size() int64        { return 0 }
func (fakeFileInfo) Mode() os.FileMode  { return 0o755 }
func (fakeFileInfo) ModTime() time.Time { return time.Time{} }
func (fakeFileInfo) IsDir() bool        { return true }
func (fakeFileInfo) Sys() any           { return nil }

func TestRegistriesConfContent(t *testing.T) {
	// Exercise via configure write path
	home := t.TempDir()
	rec := &recordingRunner{
		outputs: map[string]string{
			"loginctl show-user alice --property=Linger": "Linger=yes",
		},
	}
	_, err := podman.Configure(context.Background(), rec, podman.ConfigureOpts{
		Info:       platform.Info{GOOS: platform.OSLinux, Packager: platform.PackagerDNF},
		Username:   "alice",
		HomeDir:    home,
		Registries: []string{"docker.io", "ghcr.io"},
		ReadFile: func(name string) ([]byte, error) {
			if name == "/etc/subuid" || name == "/etc/subgid" {
				return []byte("alice:100000:65536\n"), nil
			}
			return nil, os.ErrNotExist
		},
		WriteFile: os.WriteFile,
		MkdirAll:  os.MkdirAll,
		Stat:      os.Stat,
	})
	require.NoError(t, err)
	data, err := os.ReadFile(filepath.Join(home, ".config", "containers", "registries.conf"))
	require.NoError(t, err)
	require.Contains(t, string(data), "docker.io")
	require.Contains(t, string(data), "ghcr.io")
	require.Contains(t, string(data), "unqualified-search-registries")
}
