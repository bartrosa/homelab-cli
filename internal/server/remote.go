// Package server runs commands and deploy workflows on the remote homelab host.
package server

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bartrosa/homelab-cli/internal/config"
	"github.com/bartrosa/homelab-cli/internal/executil"
)

// Target holds SSH connection parameters.
type Target struct {
	Host string
	User string
	Port int
	Path string
}

// TargetFromConfig reads server.* from lab config.
func TargetFromConfig(cfg *config.Config) (Target, error) {
	s := cfg.Server
	if s.Host == "" {
		return Target{}, fmt.Errorf("server.host not set in config")
	}
	user := s.User
	if user == "" {
		user = "root"
	}
	port := s.Port
	if port == 0 {
		port = 22
	}
	if s.Path == "" {
		return Target{}, fmt.Errorf("server.path not set in config")
	}
	return Target{Host: s.Host, User: user, Port: port, Path: s.Path}, nil
}

// Run executes a shell command on the remote server in server.path.
func Run(ctx context.Context, t Target, command string, stdin io.Reader, stdout, stderr io.Writer) error {
	if strings.TrimSpace(command) == "" {
		return fmt.Errorf("empty remote command")
	}
	args := []string{
		"-p", fmt.Sprintf("%d", t.Port),
		"-o", "StrictHostKeyChecking=accept-new",
		fmt.Sprintf("%s@%s", t.User, t.Host),
		fmt.Sprintf("cd %s && %s", shellQuotePath(t.Path), command),
	}
	cmd := exec.CommandContext(ctx, "ssh", args...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ssh remote run: %w", err)
	}
	return nil
}

func shellQuotePath(p string) string {
	if strings.ContainsAny(p, " \t\n'\"$`") {
		return "'" + strings.ReplaceAll(p, "'", `'\'"''`) + "'"
	}
	return p
}

// Rsync syncs local homelabRoot to remote server path.
func Rsync(ctx context.Context, cfg *config.Config, homelabRoot string, dryRun bool, stdout, stderr io.Writer) error {
	t, err := TargetFromConfig(cfg)
	if err != nil {
		return err
	}
	local, err := absPath(homelabRoot)
	if err != nil {
		return err
	}
	dst := fmt.Sprintf("%s@%s:%s/", t.User, t.Host, strings.TrimSuffix(t.Path, "/")+"/")
	sshOpts := fmt.Sprintf("ssh -p %d -o StrictHostKeyChecking=accept-new", t.Port)
	args := []string{
		"-avz",
		"--exclude", ".git",
		"--exclude", ".env",
		"--exclude", ".venv",
		"--exclude", "yt_playlist_downloads",
		"--exclude", "*.log",
		"-e", sshOpts,
		local + "/",
		dst,
	}
	if dryRun {
		args = append([]string{"--dry-run"}, args...)
	}
	ex := executil.NewRunner(stdout, stderr)
	ex.DryRun = dryRun
	return ex.Run(ctx, "rsync", args...)
}

func absPath(p string) (string, error) {
	if p == "" {
		return "", fmt.Errorf("local path required")
	}
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		p = filepath.Join(home, strings.TrimPrefix(p, "~/"))
	}
	return filepath.Abs(p)
}
