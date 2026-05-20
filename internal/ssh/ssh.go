// Package ssh provides host inventory and connection helpers.
package ssh

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/bartrosa/homelab-cli/internal/config"
	"github.com/bartrosa/homelab-cli/internal/homelabroot"
	"github.com/bartrosa/homelab-cli/internal/server"
)

// Connect opens an interactive SSH session to a configured host alias.
func Connect(ctx context.Context, cfg *config.Config, alias string) error {
	host, ok := cfg.SSH.Hosts[alias]
	if !ok {
		return fmt.Errorf("unknown host %q (configured: %s)", alias, strings.Join(cfg.SSHHostNames(), ", "))
	}
	target := host.Target()
	args := []string{}
	if host.Port > 0 && host.Port != 22 {
		args = append(args, "-p", fmt.Sprintf("%d", host.Port))
	}
	if host.IdentityFile != "" {
		args = append(args, "-i", expandHome(host.IdentityFile))
	}
	args = append(args, target)
	cmd := exec.CommandContext(ctx, "ssh", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ssh %s: %w", alias, err)
	}
	return nil
}

// SyncRepo rsyncs homelab to the configured remote server (ssh + rsync).
func SyncRepo(ctx context.Context, cfg *config.Config, homelabRoot string, dryRun bool, stdout, stderr io.Writer) error {
	root, err := homelabroot.Resolve(firstNonEmpty(homelabRoot, cfg.Homelab.Root))
	if err != nil {
		return err
	}
	return server.Rsync(ctx, cfg, root, dryRun, stdout, stderr)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func expandHome(p string) string {
	if p == "" {
		return p
	}
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return p
		}
		return strings.Replace(p, "~", home, 1)
	}
	return p
}
