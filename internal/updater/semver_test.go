package updater_test

import (
	"testing"

	"github.com/bartrosa/homelab-cli/internal/updater"
	"github.com/stretchr/testify/require"
)

func TestCompareVersions_order(t *testing.T) {
	require.Equal(t, -1, updater.CompareVersions("v0.1.0", "v0.2.0"))
	require.Equal(t, 1, updater.CompareVersions("0.3.0", "0.2.0"))
	require.Equal(t, 0, updater.CompareVersions("v1.0.0", "1.0.0"))
}

func TestCompareVersions_dev(t *testing.T) {
	require.Equal(t, -1, updater.CompareVersions("dev", "v0.1.0"))
	require.Equal(t, 1, updater.CompareVersions("v0.1.0", "dev"))
}

func TestCompareVersions_prerelease(t *testing.T) {
	require.Equal(t, -1, updater.CompareVersions("v1.0.0-rc1", "v1.0.0"))
	require.Equal(t, 1, updater.CompareVersions("v1.0.0", "v1.0.0-rc1"))
}

func TestIsNewer(t *testing.T) {
	require.True(t, updater.IsNewer("v0.1.0", "v0.2.0"))
	require.False(t, updater.IsNewer("v0.2.0", "v0.2.0"))
}
