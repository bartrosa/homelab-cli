package exec_test

import (
	"context"
	"strings"
	"testing"

	"github.com/bartrosa/homelab-cli/internal/exec"
	"github.com/stretchr/testify/require"
)

type mockRunner struct {
	calls []string
}

func (m *mockRunner) Run(_ context.Context, name string, args ...string) error {
	m.calls = append(m.calls, name+" "+strings.Join(args, " "))
	return nil
}

func (m *mockRunner) RunWithOutput(_ context.Context, name string, args ...string) (string, error) {
	m.calls = append(m.calls, name+" "+strings.Join(args, " "))
	return "ok", nil
}

func TestMockRunner_recordsCalls(t *testing.T) {
	m := &mockRunner{}
	require.NoError(t, m.Run(context.Background(), "echo", "hello"))
	out, err := m.RunWithOutput(context.Background(), "true")
	require.NoError(t, err)
	require.Equal(t, "ok", out)
	require.Len(t, m.calls, 2)
}

func TestOSRunner_true(t *testing.T) {
	r := exec.NewOSRunner(nil, nil)
	require.NoError(t, r.Run(context.Background(), "true"))
	out, err := r.RunWithOutput(context.Background(), "echo", "lab")
	require.NoError(t, err)
	require.Equal(t, "lab", out)
}
