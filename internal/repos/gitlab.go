// Package repos implements repository backup and sync workflows.
package repos

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/bartrosa/homelab-cli/internal/config"
	"github.com/bartrosa/homelab-cli/internal/executil"
	"github.com/bartrosa/homelab-cli/internal/homelabroot"
)

// GitLabBackup runs the homelab GitLab account backup script.
func GitLabBackup(ctx context.Context, cfg *config.Config, homelabRoot string, dryRun bool, stdout, stderr io.Writer, jobs int) error {
	root, err := homelabroot.Resolve(firstNonEmpty(homelabRoot, cfg.Homelab.Root))
	if err != nil {
		return err
	}
	script := filepath.Join(root, "tools", "gitlab", "backup_account.py")
	if _, err := os.Stat(script); err != nil {
		return fmt.Errorf("gitlab backup script missing at %s", script)
	}

	tokenEnv := "GITLAB_TOKEN"
	for _, p := range cfg.Repos.Providers {
		if strings.EqualFold(p.Kind, "gitlab") && p.TokenEnv != "" {
			tokenEnv = p.TokenEnv
			break
		}
	}
	if os.Getenv(tokenEnv) == "" && !dryRun {
		return fmt.Errorf("set %s (Personal Access Token with read_api + read_repository)", tokenEnv)
	}

	backupDir := cfg.Repos.BackupDir
	if backupDir == "" {
		backupDir = "~/backups/repos/gitlab"
	}
	backupDir = expandHome(backupDir)

	args := []string{script, "--backup-root", backupDir}
	if jobs > 0 {
		args = append(args, "--jobs", fmt.Sprintf("%d", jobs))
	}

	ex := executil.NewRunner(stdout, stderr)
	ex.DryRun = dryRun
	ex.WorkDir = filepath.Join(root, "tools", "gitlab")
	return ex.Run(ctx, "python3", args...)
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func expandHome(p string) string {
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return p
		}
		return strings.Replace(p, "~", home, 1)
	}
	return p
}
