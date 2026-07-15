package bootstrap

import (
	"context"

	"github.com/bartrosa/homelab-cli/internal/exec"
)

// EnsureFlathubRemote adds Flathub remote on Silverblue if missing.
func EnsureFlathubRemote(ctx context.Context, runner exec.Runner) error {
	return EnsureFlathub(ctx, runner)
}
