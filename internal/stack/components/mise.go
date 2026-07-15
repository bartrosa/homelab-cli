package components

import (
	"context"

	"github.com/bartrosa/homelab-cli/internal/stack"
)

type miseComponent struct {
	id, displayName, description, defaultVer string
	category                                 stack.Category
	requires                                 []string
	extraPaths                               []stack.PathEntry
	postInstall                              func(context.Context, *stack.Env, stack.InstallOptions) error
}

func (m *miseComponent) ID() string               { return m.id }
func (m *miseComponent) DisplayName() string      { return m.displayName }
func (m *miseComponent) Category() stack.Category { return m.category }
func (m *miseComponent) Description() string      { return m.description }
func (m *miseComponent) DefaultVersion() string   { return m.defaultVer }
func (m *miseComponent) Requires() []string       { return m.requires }
func (m *miseComponent) PathEntries() []stack.PathEntry {
	base := []stack.PathEntry{{
		Shell: "all", Marker: "mise",
		Content: `if [ -x "$HOME/.local/bin/mise" ]; then
    eval "$($HOME/.local/bin/mise activate bash)"
fi`,
	}}
	return append(base, m.extraPaths...)
}

func (m *miseComponent) IsInstalled(ctx context.Context, env *stack.Env) (bool, string, error) {
	ok, ver := miseInstalled(ctx, env, m.id)
	return ok, ver, nil
}

func (m *miseComponent) Install(ctx context.Context, env *stack.Env, opts stack.InstallOptions) error {
	if err := ensureMise(ctx, env, opts.DryRun); err != nil {
		return err
	}
	ver := opts.Version
	if ver == "" {
		ver = m.defaultVer
	}
	spec := miseSpec(m.id, ver)
	if opts.DryRun {
		return env.Runner.Run(ctx, "mise", "use", "-g", spec)
	}
	if err := env.Runner.Run(ctx, "mise", "use", "-g", spec); err != nil {
		return err
	}
	if m.postInstall != nil {
		return m.postInstall(ctx, env, opts)
	}
	return nil
}

func registerMiseComponents() {
	langs := []miseComponent{
		{id: "python", displayName: "Python", category: stack.CategoryLanguage, description: "Python via mise", defaultVer: "3.13"},
		{id: "node", displayName: "Node.js", category: stack.CategoryLanguage, description: "Node.js LTS via mise", defaultVer: "lts"},
		{id: "bun", displayName: "Bun", category: stack.CategoryLanguage, description: "Bun via mise", defaultVer: "latest"},
		{id: "go", displayName: "Go", category: stack.CategoryLanguage, description: "Go via mise", defaultVer: "1.25", extraPaths: []stack.PathEntry{
			{Shell: "all", Marker: "go-path", Content: `export GOPATH="$HOME/go"\nexport PATH="$GOPATH/bin:$PATH"`},
		}},
		{id: "zig", displayName: "Zig", category: stack.CategoryLanguage, description: "Zig via mise (+ zls)", defaultVer: "latest", postInstall: func(ctx context.Context, env *stack.Env, _ stack.InstallOptions) error {
			return env.Runner.Run(ctx, "mise", "use", "-g", "zls@latest")
		}},
		{id: "lua", displayName: "Lua", category: stack.CategoryLanguage, description: "Lua via mise", defaultVer: "5.4"},
		{id: "java", displayName: "Java", category: stack.CategoryLanguage, description: "Java LTS via mise", defaultVer: "21", extraPaths: []stack.PathEntry{
			{Shell: "all", Marker: "java-home", Content: `if command -v mise > /dev/null 2>&1; then export JAVA_HOME="$(mise where java 2>/dev/null || true)"; fi`},
		}},
		{id: "kotlin", displayName: "Kotlin", category: stack.CategoryLanguage, description: "Kotlin via mise", defaultVer: "latest", requires: []string{"java"}},
		{id: "erlang", displayName: "Erlang", category: stack.CategoryLanguage, description: "Erlang via mise", defaultVer: "latest"},
		{id: "elixir", displayName: "Elixir", category: stack.CategoryLanguage, description: "Elixir via mise", defaultVer: "latest", requires: []string{"erlang"}},
		{id: "deno", displayName: "Deno", category: stack.CategoryLanguage, description: "Deno via mise", defaultVer: "latest"},
	}
	for i := range langs {
		c := langs[i]
		stack.Register(&c)
	}
}
