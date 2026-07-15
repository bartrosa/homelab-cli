package stack_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/bartrosa/homelab-cli/internal/stack"
	"github.com/stretchr/testify/require"
)

type mockComponent struct {
	id        string
	ver       string
	requires  []string
	installed bool
}

func (m *mockComponent) ID() string                     { return m.id }
func (m *mockComponent) DisplayName() string            { return m.id }
func (m *mockComponent) Category() stack.Category       { return stack.CategoryLanguage }
func (m *mockComponent) Description() string            { return "mock" }
func (m *mockComponent) DefaultVersion() string         { return "1" }
func (m *mockComponent) Requires() []string             { return m.requires }
func (m *mockComponent) PathEntries() []stack.PathEntry { return nil }
func (m *mockComponent) IsInstalled(context.Context, *stack.Env) (bool, string, error) {
	return m.installed, m.ver, nil
}

func (m *mockComponent) Install(context.Context, *stack.Env, stack.InstallOptions) error {
	m.installed = true
	return nil
}

func TestResolveOrder_dependencies(t *testing.T) {
	stack.Register(&mockComponent{id: "java", ver: "21"})
	stack.Register(&mockComponent{id: "kotlin", ver: "latest", requires: []string{"java"}})

	env := &stack.Env{Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}}
	steps, err := stack.Plan(context.Background(), env, []string{"kotlin"}, stack.InstallOptions{})
	require.NoError(t, err)
	require.Len(t, steps, 2)
	require.Equal(t, "java", steps[0].ID)
	require.Equal(t, "kotlin", steps[1].ID)
}

func TestPlan_skipInstalled(t *testing.T) {
	stack.Register(&mockComponent{id: "git", ver: "2.43", installed: true})

	env := &stack.Env{Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}}
	steps, err := stack.Plan(context.Background(), env, []string{"git"}, stack.InstallOptions{})
	require.NoError(t, err)
	require.Len(t, steps, 1)
	require.Equal(t, "skip", steps[0].Action)
}
