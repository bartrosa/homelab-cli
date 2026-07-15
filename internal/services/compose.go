package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bartrosa/homelab-cli/internal/exec"
	"github.com/bartrosa/homelab-cli/internal/prompt"
)

// ComposeRunner runs compose up/down against a project directory.
type ComposeRunner interface {
	Up(ctx context.Context, projectDir string) error
	Down(ctx context.Context, projectDir string) error
	Status(ctx context.Context, projectDir string) (running bool, detail string, err error)
}

// DetectRuntime picks docker or podman when prefer is auto/empty.
func DetectRuntime(ctx context.Context, r exec.Runner, prefer string) (string, error) {
	prefer = strings.ToLower(strings.TrimSpace(prefer))
	switch prefer {
	case "", "auto":
		if err := r.Run(ctx, "docker", "info"); err == nil {
			return "docker", nil
		}
		if err := r.Run(ctx, "podman", "info"); err == nil {
			return "podman", nil
		}
		return "", fmt.Errorf("no container runtime found (tried docker, podman)")
	case "docker", "podman":
		return prefer, nil
	default:
		if strings.HasPrefix(prefer, "podman") {
			return "podman", nil
		}
		if strings.HasPrefix(prefer, "docker") {
			return "docker", nil
		}
		return "", fmt.Errorf("unsupported runtime %q", prefer)
	}
}

// OSComposeRunner implements ComposeRunner via docker/podman compose CLIs.
type OSComposeRunner struct {
	Runner  exec.Runner
	Runtime string
}

// NewComposeRunner returns a compose runner for the given runtime.
func NewComposeRunner(r exec.Runner, runtime string) *OSComposeRunner {
	return &OSComposeRunner{Runner: r, Runtime: runtime}
}

// Up starts services in projectDir (expects compose.yml).
func (c *OSComposeRunner) Up(ctx context.Context, projectDir string) error {
	return c.run(ctx, projectDir, "up", "-d")
}

// Down stops services in projectDir.
func (c *OSComposeRunner) Down(ctx context.Context, projectDir string) error {
	return c.run(ctx, projectDir, "down")
}

// Status checks whether any service is running.
func (c *OSComposeRunner) Status(ctx context.Context, projectDir string) (bool, string, error) {
	out, err := c.runOutput(ctx, projectDir, "ps", "--status", "running", "--format", "{{.Name}}")
	if err != nil {
		return false, "", err
	}
	out = strings.TrimSpace(out)
	if out == "" {
		return false, "stopped", nil
	}
	lines := strings.Split(out, "\n")
	return len(lines) > 0, fmt.Sprintf("%d running", len(lines)), nil
}

func (c *OSComposeRunner) run(ctx context.Context, dir string, args ...string) error {
	name, cargs := composeCommand(c.Runtime, args...)
	runner := withWorkDir(c.Runner, dir)
	return runner.Run(ctx, name, cargs...)
}

func (c *OSComposeRunner) runOutput(ctx context.Context, dir string, args ...string) (string, error) {
	name, cargs := composeCommand(c.Runtime, args...)
	runner := withWorkDir(c.Runner, dir)
	return runner.RunWithOutput(ctx, name, cargs...)
}

func composeCommand(runtime string, args ...string) (string, []string) {
	switch runtime {
	case "docker":
		return "docker", append([]string{"compose", "-f", "compose.yml"}, args...)
	default:
		return "podman-compose", append([]string{"-f", "compose.yml"}, args...)
	}
}

func withWorkDir(r exec.Runner, dir string) exec.Runner {
	if osr, ok := r.(*exec.OSRunner); ok {
		cp := *osr
		cp.WorkDir = dir
		return &cp
	}
	return &workDirRunner{Runner: r, WorkDir: dir}
}

type workDirRunner struct {
	exec.Runner
	WorkDir string
}

func (w *workDirRunner) Run(ctx context.Context, name string, args ...string) error {
	if osr, ok := w.Runner.(*exec.OSRunner); ok {
		cp := *osr
		cp.WorkDir = w.WorkDir
		return cp.Run(ctx, name, args...)
	}
	return w.Runner.Run(ctx, name, args...)
}

func (w *workDirRunner) RunWithOutput(ctx context.Context, name string, args ...string) (string, error) {
	if osr, ok := w.Runner.(*exec.OSRunner); ok {
		cp := *osr
		cp.WorkDir = w.WorkDir
		return cp.RunWithOutput(ctx, name, args...)
	}
	return w.Runner.RunWithOutput(ctx, name, args...)
}

// ManagedService is a compose-backed Service with embedded templates.
type ManagedService struct {
	Meta           ServiceMeta
	SchemaDef      ConfigSchema
	Deps           []string
	ExtraTemplates map[string]string // tmpl file -> output file
	TemplateDataFn func(values map[string]any, dirs map[string]string) (any, error)
}

// NewManagedService builds a standard compose service.
func NewManagedService(meta ServiceMeta, schema ConfigSchema, deps []string, extra map[string]string) *ManagedService {
	return &ManagedService{
		Meta:           meta,
		SchemaDef:      schema,
		Deps:           deps,
		ExtraTemplates: extra,
	}
}

// ID returns the service identifier.
func (m *ManagedService) ID() string { return m.Meta.ID }

// DisplayName returns the human-readable service name.
func (m *ManagedService) DisplayName() string { return m.Meta.DisplayName }

// Category returns the service category.
func (m *ManagedService) Category() Category { return m.Meta.Category }

// Description returns a short service summary.
func (m *ManagedService) Description() string { return m.Meta.Description }

// Schema returns the init configuration schema.
func (m *ManagedService) Schema() ConfigSchema { return m.SchemaDef }

// DependsOn lists services that should start before this one.
func (m *ManagedService) DependsOn() []string { return append([]string(nil), m.Deps...) }

// Init renders templates and writes service configuration.
func (m *ManagedService) Init(ctx context.Context, opts InitOptions) error {
	return initManaged(ctx, m, opts)
}

// Status reports whether containers are running.
func (m *ManagedService) Status(ctx context.Context, opts InitOptions) (Status, error) {
	return statusManaged(ctx, m, opts)
}

// Up starts the compose stack.
func (m *ManagedService) Up(ctx context.Context, opts InitOptions) error {
	return upManaged(ctx, m, opts)
}

// Down stops the compose stack.
func (m *ManagedService) Down(ctx context.Context, opts InitOptions) error {
	return downManaged(ctx, m, opts)
}

func initManaged(ctx context.Context, m *ManagedService, opts InitOptions) error {
	if opts.DryRun {
		return nil
	}
	values, err := collectValues(m.Schema(), opts)
	if err != nil {
		return err
	}
	if err := FillSecrets(m.Schema(), values); err != nil {
		return err
	}
	stateDir, err := StateDir(m.ID())
	if err != nil {
		return err
	}
	dataDir, err := DataDir(m.ID())
	if err != nil {
		return err
	}
	cfgDir, err := ConfigDir(m.ID())
	if err != nil {
		return err
	}
	for _, d := range []string{stateDir, dataDir, cfgDir} {
		if err := os.MkdirAll(d, 0o750); err != nil {
			return err
		}
	}
	dirs := map[string]string{
		"StateDir":  stateDir,
		"DataDir":   dataDir,
		"ConfigDir": cfgDir,
	}
	var data any
	if m.TemplateDataFn != nil {
		data, err = m.TemplateDataFn(values, dirs)
	} else {
		data, err = defaultTemplateData(m.ID(), values, dirs)
	}
	if err != nil {
		return err
	}
	compose, err := Render(m.ID(), "compose.yml.tmpl", data)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(stateDir, "compose.yml"), []byte(compose), 0o644); err != nil {
		return err
	}
	for tmpl, out := range m.ExtraTemplates {
		rendered, err := Render(m.ID(), tmpl, data)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(stateDir, out), []byte(rendered), 0o644); err != nil {
			return err
		}
	}
	envVars := envFromValues(m.Schema(), values)
	envVars["DATA_DIR"] = dataDir
	envVars["CONFIG_DIR"] = cfgDir
	envVars["STATE_DIR"] = stateDir
	if err := WriteEnvFile(filepath.Join(stateDir, ".env"), envVars); err != nil {
		return err
	}
	runtime, err := DetectRuntime(ctx, opts.Runner, opts.Runtime)
	if err != nil {
		return err
	}
	return EnsureNetwork(ctx, opts.Runner, runtime)
}

func statusManaged(ctx context.Context, m *ManagedService, opts InitOptions) (Status, error) {
	st := Status{ID: m.ID()}
	stateDir, err := StateDir(m.ID())
	if err != nil {
		return st, err
	}
	runtime, err := DetectRuntime(ctx, opts.Runner, opts.Runtime)
	if err != nil {
		return st, err
	}
	cr := NewComposeRunner(opts.Runner, runtime)
	running, detail, err := cr.Status(ctx, stateDir)
	st.Running = running
	st.Detail = detail
	return st, err
}

func upManaged(ctx context.Context, m *ManagedService, opts InitOptions) error {
	if opts.DryRun {
		return nil
	}
	stateDir, err := StateDir(m.ID())
	if err != nil {
		return err
	}
	runtime, err := DetectRuntime(ctx, opts.Runner, opts.Runtime)
	if err != nil {
		return err
	}
	if err := EnsureNetwork(ctx, opts.Runner, runtime); err != nil {
		return err
	}
	cr := NewComposeRunner(opts.Runner, runtime)
	return cr.Up(ctx, stateDir)
}

func downManaged(ctx context.Context, m *ManagedService, opts InitOptions) error {
	if opts.DryRun {
		return nil
	}
	stateDir, err := StateDir(m.ID())
	if err != nil {
		return err
	}
	runtime, err := DetectRuntime(ctx, opts.Runner, opts.Runtime)
	if err != nil {
		return err
	}
	cr := NewComposeRunner(opts.Runner, runtime)
	return cr.Down(ctx, stateDir)
}

func defaultTemplateData(id string, values map[string]any, dirs map[string]string) (map[string]any, error) {
	data := map[string]any{
		"ID":        id,
		"Values":    values,
		"StateDir":  dirs["StateDir"],
		"DataDir":   dirs["DataDir"],
		"ConfigDir": dirs["ConfigDir"],
	}
	for k, v := range values {
		data[k] = v
	}
	return data, nil
}

func envFromValues(schema ConfigSchema, values map[string]any) map[string]string {
	out := map[string]string{}
	for _, f := range schema.Fields {
		if v, ok := values[f.Name]; ok {
			out[strings.ToUpper(f.Name)] = fmt.Sprint(v)
		}
	}
	return out
}

func collectValues(schema ConfigSchema, opts InitOptions) (map[string]any, error) {
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
		return out, prompt.ValidateSchema(toPromptSchema(schema), out)
	}
	if opts.Prompter == nil {
		return nil, fmt.Errorf("prompter required for interactive init")
	}
	asked, err := prompt.AskAll(opts.Prompter, toPromptSchema(schema), out)
	if err != nil {
		return nil, err
	}
	for k, v := range asked {
		out[k] = v
	}
	return out, nil
}

func toPromptSchema(s ConfigSchema) prompt.Schema {
	fields := make([]prompt.Field, len(s.Fields))
	for i, f := range s.Fields {
		fields[i] = prompt.Field{
			Name:     f.Name,
			Label:    f.Label,
			Type:     prompt.FieldType(f.Type),
			Default:  f.Default,
			Required: f.Required,
			Options:  f.Options,
		}
	}
	return prompt.Schema{Fields: fields}
}
