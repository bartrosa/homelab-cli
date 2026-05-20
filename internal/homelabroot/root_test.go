package homelabroot_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bartrosa/homelab-cli/internal/homelabroot"
	"github.com/stretchr/testify/require"
)

func TestResolve_findsMarkerInTempDir(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "tools", "media"), 0o750))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "ml-stack"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "ml-stack", "podman-compose.yml"), []byte("version: '3'\n"), 0o644))

	got, err := homelabroot.Resolve(dir)
	require.NoError(t, err)
	require.Equal(t, dir, got)
}

func TestResolve_missing(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("LAB_HOMELAB_ROOT", "")
	_, err := homelabroot.Resolve(filepath.Join(home, "not-a-homelab-repo"))
	require.Error(t, err)
}
