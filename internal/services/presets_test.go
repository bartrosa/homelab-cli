package services_test

import (
	"testing"

	"github.com/bartrosa/homelab-cli/internal/services"
	_ "github.com/bartrosa/homelab-cli/internal/services/register"
	"github.com/stretchr/testify/require"
)

func TestResolvePreset_graphrag(t *testing.T) {
	ids, err := services.ResolvePreset("graphrag", nil)
	require.NoError(t, err)
	require.Equal(t, []string{"arcadedb", "qdrant", "minio", "postgres"}, ids)
}

func TestResolvePreset_graphLab(t *testing.T) {
	ids, err := services.ResolvePreset("graph-lab", nil)
	require.NoError(t, err)
	require.Equal(t, []string{"arcadedb", "nebulagraph"}, ids)
}

func TestGraphServicesRegistered(t *testing.T) {
	for _, id := range []string{"arcadedb", "nebulagraph"} {
		svc, ok := services.Lookup(id)
		require.True(t, ok, id)
		require.Equal(t, services.CategoryGraph, svc.Category())
	}
}
