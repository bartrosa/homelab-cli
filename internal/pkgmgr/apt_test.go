package pkgmgr_test

import (
	"context"
	"strings"
	"testing"

	"github.com/bartrosa/homelab-cli/internal/pkgmgr"
	"github.com/stretchr/testify/require"
)

type recordingRunner struct {
	calls []string
}

func (r *recordingRunner) Run(_ context.Context, name string, args ...string) error {
	r.calls = append(r.calls, name+" "+strings.Join(args, " "))
	return nil
}

func (r *recordingRunner) RunWithOutput(ctx context.Context, name string, args ...string) (string, error) {
	return "", r.Run(ctx, name, args...)
}

func TestAPT_Install_buildsExpectedArgs(t *testing.T) {
	rec := &recordingRunner{}
	apt := &pkgmgr.APT{Runner: rec, Sudo: true}
	require.NoError(t, apt.Install(context.Background(), "git", "curl"))
	require.Len(t, rec.calls, 1)
	require.Contains(t, rec.calls[0], "apt-get install -y --no-install-recommends git curl")
}
