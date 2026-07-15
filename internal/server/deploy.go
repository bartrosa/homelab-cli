package server

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/bartrosa/homelab-cli/internal/config"
	"github.com/bartrosa/homelab-cli/internal/postgres"
)

// DeployMode selects post-sync action.
type DeployMode string

const (
	// DeploySync rsync only.
	DeploySync DeployMode = "sync"
	// DeployProvision rsync plus postgres apply.
	DeployProvision DeployMode = "provision"
	// DeployCompose rsync plus remote compose up.
	DeployCompose DeployMode = "compose"
	// DeployFull provision and compose.
	DeployFull DeployMode = "full"
)

// Deploy syncs homelab to server and optionally runs provision / compose.
func Deploy(ctx context.Context, cfg *config.Config, homelabRoot string, mode DeployMode, dryRun bool, stdout, stderr io.Writer) error {
	if err := Rsync(ctx, cfg, homelabRoot, dryRun, stdout, stderr); err != nil {
		return err
	}
	t, err := TargetFromConfig(cfg)
	if err != nil {
		return err
	}
	switch mode {
	case DeploySync, "":
		fmt.Fprintln(stdout, "Sync zakończony.")
		return nil
	case DeployProvision:
		fmt.Fprintln(stdout, "--- Provision PostgreSQL (lab postgres apply) ---")
		cfgPath := filepath.Join(homelabRoot, "postgres", "config", "instances.yaml")
		pgCfg, err := postgres.LoadConfig(cfgPath)
		if err != nil {
			return err
		}
		return postgres.Apply(ctx, pgCfg, dryRun)
	case DeployCompose:
		fmt.Fprintln(stdout, "--- Podman Compose (ml-stack) ---")
		return Run(ctx, t, "cd ml-stack && podman-compose up -d", nil, stdout, stderr)
	case DeployFull:
		cfgPath := filepath.Join(homelabRoot, "postgres", "config", "instances.yaml")
		pgCfg, err := postgres.LoadConfig(cfgPath)
		if err != nil {
			return err
		}
		if err := postgres.Apply(ctx, pgCfg, dryRun); err != nil {
			return err
		}
		fmt.Fprintln(stdout, "--- Podman Compose (ml-stack) ---")
		return Run(ctx, t, "cd ml-stack && podman-compose up -d", nil, stdout, stderr)
	default:
		return fmt.Errorf("unknown deploy mode %q", mode)
	}
}
