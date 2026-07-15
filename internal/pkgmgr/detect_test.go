package pkgmgr_test

import (
	"testing"

	"github.com/bartrosa/homelab-cli/internal/pkgmgr"
	"github.com/stretchr/testify/require"
)

func TestParseOSRelease_ubuntu(t *testing.T) {
	content := []byte(`NAME="Ubuntu"
ID=ubuntu
VERSION_ID="24.04"
`)
	osType, mgr := pkgmgr.ParseOSRelease(content)
	require.Equal(t, pkgmgr.OSUbuntu, osType)
	require.Equal(t, "apt", mgr)
}

func TestParseOSRelease_silverblue(t *testing.T) {
	content := []byte(`NAME="Fedora Linux"
ID=fedora
VARIANT_ID=silverblue
VERSION_ID=41
`)
	osType, mgr := pkgmgr.ParseOSRelease(content)
	require.Equal(t, pkgmgr.OSSilverblue, osType)
	require.Equal(t, "rpm-ostree", mgr)
}

func TestParseOSRelease_debian(t *testing.T) {
	content := []byte(`NAME="Debian GNU/Linux"
ID=debian
VERSION_ID="12"
`)
	osType, mgr := pkgmgr.ParseOSRelease(content)
	require.Equal(t, pkgmgr.OSType("debian"), osType)
	require.Equal(t, "apt", mgr)
}

func TestBuildAPTCommand(t *testing.T) {
	cmd := pkgmgr.BuildAPTCommand("git", "curl")
	require.Equal(t, []string{"apt-get", "install", "-y", "--no-install-recommends", "git", "curl"}, cmd)
}

func TestBuildRPMOstreeCommand(t *testing.T) {
	cmd := pkgmgr.BuildRPMOstreeCommand("distrobox")
	require.Equal(t, []string{"install", "--idempotent", "-y", "distrobox"}, cmd)
}
