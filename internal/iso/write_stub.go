//go:build !linux

package iso

import (
	"context"
	"fmt"
	"io"

	"github.com/bartrosa/homelab-cli/internal/clierrors"
	"github.com/bartrosa/homelab-cli/internal/exec"
)

// WriteOptions configures ISO write to block device.
type WriteOptions struct {
	ISOPath   string
	Device    string
	Yes       bool
	Force     bool
	NoColor   bool
	BlockSize string
	Stdout    io.Writer
	Stderr    io.Writer
	Runner    exec.Runner
	Confirm   func(prompt string) (string, error)
}

// WriteISO burns an ISO to a block device (Linux only).
func WriteISO(ctx context.Context, opts WriteOptions) error {
	return fmt.Errorf("lab iso write: %w", clierrors.ErrNotImplemented)
}
