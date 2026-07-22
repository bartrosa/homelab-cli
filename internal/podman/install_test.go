package podman_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/bartrosa/homelab-cli/internal/platform"
	"github.com/bartrosa/homelab-cli/internal/podman"
	"github.com/stretchr/testify/require"
)

type recordingRunner struct {
	calls   []string
	outputs map[string]string
	errors  map[string]error
}

func (r *recordingRunner) Run(_ context.Context, name string, args ...string) error {
	key := name + " " + strings.Join(args, " ")
	r.calls = append(r.calls, key)
	if r.errors != nil {
		if err, ok := r.errors[key]; ok {
			return err
		}
	}
	return nil
}

func (r *recordingRunner) RunWithOutput(_ context.Context, name string, args ...string) (string, error) {
	key := name + " " + strings.Join(args, " ")
	r.calls = append(r.calls, key)
	if r.errors != nil {
		if err, ok := r.errors[key]; ok {
			return "", err
		}
	}
	if r.outputs != nil {
		if out, ok := r.outputs[key]; ok {
			return out, nil
		}
	}
	return "", nil
}

func TestDetectMethod_apt(t *testing.T) {
	m, err := podman.DetectMethod(platform.Info{GOOS: platform.OSLinux, Packager: platform.PackagerAPT})
	require.NoError(t, err)
	require.Equal(t, podman.MethodAPT, m.Kind)
	cmds := podman.PlannedInstallCommands(m)
	require.Equal(t, []string{"sudo", "apt-get", "update"}, cmds[0])
	require.Contains(t, strings.Join(cmds[1], " "), "podman")
	require.Contains(t, strings.Join(cmds[1], " "), "passt")
}

func TestDetectMethod_dnf(t *testing.T) {
	m, err := podman.DetectMethod(platform.Info{GOOS: platform.OSLinux, Packager: platform.PackagerDNF})
	require.NoError(t, err)
	require.Equal(t, podman.MethodDNF, m.Kind)
}

func TestDetectMethod_silverblue(t *testing.T) {
	m, err := podman.DetectMethod(platform.Info{
		GOOS:         platform.OSLinux,
		Packager:     platform.PackagerDNF,
		IsSilverblue: true,
	})
	require.NoError(t, err)
	require.Equal(t, podman.MethodSilverblue, m.Kind)
	require.Nil(t, podman.PlannedInstallCommands(m))
}

func TestDetectMethod_brew(t *testing.T) {
	m, err := podman.DetectMethod(platform.Info{GOOS: platform.OSDarwin, Packager: platform.PackagerBrew})
	require.NoError(t, err)
	require.Equal(t, podman.MethodBrew, m.Kind)
	cmds := podman.PlannedInstallCommands(m)
	require.Equal(t, []string{"brew", "install", "podman"}, cmds[0])
	require.Equal(t, []string{"podman", "machine", "init"}, cmds[1])
}

func TestInstall_dryRun_doesNotCallRunner(t *testing.T) {
	rec := &recordingRunner{}
	var out bytes.Buffer
	err := podman.Install(context.Background(), rec, podman.InstallOpts{
		Info:   platform.Info{GOOS: platform.OSLinux, Packager: platform.PackagerAPT},
		DryRun: true,
		Stdout: &out,
	})
	require.NoError(t, err)
	require.Empty(t, rec.calls)
	require.Contains(t, out.String(), "apt-get")
}

func TestInstall_apt_runsUpdateAndInstall(t *testing.T) {
	rec := &recordingRunner{}
	err := podman.Install(context.Background(), rec, podman.InstallOpts{
		Info: platform.Info{GOOS: platform.OSLinux, Packager: platform.PackagerAPT},
	})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(rec.calls), 2)
	require.Contains(t, rec.calls[0], "apt-get update")
	require.Contains(t, rec.calls[1], "apt-get install")
}

func TestUpgrade_dryRun(t *testing.T) {
	rec := &recordingRunner{}
	var out bytes.Buffer
	err := podman.Upgrade(context.Background(), rec, podman.InstallOpts{
		Info:   platform.Info{GOOS: platform.OSLinux, Packager: platform.PackagerDNF},
		DryRun: true,
		Stdout: &out,
	})
	require.NoError(t, err)
	require.Empty(t, rec.calls)
	require.Contains(t, out.String(), "dnf upgrade")
}

func TestRemove_silverblue_refused(t *testing.T) {
	rec := &recordingRunner{}
	err := podman.Remove(context.Background(), rec, podman.InstallOpts{
		Info: platform.Info{GOOS: platform.OSLinux, Packager: platform.PackagerDNF, IsSilverblue: true},
	})
	require.Error(t, err)
	require.Empty(t, rec.calls)
}
