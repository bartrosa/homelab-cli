package bootstrap_test

import (
	"testing"

	"github.com/bartrosa/homelab-cli/internal/bootstrap"
	"github.com/stretchr/testify/require"
)

func TestLoadEmbedded_laptopMacos(t *testing.T) {
	p, err := bootstrap.LoadEmbedded("laptop-macos")
	require.NoError(t, err)
	require.Equal(t, "laptop-macos", p.Name)
	require.NotEmpty(t, p.Steps)
}

func TestListEmbedded_includesProfiles(t *testing.T) {
	names, err := bootstrap.ListEmbedded()
	require.NoError(t, err)
	require.Contains(t, names, "server-ubuntu")
	require.Contains(t, names, "laptop-macos")
}

func TestLoadEmbedded_unknown(t *testing.T) {
	_, err := bootstrap.LoadEmbedded("no-such-profile-xyz")
	require.Error(t, err)
}

func TestLoadFromYAML_emptySteps(t *testing.T) {
	_, err := bootstrap.LoadFromYAML([]byte("name: x\nsteps: []\n"))
	require.Error(t, err)
}
