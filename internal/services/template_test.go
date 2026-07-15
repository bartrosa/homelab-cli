package services_test

import (
	"strings"
	"testing"

	"github.com/bartrosa/homelab-cli/internal/services"
	"github.com/stretchr/testify/require"
)

func TestRender_postgres_compose(t *testing.T) {
	out, err := services.Render("postgres", "compose.yml.tmpl", map[string]any{
		"BuildImage": false,
		"DataDir":    "/tmp/pg",
	})
	require.NoError(t, err)
	require.Contains(t, out, "postgres:16.4-alpine")
	require.Contains(t, out, "homelab-net")
	require.Contains(t, out, "external: true")
}

func TestRender_postgres_dockerfile_plugins(t *testing.T) {
	out, err := services.Render("postgres", "Dockerfile.tmpl", map[string]any{
		"Plugins": []string{"pgvector", "postgis"},
	})
	require.NoError(t, err)
	require.Contains(t, out, "postgresql-pgvector")
	require.Contains(t, out, "postgis")
	require.NotContains(t, out, "timescaledb-toolkit")
}

func TestTemplateFuncs_hasJoinDefault(t *testing.T) {
	out, err := services.Render("postgres", "Dockerfile.tmpl", map[string]any{
		"Plugins": []string{"pgvector"},
	})
	require.NoError(t, err)
	require.True(t, strings.Contains(out, "pgvector"))
}

func TestMaskSensitive(t *testing.T) {
	require.Equal(t, "****", services.MaskSensitive("ab"))
	require.Equal(t, "se**et", services.MaskSensitive("secret"))
}
