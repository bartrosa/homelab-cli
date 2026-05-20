package media_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/bartrosa/homelab-cli/internal/media"
	"github.com/stretchr/testify/require"
)

func TestConvertHEIC_noFiles(t *testing.T) {
	dir := t.TempDir()
	_, err := media.ConvertHEIC(context.Background(), nil, nil, media.HEICOptions{Dir: dir})
	require.Error(t, err)
	require.Contains(t, err.Error(), "no .HEIC")
}

func TestConvertHEIC_notADirectory(t *testing.T) {
	f := filepath.Join(t.TempDir(), "x.txt")
	require.NoError(t, os.WriteFile(f, []byte("x"), 0o644))
	_, err := media.ConvertHEIC(context.Background(), nil, nil, media.HEICOptions{Dir: f})
	require.Error(t, err)
	require.Contains(t, err.Error(), "not a directory")
}

func TestConvertHEIC_dryRun_skipsExistingJPG(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "photo.HEIC"), []byte{0}, 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "photo.HEIC.jpg"), []byte{0}, 0o644))

	var stderr bytes.Buffer
	n, err := media.ConvertHEIC(context.Background(), nil, &stderr, media.HEICOptions{
		Dir: dir, DryRun: true,
	})
	require.NoError(t, err)
	require.Equal(t, 0, n)
	require.Contains(t, stderr.String(), "skip")
}
