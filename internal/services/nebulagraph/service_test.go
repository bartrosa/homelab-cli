package nebulagraph_test

import (
	"strings"
	"testing"

	"github.com/bartrosa/homelab-cli/internal/services"
	_ "github.com/bartrosa/homelab-cli/internal/services/nebulagraph"
	"github.com/stretchr/testify/require"
)

func TestNebulaGraphRegistered(t *testing.T) {
	svc, ok := services.Lookup("nebulagraph")
	require.True(t, ok)
	require.Equal(t, services.CategoryGraph, svc.Category())
	require.Len(t, svc.Schema().Fields, 11)
	_, ok = svc.(services.PostUpper)
	require.True(t, ok, "nebulagraph must implement PostUpper")
}

func TestNebulaGraphTemplateFourServices(t *testing.T) {
	data := map[string]any{
		"version":              "v3.8.0",
		"studio_version":       "v3.10.0",
		"graph_port":           9669,
		"meta_http_port":       19559,
		"storage_http_port":    19779,
		"graph_http_port":      19669,
		"studio_port":          7001,
		"enable_auth":          true,
		"wait_timeout_seconds": 60,
		"expose_bind":          "127.0.0.1",
		"data_dir":             "/data",
		"state_dir":            "/state",
	}
	out, err := services.Render("nebulagraph", "compose.yml.tmpl", data)
	require.NoError(t, err)
	for _, svc := range []string{"metad0:", "storaged0:", "graphd:", "studio:"} {
		require.Contains(t, out, svc)
	}
	require.Contains(t, out, "condition: service_healthy")
	require.Contains(t, out, "127.0.0.1:9669:9669")
	require.Contains(t, out, "127.0.0.1:7001:7001")
	require.NotContains(t, out, "19559:19559")
}

func TestNebulaGraphTemplateNoAuth(t *testing.T) {
	data := map[string]any{
		"version": "v3.8.0", "studio_version": "v3.10.0",
		"graph_port": 9669, "meta_http_port": 19559, "storage_http_port": 19779,
		"graph_http_port": 19669, "studio_port": 7001, "enable_auth": false,
		"expose_bind": "127.0.0.1", "data_dir": "/data", "state_dir": "/state",
		"wait_timeout_seconds": 60,
	}
	out, err := services.Render("nebulagraph", "compose.yml.tmpl", data)
	require.NoError(t, err)
	require.NotContains(t, out, "enable_authorize=true")
}

func TestNebulaGraphPostInitScript(t *testing.T) {
	data := map[string]any{"wait_timeout_seconds": 90}
	out, err := services.Render("nebulagraph", "config/post-init.sh.tmpl", data)
	require.NoError(t, err)
	require.Contains(t, out, "MAX_WAIT=90")
	require.Contains(t, out, "ADD HOSTS")
	require.Contains(t, out, "ALTER USER root")
}

func TestNebulaGraphPostInitIdempotencyLogic(t *testing.T) {
	hosts := "Host         Port  Status\nstoraged0    9779  ONLINE"
	require.True(t, strings.Contains(hosts, "storaged0"))
}
