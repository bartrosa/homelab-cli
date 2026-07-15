//go:build !linux

package iso

import (
	"context"
	"fmt"
	"io"

	"github.com/bartrosa/homelab-cli/internal/clierrors"
	"github.com/bartrosa/homelab-cli/internal/exec"
)

// ListDisks is only supported on Linux.
func ListDisks(ctx context.Context, runner exec.Runner) ([]Disk, error) {
	return nil, fmt.Errorf("lab iso disks: %w", clierrors.ErrNotImplemented)
}

// WriteDisksTable prints disk listing.
func WriteDisksTable(w io.Writer, disks []Disk) {
	for _, d := range disks {
		fmt.Fprintln(w, FormatDiskLine(d))
	}
}

// InspectDevice is only supported on Linux.
func InspectDevice(ctx context.Context, runner exec.Runner, device string) (DeviceInfo, error) {
	return DeviceInfo{}, fmt.Errorf("lab iso write: %w", clierrors.ErrNotImplemented)
}

// DeviceInfo holds metadata for write safety checks.
type DeviceInfo struct {
	Device      string
	Size        string
	Model       string
	Tran        string
	RM          bool
	Mountpoints []string
}

func unmountDevice(ctx context.Context, runner exec.Runner, device string) error {
	return nil
}
