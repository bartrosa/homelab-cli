// Package nebulagraph provides the NebulaGraph distributed graph compose service.
package nebulagraph

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bartrosa/homelab-cli/internal/exec"
	"github.com/bartrosa/homelab-cli/internal/services"
)

const serviceID = "nebulagraph"

const description = `Distributed graph database designed for massive, highly-connected datasets (billions of vertices, trillions of edges). Apache 2.0 license, CNCF Database Landscape. openCypher-compatible (nGQL). Used at Tencent, Vivo, Meituan, JD Digits. Recommended for scale-out knowledge graphs and production workloads requiring horizontal scalability.`

func init() {
	services.Register(&Service{})
}

// Service implements NebulaGraph with post-up host registration.
type Service struct{}

// ID returns the service identifier.
func (s *Service) ID() string { return serviceID }

// DisplayName returns the human-readable service name.
func (s *Service) DisplayName() string { return "NebulaGraph" }

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
		services.PortField("graph_port", "nGQL client port", 9669),
		{Name: "meta_http_port", Label: "Meta HTTP port (internal)", Type: services.FieldTypeInt, Default: 19559, Required: true},
		{Name: "storage_http_port", Label: "Storage HTTP port (internal)", Type: services.FieldTypeInt, Default: 19779, Required: true},
		{Name: "graph_http_port", Label: "Graph HTTP port (internal)", Type: services.FieldTypeInt, Default: 19669, Required: true},
		services.PortField("studio_port", "Studio UI port", 7001),
		services.PasswordField("root_password", "Root password"),
		{
			Name: "expose", Label: "Expose mode", Type: services.FieldTypeSelect,
			Default: "local", Required: true, Options: []string{"local", "lan", "tailscale"},
		},
		{Name: "version", Label: "NebulaGraph version", Type: services.FieldTypeString, Default: "v3.8.0", Required: true},
		{Name: "studio_version", Label: "Studio version", Type: services.FieldTypeString, Default: "v3.10.0", Required: true},
		{Name: "enable_auth", Label: "Enable authentication", Type: services.FieldTypeBool, Default: true},
		{Name: "wait_timeout_seconds", Label: "Cluster ready timeout (seconds)", Type: services.FieldTypeInt, Default: 60, Required: true},
	}}
}

// Init renders compose config, post-init script, and secrets.
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
	dataDir, err := services.DataDir(serviceID)
	if err != nil {
		return err
	}
	for _, d := range []string{stateDir, dataDir, filepath.Join(stateDir, "config"), filepath.Join(stateDir, "logs/meta0"), filepath.Join(stateDir, "logs/storage0"), filepath.Join(stateDir, "logs/graphd")} {
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
	postInit, err := services.Render(serviceID, "config/post-init.sh.tmpl", data)
	if err != nil {
		return err
	}
	postInitPath := filepath.Join(stateDir, "config", "post-init.sh")
	if err := os.WriteFile(postInitPath, []byte(postInit), 0o755); err != nil {
		return err
	}
	env := map[string]string{
		"NEBULA_ROOT_PASSWORD": fmt.Sprint(values["root_password"]),
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
	return services.EnsureNetwork(ctx, opts.Runner, runtime)
}

// Up starts the NebulaGraph compose stack.
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

// PostUp registers storage hosts and sets the root password after compose up.
func (s *Service) PostUp(ctx context.Context, opts services.InitOptions) error {
	if opts.DryRun {
		return nil
	}
	stateDir, err := services.StateDir(serviceID)
	if err != nil {
		return err
	}
	values, err := loadValues(stateDir)
	if err != nil {
		return err
	}
	timeout := intValue(values["wait_timeout_seconds"], 60)
	runtime, err := services.DetectRuntime(ctx, opts.Runner, opts.Runtime)
	if err != nil {
		return err
	}
	if err := waitHealthy(ctx, opts, runtime, stateDir, timeout); err != nil {
		fmt.Fprintf(opts.Stderr, "warning: nebulagraph cluster not healthy within %ds: %v\n", timeout, err)
	}
	script := filepath.Join(stateDir, "config", "post-init.sh")
	if err := runPostInit(ctx, opts, runtime, stateDir, script); err != nil {
		fmt.Fprintf(opts.Stderr, "warning: Post-init failed — you may need to run it manually: bash %s\n", script)
		return err
	}
	return nil
}

// Down stops the NebulaGraph compose stack.
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

// Status reports cluster runtime and storage health.
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
	if storageOnline(ctx, opts, runtime, stateDir) {
		st.Healthy = true
		st.Detail = "running, storage ONLINE"
	} else {
		st.Detail = "running, storage not ONLINE (run post-init?)"
	}
	return st, nil
}

// Connect prints NebulaGraph endpoints (optional nebula-console session).
func (s *Service) Connect(ctx context.Context, opts services.InitOptions, interactive bool) error {
	stateDir, err := services.StateDir(serviceID)
	if err != nil {
		return err
	}
	values, err := loadValues(stateDir)
	if err != nil {
		return fmt.Errorf("read saved config: %w (run lab services init %s first)", err, serviceID)
	}
	host := services.ExposeBind(fmt.Sprint(values["expose"]))
	if host == "0.0.0.0" {
		host = "127.0.0.1"
	}
	graphPort := intValue(values["graph_port"], 9669)
	studioPort := intValue(values["studio_port"], 7001)
	fmt.Fprintf(opts.Stdout, "Note: NebulaGraph is a distributed graph database. For built-in vector search in one process, consider ArcadeDB.\n\n")
	fmt.Fprintf(opts.Stdout, "NebulaGraph endpoints:\n\n")
	fmt.Fprintf(opts.Stdout, "  nGQL (client protocol):   %s:%d\n", host, graphPort)
	fmt.Fprintf(opts.Stdout, "  Studio UI (browser):      http://%s:%d\n", host, studioPort)
	fmt.Fprintf(opts.Stdout, "                            Connect: graphd, port 9669, user root, password from .env\n\n")
	fmt.Fprintf(opts.Stdout, "Credentials:\n  User:     root\n  Password: see .env (NEBULA_ROOT_PASSWORD)\n\n")
	fmt.Fprintln(opts.Stdout, "Language: nGQL (openCypher-compatible)")
	fmt.Fprintln(opts.Stdout, "Clients:  nebula-console, nebula-python, nebula-java, nebula-go")
	fmt.Fprintln(opts.Stdout, "Docs:     https://docs.nebula-graph.io/")
	if interactive {
		runtime, err := services.DetectRuntime(ctx, opts.Runner, opts.Runtime)
		if err != nil {
			return err
		}
		pass := envPassword(stateDir)
		name, args := composeExec(runtime, stateDir, "graphd", "nebula-console", "-addr", "graphd", "-port", "9669", "-u", "root", "-p", pass)
		return workDirRunner(opts.Runner, stateDir, nil).Run(ctx, name, args...)
	}
	return nil
}

func templateData(values map[string]any, dataDir, stateDir string) map[string]any {
	return map[string]any{
		"version":              values["version"],
		"studio_version":       values["studio_version"],
		"graph_port":           intValue(values["graph_port"], 9669),
		"meta_http_port":       intValue(values["meta_http_port"], 19559),
		"storage_http_port":    intValue(values["storage_http_port"], 19779),
		"graph_http_port":      intValue(values["graph_http_port"], 19669),
		"studio_port":          intValue(values["studio_port"], 7001),
		"enable_auth":          boolValue(values["enable_auth"], true),
		"wait_timeout_seconds": intValue(values["wait_timeout_seconds"], 60),
		"expose_bind":          services.ExposeBind(fmt.Sprint(values["expose"])),
		"data_dir":             dataDir,
		"state_dir":            stateDir,
	}
}

func waitHealthy(ctx context.Context, opts services.InitOptions, runtime, stateDir string, timeoutSec int) error {
	deadline := time.Now().Add(time.Duration(timeoutSec) * time.Second)
	for time.Now().Before(deadline) {
		out, err := composeOutput(ctx, opts, runtime, stateDir, "ps", "--format", "{{.Service}} {{.Health}}")
		if err == nil && strings.Contains(out, "graphd") && strings.Contains(out, "healthy") {
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("timeout waiting for healthy containers")
}

func runPostInit(ctx context.Context, opts services.InitOptions, runtime, stateDir, script string) error {
	composeCmd := "docker compose"
	if runtime == "podman" {
		composeCmd = "podman-compose"
	}
	scriptBody, err := os.ReadFile(script)
	if err != nil {
		return err
	}
	patched := strings.ReplaceAll(string(scriptBody), "docker compose", composeCmd)
	tmp := filepath.Join(stateDir, "config", ".post-init-run.sh")
	if err := os.WriteFile(tmp, []byte(patched), 0o755); err != nil {
		return err
	}
	runner := workDirRunner(opts.Runner, stateDir, []string{
		"NEBULA_ROOT_PASSWORD=" + envPassword(stateDir),
	})
	return runner.Run(ctx, "bash", tmp)
}

func storageOnline(ctx context.Context, opts services.InitOptions, runtime, stateDir string) bool {
	pass := envPassword(stateDir)
	if pass == "" {
		pass = "nebula"
	}
	out, err := composeOutput(ctx, opts, runtime, stateDir, "exec", "-T", "graphd", "nebula-console", "-addr", "graphd", "-port", "9669", "-u", "root", "-p", pass, "-e", "SHOW HOSTS;")
	if err != nil {
		return false
	}
	return strings.Contains(out, "storaged0") && strings.Contains(strings.ToUpper(out), "ONLINE")
}

func envPassword(stateDir string) string {
	f, err := os.Open(filepath.Join(stateDir, ".env"))
	if err != nil {
		return ""
	}
	defer func() { _ = f.Close() }()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "NEBULA_ROOT_PASSWORD=") {
			return strings.TrimPrefix(line, "NEBULA_ROOT_PASSWORD=")
		}
	}
	return ""
}

func composeOutput(ctx context.Context, opts services.InitOptions, runtime, stateDir string, args ...string) (string, error) {
	name, cargs := composeCommand(runtime, stateDir, args...)
	runner := workDirRunner(opts.Runner, stateDir, nil)
	return runner.RunWithOutput(ctx, name, cargs...)
}

func composeCommand(runtime, stateDir string, args ...string) (string, []string) {
	switch runtime {
	case "docker":
		base := []string{"compose", "-f", filepath.Join(stateDir, "compose.yml")}
		return "docker", append(base, args...)
	default:
		base := []string{"-f", filepath.Join(stateDir, "compose.yml")}
		return "podman-compose", append(base, args...)
	}
}

func composeExec(runtime, stateDir, service string, cmdArgs ...string) (string, []string) {
	execArgs := append([]string{"exec", "-T", service}, cmdArgs...)
	return composeCommand(runtime, stateDir, execArgs...)
}

func workDirRunner(r exec.Runner, dir string, extraEnv []string) exec.Runner {
	if osr, ok := r.(*exec.OSRunner); ok {
		cp := *osr
		cp.WorkDir = dir
		cp.Env = append(cp.Env, extraEnv...)
		return &cp
	}
	return r
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
