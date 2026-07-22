package podman_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/bartrosa/homelab-cli/internal/platform"
	"github.com/bartrosa/homelab-cli/internal/podman"
	"github.com/stretchr/testify/require"
)

func TestDoctor_versionFail(t *testing.T) {
	rec := &recordingRunner{
		errors: map[string]error{
			"podman --version": context.Canceled,
		},
	}
	report, err := podman.Doctor(context.Background(), rec, podman.DoctorOpts{
		Info:    platform.Info{GOOS: platform.OSDarwin, Packager: platform.PackagerBrew},
		HomeDir: t.TempDir(),
	})
	require.NoError(t, err)
	require.NotEmpty(t, report.Checks)
	require.Equal(t, podman.StatusFail, report.Checks[0].Status)
	require.Equal(t, "podman --version", report.Checks[0].Name)
}

func TestDoctor_linuxChecks(t *testing.T) {
	home := t.TempDir()
	quadlet := filepath.Join(home, ".config", "containers", "systemd")
	require.NoError(t, os.MkdirAll(quadlet, 0o755))
	reg := filepath.Join(home, ".config", "containers", "registries.conf")
	require.NoError(t, os.WriteFile(reg, []byte(`unqualified-search-registries = ["docker.io"]`+"\n"), 0o644))

	rec := &recordingRunner{
		//nolint:gosec // G101: test fixture keys are loginctl property names, not credentials
		outputs: map[string]string{
			"podman --version": "podman version 5.0.0",
			"podman info --format {{.Store.GraphDriverName}}":        "overlay",
			"podman unshare cat /proc/self/uid_map":                  "         0          0 4294967295",
			"bash -c command -v pasta || command -v passt || true":   "/usr/bin/pasta",
			"bash -c command -v slirp4netns || true":                 "",
			"loginctl show-user alice --property=Linger":             "Linger=yes",
			"systemctl --user is-active podman.socket":               "inactive",
		},
	}
	report, err := podman.Doctor(context.Background(), rec, podman.DoctorOpts{
		Info:     platform.Info{GOOS: platform.OSLinux, Packager: platform.PackagerDNF},
		HomeDir:  home,
		Username: "alice",
		ReadFile: os.ReadFile,
		Stat:     os.Stat,
	})
	require.NoError(t, err)

	byName := map[string]podman.Check{}
	for _, c := range report.Checks {
		byName[c.Name] = c
	}
	require.Equal(t, podman.StatusOK, byName["podman --version"].Status)
	require.Equal(t, podman.StatusOK, byName["storage driver"].Status)
	require.Equal(t, podman.StatusOK, byName["uid_map"].Status)
	require.Equal(t, podman.StatusOK, byName["network backend"].Status)
	require.Equal(t, podman.StatusOK, byName["linger"].Status)
	require.Equal(t, podman.StatusOK, byName["quadlet dir"].Status)
	require.Equal(t, podman.StatusOK, byName["registries"].Status)
	require.Equal(t, podman.StatusWarn, byName["podman.socket"].Status)
}

func TestDoctor_missingQuadlet(t *testing.T) {
	home := t.TempDir()
	rec := &recordingRunner{
		outputs: map[string]string{
			"podman --version": "podman version 5.0.0",
			"podman info --format {{.Store.GraphDriverName}}":      "overlay",
			"podman unshare cat /proc/self/uid_map":                "0 0 1",
			"bash -c command -v pasta || command -v passt || true": "",
			"bash -c command -v slirp4netns || true":               "/usr/bin/slirp4netns",
			"loginctl show-user alice --property=Linger":           "Linger=no",
			"systemctl --user is-active podman.socket":             "inactive",
		},
	}
	report, err := podman.Doctor(context.Background(), rec, podman.DoctorOpts{
		Info:     platform.Info{GOOS: platform.OSLinux, Packager: platform.PackagerAPT},
		HomeDir:  home,
		Username: "alice",
		ReadFile: os.ReadFile,
		Stat:     os.Stat,
	})
	require.NoError(t, err)
	byName := map[string]podman.Check{}
	for _, c := range report.Checks {
		byName[c.Name] = c
	}
	require.Equal(t, podman.StatusFail, byName["quadlet dir"].Status)
	require.Equal(t, podman.StatusFail, byName["linger"].Status)
	require.Equal(t, podman.StatusFail, byName["registries"].Status)
	require.Equal(t, podman.StatusWarn, byName["network backend"].Status)
}
