package postgres_test

import (
	"testing"

	"github.com/bartrosa/homelab-cli/internal/services"
	_ "github.com/bartrosa/homelab-cli/internal/services/postgres"
	"github.com/stretchr/testify/require"
)

func TestPostgresRegistered(t *testing.T) {
	svc, ok := services.Lookup("postgres")
	require.True(t, ok)
	require.Equal(t, "PostgreSQL", svc.DisplayName())
	require.Equal(t, services.CategoryDatabase, svc.Category())
}

func TestPostgresSchemaPlugins(t *testing.T) {
	svc, ok := services.Lookup("postgres")
	require.True(t, ok)
	schema := svc.Schema()
	var plugins *services.Field
	for i := range schema.Fields {
		if schema.Fields[i].Name == "plugins" {
			plugins = &schema.Fields[i]
			break
		}
	}
	require.NotNil(t, plugins)
	require.Equal(t, services.FieldTypeMultiSelect, plugins.Type)
	require.Equal(t, []string{"pgvector", "postgis", "timescaledb"}, plugins.Options)
}

func TestPostgresComposeBuildWhenPlugins(t *testing.T) {
	out, err := services.Render("postgres", "compose.yml.tmpl", map[string]any{
		"BuildImage": true,
		"DataDir":    "/tmp/data",
	})
	require.NoError(t, err)
	require.Contains(t, out, "build: .")
}

func TestPostgresComposeImageWhenNoPlugins(t *testing.T) {
	out, err := services.Render("postgres", "compose.yml.tmpl", map[string]any{
		"BuildImage": false,
		"DataDir":    "/tmp/data",
	})
	require.NoError(t, err)
	require.Contains(t, out, "postgres:16.4-alpine")
}
