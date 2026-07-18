package tofu_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/bartrosa/homelab-cli/internal/platform"
	"github.com/bartrosa/homelab-cli/internal/tofu"
	"github.com/stretchr/testify/require"
)

func TestDetectMethod_brew(t *testing.T) {
	m, err := tofu.DetectMethod(platform.Info{GOOS: platform.OSDarwin, Packager: platform.PackagerBrew})
	require.NoError(t, err)
	require.Equal(t, tofu.MethodBrew, m.Kind)
	require.Contains(t, m.Reason, "Homebrew")
}

func TestDetectMethod_linuxScript(t *testing.T) {
	cases := []struct {
		packager string
		script   string
	}{
		{platform.PackagerAPT, "deb"},
		{platform.PackagerDNF, "rpm"},
	}
	for _, tc := range cases {
		m, err := tofu.DetectMethod(platform.Info{GOOS: platform.OSLinux, Packager: tc.packager})
		require.NoError(t, err)
		require.Equal(t, tofu.MethodScript, m.Kind)
		require.Equal(t, tc.script, m.ScriptMethod)
		cmds := tofu.PlannedCommands(m, false)
		require.Len(t, cmds, 1)
		require.Contains(t, cmds[0][2], "--install-method "+tc.script)
		require.NotContains(t, cmds[0][2], "--install-method auto")
	}
}

func TestDetectMethod_linuxSnapFallback(t *testing.T) {
	m, err := tofu.DetectMethodWith(
		platform.Info{GOOS: platform.OSLinux, Packager: platform.PackagerUnknown},
		func(name string) bool { return name == "snap" },
	)
	require.NoError(t, err)
	require.Equal(t, tofu.MethodSnap, m.Kind)
	cmds := tofu.PlannedCommands(m, false)
	require.Equal(t, []string{"sudo", "snap", "install", "--classic", "opentofu"}, cmds[0])
}

func TestDetectMethod_windowsWinget(t *testing.T) {
	m, err := tofu.DetectMethodWith(
		platform.Info{GOOS: "windows"},
		func(name string) bool { return name == "winget" },
	)
	require.NoError(t, err)
	require.Equal(t, tofu.MethodWinget, m.Kind)
	cmds := tofu.PlannedCommands(m, false)
	require.Equal(t, []string{"winget", "install", "--id", "OpenTofu.OpenTofu"}, cmds[0])
	cmdsUp := tofu.PlannedCommands(m, true)
	require.Equal(t, []string{"winget", "upgrade", "--id", "OpenTofu.OpenTofu"}, cmdsUp[0])
}

func TestDetectMethod_unsupported(t *testing.T) {
	_, err := tofu.DetectMethod(platform.Info{GOOS: "plan9"})
	require.Error(t, err)
}

func TestInstall_dryRun_doesNotCallRunner(t *testing.T) {
	rec := &recordingRunner{}
	var out bytes.Buffer
	err := tofu.Install(context.Background(), rec, tofu.InstallOpts{
		Info:   platform.Info{GOOS: platform.OSDarwin, Packager: platform.PackagerBrew},
		DryRun: true,
		Stdout: &out,
	})
	require.NoError(t, err)
	require.Empty(t, rec.calls)
	require.Contains(t, out.String(), "method: brew")
	require.Contains(t, out.String(), "brew install opentofu")
}

func TestUpgrade_dryRun_brew(t *testing.T) {
	rec := &recordingRunner{}
	var out bytes.Buffer
	err := tofu.Upgrade(context.Background(), rec, tofu.InstallOpts{
		Info:   platform.Info{GOOS: platform.OSDarwin, Packager: platform.PackagerBrew},
		DryRun: true,
		Stdout: &out,
	})
	require.NoError(t, err)
	require.Empty(t, rec.calls)
	require.Contains(t, out.String(), "brew upgrade opentofu")
}

func TestInstall_brew_runsInstall(t *testing.T) {
	rec := &recordingRunner{}
	err := tofu.Install(context.Background(), rec, tofu.InstallOpts{
		Info: platform.Info{GOOS: platform.OSDarwin, Packager: platform.PackagerBrew},
	})
	require.NoError(t, err)
	require.Equal(t, []string{"brew install opentofu"}, rec.calls)
}

type recordingRunner struct {
	calls []string
}

func (r *recordingRunner) Run(_ context.Context, name string, args ...string) error {
	r.calls = append(r.calls, name+" "+strings.Join(args, " "))
	return nil
}

func (r *recordingRunner) RunWithOutput(ctx context.Context, name string, args ...string) (string, error) {
	_ = r.Run(ctx, name, args...)
	return "OpenTofu v1.9.0", nil
}
