// Package arcadedb provides the ArcadeDB graph compose service.
package arcadedb

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bartrosa/homelab-cli/internal/services"
)

const serviceID = "arcadedb"

const description = `Multi-model database with Apache 2.0 license and explicit "Apache 2.0 forever" commitment. Native support for graph (Cypher/Gremlin), document (MQL), key-value, vector search, full-text, and time-series in a single ACID engine. Built-in MCP server for LLM integration. Recommended primary graph choice for GraphRAG and knowledge graph workloads.`

func init() {
	services.Register(&Service{})
}

// Service implements the ArcadeDB homelab service.
type Service struct{}

// ID returns the service identifier.
func (s *Service) ID() string { return serviceID }

// DisplayName returns the human-readable service name.
func (s *Service) DisplayName() string { return "ArcadeDB" }

// Category returns the graph service category.
func (s *Service) Category() services.Category {
	return services.CategoryGraph
}

// Description returns the service summary shown in lab services info.
func (s *Service) Description() string { return description }

// DependsOn lists services that must start before this one.
func (s *Service) DependsOn() []string { return nil }

// Schema returns the interactive init configuration schema.
func (s *Service) Schema() services.ConfigSchema {
	return services.ConfigSchema{Fields: []services.Field{
		services.PortField("http_port", "HTTP API + Studio UI port", 2480),
		services.PortField("binary_port", "Binary protocol port", 2424),
		services.PasswordField("root_password", "Root password (min 8 chars)"),
		{
			Name: "databases", Label: "Databases to auto-create", Type: services.FieldTypeMultiSelect,
			Default: []string{}, Options: []string{},
		},
		{
			Name: "default_db_users", Label: "Default users for new databases", Type: services.FieldTypeString,
			Default: "admin:PLAIN_PASSWORD:admin", Required: true,
		},
		{
			Name: "expose", Label: "Expose mode", Type: services.FieldTypeSelect,
			Default: "local", Required: true, Options: []string{"local", "lan", "tailscale"},
		},
		{
			Name: "version", Label: "ArcadeDB version", Type: services.FieldTypeString,
			Default: "24.11.1", Required: true,
		},
		{Name: "enable_gremlin", Label: "Enable Gremlin server", Type: services.FieldTypeBool, Default: true},
		{Name: "enable_mongo_protocol", Label: "Enable MongoDB wire protocol", Type: services.FieldTypeBool, Default: false},
		{Name: "enable_redis_protocol", Label: "Enable Redis wire protocol", Type: services.FieldTypeBool, Default: false},
		{Name: "enable_mcp", Label: "Enable MCP server plugin", Type: services.FieldTypeBool, Default: true},
		{
			Name: "heap_size", Label: "JVM heap size", Type: services.FieldTypeString,
			Default: "512m", Required: true,
		},
	}}
}

// Init renders compose config and secrets for ArcadeDB.
func (s *Service) Init(ctx context.Context, opts services.InitOptions) error {
	if opts.DryRun {
		return nil
	}
	stateDir, err := services.StateDir(serviceID)
	if err != nil {
		return err
	}
	composePath := filepath.Join(stateDir, "compose.yml")
	if _, err := os.Stat(composePath); err == nil && !opts.Force {
		fmt.Fprintf(opts.Stdout, "%s: already initialized (use --force to regenerate)\n", serviceID)
		return nil
	}
	values, err := collectValues(s.Schema(), opts)
	if err != nil {
		return err
	}
	if err := services.FillSecrets(s.Schema(), values); err != nil {
		return err
	}
	if err := validateRootPassword(values); err != nil {
		return err
	}
	dataDir, err := services.DataDir(serviceID)
	if err != nil {
		return err
	}
	cfgDir, err := services.ConfigDir(serviceID)
	if err != nil {
		return err
	}
	for _, d := range []string{stateDir, dataDir, cfgDir} {
		if err := os.MkdirAll(d, 0o750); err != nil {
			return err
		}
	}
	data := templateData(values, dataDir, stateDir)
	compose, err := services.Render(serviceID, "compose.yml.tmpl", data)
	if err != nil {
		return err
	}
	if err := os.WriteFile(composePath, []byte(compose), 0o644); err != nil {
		return err
	}
	env := map[string]string{
		"ARCADEDB_ROOT_PASSWORD": fmt.Sprint(values["root_password"]),
	}
	if err := services.WriteEnvFile(filepath.Join(stateDir, ".env"), env); err != nil {
		return err
	}
	if err := saveValues(stateDir, values); err != nil {
		return err
	}
	runtime, err := services.DetectRuntime(ctx, opts.Runner, opts.Runtime)
	if err != nil {
		return err
	}
	if err := services.EnsureNetwork(ctx, opts.Runner, runtime); err != nil {
		return err
	}
	if dbs := stringSlice(values["databases"]); len(dbs) > 0 {
		fmt.Fprintf(opts.Stdout, "Databases will be auto-created on first startup: %s\n", strings.Join(dbs, ", "))
		fmt.Fprintln(opts.Stdout, "Default admin user for each database will be 'admin' with password from ARCADEDB_ROOT_PASSWORD.")
	}
	if enableMCP(values) && !mcpSupportedVersion(fmt.Sprint(values["version"])) {
		fmt.Fprintf(opts.Stderr, "warning: enable_mcp=true but MCP plugin availability in %s is unverified — see TODO in compose output\n", values["version"])
	}
	return nil
}

// Up starts the ArcadeDB compose stack.
func (s *Service) Up(ctx context.Context, opts services.InitOptions) error {
	if opts.DryRun {
		return nil
	}
	stateDir, err := services.StateDir(serviceID)
	if err != nil {
		return err
	}
	runtime, err := services.DetectRuntime(ctx, opts.Runner, opts.Runtime)
	if err != nil {
		return err
	}
	if err := services.EnsureNetwork(ctx, opts.Runner, runtime); err != nil {
		return err
	}
	return services.NewComposeRunner(opts.Runner, runtime).Up(ctx, stateDir)
}

// Down stops the ArcadeDB compose stack.
func (s *Service) Down(ctx context.Context, opts services.InitOptions) error {
	if opts.DryRun {
		return nil
	}
	stateDir, err := services.StateDir(serviceID)
	if err != nil {
		return err
	}
	runtime, err := services.DetectRuntime(ctx, opts.Runner, opts.Runtime)
	if err != nil {
		return err
	}
	return services.NewComposeRunner(opts.Runner, runtime).Down(ctx, stateDir)
}

// Status reports runtime and readiness for ArcadeDB.
func (s *Service) Status(ctx context.Context, opts services.InitOptions) (services.Status, error) {
	st := services.Status{ID: serviceID}
	stateDir, err := services.StateDir(serviceID)
	if err != nil {
		return st, err
	}
	runtime, err := services.DetectRuntime(ctx, opts.Runner, opts.Runtime)
	if err != nil {
		return st, err
	}
	cr := services.NewComposeRunner(opts.Runner, runtime)
	running, detail, err := cr.Status(ctx, stateDir)
	st.Running = running
	st.Detail = detail
	if !running {
		return st, err
	}
	values, _ := loadValues(stateDir)
	port := intValue(values["http_port"], 2480)
	if _, err := opts.Runner.RunWithOutput(ctx, "curl", "-sf", fmt.Sprintf("http://127.0.0.1:%d/api/v1/ready", port)); err == nil {
		st.Healthy = true
		st.Detail = "running, ready"
	}
	return st, nil
}

// Connect prints connection endpoints (optional interactive console).
func (s *Service) Connect(ctx context.Context, opts services.InitOptions, interactive bool) error {
	stateDir, err := services.StateDir(serviceID)
	if err != nil {
		return err
	}
	values, err := loadValues(stateDir)
	if err != nil {
		return fmt.Errorf("read saved config: %w (run lab services init %s first)", err, serviceID)
	}
	httpPort := intValue(values["http_port"], 2480)
	binaryPort := intValue(values["binary_port"], 2424)
	host := services.ExposeBind(fmt.Sprint(values["expose"]))
	if host == "0.0.0.0" {
		host = "127.0.0.1"
	}
	dbs := stringSlice(values["databases"])
	fmt.Fprintf(opts.Stdout, "Note: ArcadeDB is a multi-model database with built-in vector search. For pure vector workloads, consider Qdrant or Weaviate.\n\n")
	fmt.Fprintf(opts.Stdout, "ArcadeDB endpoints:\n\n")
	fmt.Fprintf(opts.Stdout, "  Studio UI (browser):   http://%s:%d\n", host, httpPort)
	fmt.Fprintf(opts.Stdout, "  HTTP API:              http://%s:%d/api/v1\n", host, httpPort)
	fmt.Fprintf(opts.Stdout, "  Binary protocol:       remote:%s:%d\n", host, binaryPort)
	fmt.Fprintf(opts.Stdout, "  Cypher endpoint:       http://%s:%d/api/v1/query/<database>/cypher\n", host, httpPort)
	fmt.Fprintf(opts.Stdout, "  Gremlin endpoint:      http://%s:%d/api/v1/query/<database>/gremlin\n", host, httpPort)
	fmt.Fprintf(opts.Stdout, "  SQL endpoint:          http://%s:%d/api/v1/query/<database>/sql\n", host, httpPort)
	if enableMCP(values) {
		fmt.Fprintf(opts.Stdout, "  MCP endpoint:          stdio via arcadedb-mcp-cli (see docs)\n")
	}
	fmt.Fprintf(opts.Stdout, "\nCredentials:\n  Root user:     root\n  Root password: see .env (ARCADEDB_ROOT_PASSWORD)\n")
	if len(dbs) > 0 {
		fmt.Fprintf(opts.Stdout, "\nDatabases: %s\n", strings.Join(dbs, ", "))
	}
	if interactive {
		runtime, err := services.DetectRuntime(ctx, opts.Runner, opts.Runtime)
		if err != nil {
			return err
		}
		name, args := composeExec(runtime, stateDir, "arcadedb", "bin/console.sh")
		return opts.Runner.Run(ctx, name, args...)
	}
	return nil
}

func templateData(values map[string]any, dataDir, stateDir string) map[string]any {
	dbs := stringSlice(values["databases"])
	users := fmt.Sprint(values["default_db_users"])
	var dbParts []string
	for _, db := range dbs {
		dbParts = append(dbParts, fmt.Sprintf("%s[%s]", db, users))
	}
	return map[string]any{
		"version":                  values["version"],
		"heap_size":                values["heap_size"],
		"http_port":                intValue(values["http_port"], 2480),
		"binary_port":              intValue(values["binary_port"], 2424),
		"expose_bind":              services.ExposeBind(fmt.Sprint(values["expose"])),
		"data_dir":                 dataDir,
		"state_dir":                stateDir,
		"default_databases_string": strings.Join(dbParts, ";"),
		"plugins_string":           buildPluginsString(values),
	}
}

func buildPluginsString(values map[string]any) string {
	parts := []string{"Studio:com.arcadedb.studio.Studio"}
	if boolValue(values["enable_gremlin"], true) {
		parts = append(parts, "GremlinServer:com.arcadedb.server.gremlin.GremlinServerPlugin")
	}
	if boolValue(values["enable_mongo_protocol"], false) {
		parts = append(parts, "MongoDB:com.arcadedb.mongo.MongoDBProtocolPlugin")
	}
	if boolValue(values["enable_redis_protocol"], false) {
		parts = append(parts, "Redis:com.arcadedb.redis.RedisProtocolPlugin")
	}
	if enableMCP(values) && mcpSupportedVersion(fmt.Sprint(values["version"])) {
		parts = append(parts, "MCP:com.arcadedb.mcp.MCPServerPlugin")
	}
	return strings.Join(parts, ",")
}

func enableMCP(values map[string]any) bool {
	return boolValue(values["enable_mcp"], true)
}

// mcpSupportedVersion gates MCP plugin until verified in upstream image.
// TODO: verify MCP plugin availability in arcadedb/arcadedb:24.11.1 — https://github.com/ArcadeData/arcadedb/issues
func mcpSupportedVersion(version string) bool {
	return strings.HasPrefix(version, "24.11")
}

func validateRootPassword(values map[string]any) error {
	pw := fmt.Sprint(values["root_password"])
	if len(pw) < 8 {
		return fmt.Errorf("root_password must be at least 8 characters (ArcadeDB requirement)")
	}
	return nil
}

func saveValues(stateDir string, values map[string]any) error {
	safe := map[string]any{}
	for k, v := range values {
		if k == "root_password" {
			continue
		}
		safe[k] = v
	}
	b, err := json.MarshalIndent(safe, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(stateDir, "values.json"), b, 0o600)
}

func loadValues(stateDir string) (map[string]any, error) {
	b, err := os.ReadFile(filepath.Join(stateDir, "values.json"))
	if err != nil {
		return nil, err
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func collectValues(schema services.ConfigSchema, opts services.InitOptions) (map[string]any, error) {
	out := map[string]any{}
	for k, v := range opts.Values {
		out[k] = v
	}
	if opts.NonInteractive {
		for _, f := range schema.Fields {
			if _, ok := out[f.Name]; !ok && f.Default != nil {
				out[f.Name] = f.Default
			}
		}
		return out, nil
	}
	return out, fmt.Errorf("interactive init not implemented for %s — use --yes and --set", serviceID)
}

func stringSlice(v any) []string {
	switch t := v.(type) {
	case []string:
		return t
	case []any:
		var out []string
		for _, item := range t {
			out = append(out, fmt.Sprint(item))
		}
		return out
	case string:
		if t == "" {
			return nil
		}
		return strings.Split(t, ",")
	default:
		return nil
	}
}

func intValue(v any, def int) int {
	switch t := v.(type) {
	case int:
		return t
	case float64:
		return int(t)
	case string:
		var n int
		if _, err := fmt.Sscanf(t, "%d", &n); err == nil {
			return n
		}
	}
	return def
}

func boolValue(v any, def bool) bool {
	switch t := v.(type) {
	case bool:
		return t
	case string:
		return t == "true" || t == "1"
	default:
		return def
	}
}

func composeExec(runtime, stateDir, service, cmd string) (string, []string) {
	switch runtime {
	case "docker":
		return "docker", []string{"compose", "-f", filepath.Join(stateDir, "compose.yml"), "exec", "-T", service, cmd}
	default:
		return "podman-compose", []string{"-f", filepath.Join(stateDir, "compose.yml"), "exec", "-T", service, cmd}
	}
}
