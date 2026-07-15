package services_test

import (
	"testing"

	"github.com/bartrosa/homelab-cli/internal/services"
	_ "github.com/bartrosa/homelab-cli/internal/services/register"
	"github.com/stretchr/testify/require"
)

func TestExpandNames_preset(t *testing.T) {
	ids, err := services.ExpandNames([]string{"observability"}, nil)
	require.NoError(t, err)
	require.Equal(t, []string{"prometheus", "loki", "tempo", "grafana"}, ids)
}

func TestResolveServiceOrder_grafanaLast(t *testing.T) {
	order, err := services.ResolveServiceOrderForTest([]string{"grafana", "prometheus", "loki", "tempo"})
	require.NoError(t, err)
	require.Equal(t, []string{"prometheus", "loki", "tempo", "grafana"}, order)
}

func TestResolvePreset_mlStack(t *testing.T) {
	ids, err := services.ResolvePreset("ml-stack", nil)
	require.NoError(t, err)
	require.Contains(t, ids, "postgres")
	require.Contains(t, ids, "clickhouse")
	require.Contains(t, ids, "qdrant")
}
