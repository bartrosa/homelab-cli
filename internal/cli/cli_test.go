package cli_test

import (
	"github.com/bartrosa/homelab-cli/internal/cli"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func isolatedHome(t *testing.T) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	require.NoError(t, os.MkdirAll(filepath.Join(home, ".config", "homelab-cli"), 0o750))
}

func TestRootHelp_showsGroups(t *testing.T) {
	isolatedHome(t)

	root := cli.NewRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"--help"})

	require.NoError(t, root.ExecuteContext(context.Background()))

	out := buf.String()
	require.Contains(t, out, "Foundation")
	require.Contains(t, out, "bootstrap")
}

func TestVersion_text(t *testing.T) {
	isolatedHome(t)

	root := cli.NewRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"version"})

	require.NoError(t, root.ExecuteContext(context.Background()))
	require.Contains(t, buf.String(), "Version:")
	require.Contains(t, buf.String(), "GoVersion:")
}

func TestVersion_jsonStdout(t *testing.T) {
	isolatedHome(t)

	root := cli.NewRootCmd()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs([]string{"version", "--output", "json"})

	require.NoError(t, root.ExecuteContext(context.Background()))

	var payload map[string]string
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &payload))
	require.Contains(t, payload, "version")
	require.Contains(t, payload, "go_version")
}

func TestVersion_debugJSONLoggingToStderr(t *testing.T) {
	isolatedHome(t)

	root := cli.NewRootCmd()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs([]string{"--log-level", "debug", "--log-format", "json", "version"})

	require.NoError(t, root.ExecuteContext(context.Background()))
	require.Contains(t, stderr.String(), "version command invoked")
	require.Contains(t, stderr.String(), "DEBUG")
}

func TestBootstrapLaptop_dryRun(t *testing.T) {
	isolatedHome(t)

	root := cli.NewRootCmd()
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"--dry-run", "--homelab-root", t.TempDir(), "bootstrap", "laptop"})

	err := root.ExecuteContext(context.Background())
	require.NoError(t, err)
}

func TestBootstrapList(t *testing.T) {
	isolatedHome(t)

	root := cli.NewRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"bootstrap", "list"})

	require.NoError(t, root.ExecuteContext(context.Background()))
	require.Contains(t, buf.String(), "laptop-macos")
}

func TestMediaHelp(t *testing.T) {
	isolatedHome(t)

	root := cli.NewRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"media", "--help"})

	require.NoError(t, root.ExecuteContext(context.Background()))
	require.Contains(t, buf.String(), "heic")
}
