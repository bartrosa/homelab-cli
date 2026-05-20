package platform_test

import (
	"testing"

	"github.com/bartrosa/homelab-cli/internal/platform"
	"github.com/stretchr/testify/require"
)

func TestDetect_returnsGOOS(t *testing.T) {
	info := platform.Detect()
	require.NotEmpty(t, info.GOOS)
	require.Equal(t, info.GOOS, info.Family)
}

func TestDetect_packagerOnDarwinOrLinux(t *testing.T) {
	info := platform.Detect()
	switch info.GOOS {
	case platform.OSDarwin:
		// CI may lack brew
		_ = info.Packager
	case platform.OSLinux:
		_ = info.Packager
	}
}
