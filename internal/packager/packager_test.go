package packager_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/bartrosa/homelab-cli/internal/packager"
	"github.com/stretchr/testify/require"
)

func TestInstall_dryRun(t *testing.T) {
	var out, errOut bytes.Buffer
	mgr := packager.New(&out, &errOut, true)
	err := mgr.Install(context.Background(), "ripgrep")
	require.NoError(t, err)
	require.Contains(t, errOut.String(), "[dry-run]")
}

func TestInstall_emptyName(t *testing.T) {
	mgr := packager.New(nil, nil, false)
	err := mgr.Install(context.Background(), "  ")
	require.Error(t, err)
}
