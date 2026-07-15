package services

import (
	"context"
	"fmt"
	"net"
	"strings"

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

// ExposeBind returns the host bind address for a service expose mode.
func ExposeBind(expose string) string {
	switch strings.ToLower(strings.TrimSpace(expose)) {
	case "lan":
		return "0.0.0.0"
	case "tailscale":
		if ip := tailscaleIPv4(); ip != "" {
			return ip
		}
		return "127.0.0.1"
	default:
		return "127.0.0.1"
	}
}

func tailscaleIPv4() string {
	iface, err := net.InterfaceByName("tailscale0")
	if err != nil {
		return ""
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return ""
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok {
			if v4 := ipnet.IP.To4(); v4 != nil {
				return v4.String()
			}
		}
	}
	return ""
}
