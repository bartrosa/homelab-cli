package services

import (
	"context"
	"fmt"

	"github.com/bartrosa/homelab-cli/internal/exec"
)

// NetworkName is the shared Docker/Podman network for homelab services.
const NetworkName = "homelab-net"

// EnsureNetwork creates the shared homelab compose network if missing.
func EnsureNetwork(ctx context.Context, r exec.Runner, runtime string) error {
	switch runtime {
	case "docker":
		if err := r.Run(ctx, "docker", "network", "inspect", NetworkName); err != nil {
			return r.Run(ctx, "docker", "network", "create", NetworkName)
		}
		return nil
	case "podman":
		if err := r.Run(ctx, "podman", "network", "inspect", NetworkName); err != nil {
			return r.Run(ctx, "podman", "network", "create", NetworkName)
		}
		return nil
	default:
		return fmt.Errorf("unsupported runtime %q for network", runtime)
	}
}
