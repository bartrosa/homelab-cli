package bootstrap_test

import (
	"testing"

	"github.com/bartrosa/homelab-cli/internal/bootstrap"
	"github.com/stretchr/testify/require"
)

func TestFilterSections_only(t *testing.T) {
	got := bootstrap.SectionNamesForTest([]string{"cli-basics"}, nil)
	require.Equal(t, []string{"cli-basics"}, got)
}

func TestFilterSections_skip(t *testing.T) {
	got := bootstrap.SectionNamesForTest(nil, []string{"mise", "distrobox"})
	require.NotContains(t, got, "mise")
	require.NotContains(t, got, "distrobox")
	require.Contains(t, got, "cli-basics")
}

func TestPackageFor_apt(t *testing.T) {
	pkg, ok := bootstrap.PackageFor("fd", "apt")
	require.True(t, ok)
	require.Equal(t, "fd-find", pkg)
}

func TestParseCSV(t *testing.T) {
	require.Equal(t, []string{"docker", "mise"}, bootstrap.ParseCSV("docker,mise"))
}
