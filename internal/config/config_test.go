package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoad_defaultsWhenFileMissing(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	path := filepath.Join(home, ".config", "homelab-cli", "config.yaml")
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o750))

	cfg, err := Load(path)
	require.NoError(t, err)
	require.Equal(t, "info", cfg.LogLevel)
	require.Equal(t, "text", cfg.LogFormat)
	require.Equal(t, "auto", cfg.Services.Runtime)
}

func TestLoad_envOverridesDefaults(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("LAB_LOG_LEVEL", "debug")
	t.Setenv("LAB_SERVICES_RUNTIME", "docker")

	path := filepath.Join(home, ".config", "homelab-cli", "config.yaml")
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o750))

	cfg, err := Load(path)
	require.NoError(t, err)
	require.Equal(t, "debug", cfg.LogLevel)
	require.Equal(t, "docker", cfg.Services.Runtime)
}

func TestLoad_fileOverridesDefaults(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	path := filepath.Join(home, ".config", "homelab-cli", "config.yaml")
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o750))

	require.NoError(t, os.WriteFile(path, []byte(`log_level: warn
services:
  runtime: docker
`), 0o600))

	cfg, err := Load(path)
	require.NoError(t, err)
	require.Equal(t, "warn", cfg.LogLevel)
	require.Equal(t, "docker", cfg.Services.Runtime)
}
