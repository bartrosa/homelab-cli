package iso

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListCachedImages(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.iso"), []byte("x"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "skip.txt"), []byte("x"), 0o644))

	imgs, err := ListCachedImages(dir)
	require.NoError(t, err)
	require.Len(t, imgs, 1)
	require.Equal(t, "a.iso", imgs[0].Name)
}

func TestResolveISORef_byPath(t *testing.T) {
	dir := t.TempDir()
	iso := filepath.Join(dir, "test.iso")
	require.NoError(t, os.WriteFile(iso, []byte("data"), 0o644))

	got, err := ResolveISORef(iso, dir)
	require.NoError(t, err)
	require.Equal(t, iso, got)
}

func TestResolveISORef_byFragment(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "ubuntu-24.04.4-desktop-amd64.iso"), []byte("data"), 0o644))

	got, err := ResolveISORef("24.04.4-desktop", dir)
	require.NoError(t, err)
	require.Contains(t, got, "ubuntu-24.04.4-desktop-amd64.iso")
}
