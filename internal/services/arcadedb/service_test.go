package arcadedb_test

import (
	"strings"
	"testing"

	"github.com/bartrosa/homelab-cli/internal/services"
	_ "github.com/bartrosa/homelab-cli/internal/services/arcadedb"
	"github.com/stretchr/testify/require"
)

func TestArcadeDBRegistered(t *testing.T) {
	svc, ok := services.Lookup("arcadedb")
	require.True(t, ok)
	require.Equal(t, services.CategoryGraph, svc.Category())
	require.Len(t, svc.Schema().Fields, 12)
}

func TestArcadeDBTemplatePlugins(t *testing.T) {
	data := map[string]any{
		"version":                  "24.11.1",
		"heap_size":                "512m",
		"http_port":                2480,
		"binary_port":              2424,
		"expose_bind":              "127.0.0.1",
		"data_dir":                 "/data",
		"state_dir":                "/state",
		"default_databases_string": "knowledge_graph[admin:PLAIN_PASSWORD:admin]",
		"plugins_string":           "Studio:com.arcadedb.studio.Studio,GremlinServer:com.arcadedb.server.gremlin.GremlinServerPlugin",
	}
	out, err := services.Render("arcadedb", "compose.yml.tmpl", data)
	require.NoError(t, err)
	require.Contains(t, out, "arcadedata/arcadedb:24.11.1")
	require.Contains(t, out, "GremlinServerPlugin")
	require.Contains(t, out, "defaultDatabases=knowledge_graph")
	require.Contains(t, out, "127.0.0.1:2480:2480")
}

func TestArcadeDBTemplateExposeLan(t *testing.T) {
	data := map[string]any{
		"version":        "24.11.1",
		"heap_size":      "512m",
		"http_port":      2480,
		"binary_port":    2424,
		"expose_bind":    "0.0.0.0",
		"data_dir":       "/data",
		"state_dir":      "/state",
		"plugins_string": "Studio:com.arcadedb.studio.Studio",
	}
	out, err := services.Render("arcadedb", "compose.yml.tmpl", data)
	require.NoError(t, err)
	require.Contains(t, out, "0.0.0.0:2480:2480")
}

func TestArcadeDBPasswordValidation(t *testing.T) {
	svc, ok := services.Lookup("arcadedb")
	require.True(t, ok)
	var pw *services.Field
	for _, f := range svc.Schema().Fields {
		if f.Name == "root_password" {
			pw = &f
			break
		}
	}
	require.NotNil(t, pw)
	require.True(t, pw.Required)
	require.True(t, pw.Sensitive)
}

func TestArcadeDBPluginsCombinations(t *testing.T) {
	cases := []struct {
		name     string
		values   map[string]any
		contains []string
		omit     []string
	}{
		{
			name: "gremlin only",
			values: map[string]any{
				"enable_gremlin": true, "enable_mongo_protocol": false,
				"enable_redis_protocol": false, "enable_mcp": false, "version": "24.11.1",
			},
			contains: []string{"GremlinServerPlugin"},
			omit:     []string{"MongoDBProtocolPlugin", "MCPServerPlugin"},
		},
		{
			name: "mongo redis",
			values: map[string]any{
				"enable_gremlin": false, "enable_mongo_protocol": true,
				"enable_redis_protocol": true, "enable_mcp": false, "version": "24.11.1",
			},
			contains: []string{"MongoDBProtocolPlugin", "RedisProtocolPlugin"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			plugins := buildPluginsForTest(tc.values)
			for _, s := range tc.contains {
				require.Contains(t, plugins, s)
			}
			for _, s := range tc.omit {
				require.NotContains(t, plugins, s)
			}
		})
	}
}

func buildPluginsForTest(values map[string]any) string {
	parts := []string{"Studio:com.arcadedb.studio.Studio"}
	if boolVal(values["enable_gremlin"], true) {
		parts = append(parts, "GremlinServer:com.arcadedb.server.gremlin.GremlinServerPlugin")
	}
	if boolVal(values["enable_mongo_protocol"], false) {
		parts = append(parts, "MongoDB:com.arcadedb.mongo.MongoDBProtocolPlugin")
	}
	if boolVal(values["enable_redis_protocol"], false) {
		parts = append(parts, "Redis:com.arcadedb.redis.RedisProtocolPlugin")
	}
	if boolVal(values["enable_mcp"], true) && strings.HasPrefix(values["version"].(string), "24.11") {
		parts = append(parts, "MCP:com.arcadedb.mcp.MCPServerPlugin")
	}
	return strings.Join(parts, ",")
}

func boolVal(v any, def bool) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return def
}
