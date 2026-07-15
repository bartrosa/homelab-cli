//go:build linux

package iso

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/bartrosa/homelab-cli/internal/exec"
)

// ListDisks returns block devices with USB/SYSTEM classification.
func ListDisks(ctx context.Context, runner exec.Runner) ([]Disk, error) {
	if runner == nil {
		runner = exec.NewOSRunner(os.Stdout, os.Stderr)
	}
	out, err := runner.RunWithOutput(ctx, "lsblk", "-d", "-J", "-o", "NAME,SIZE,MODEL,TRAN,RM,TYPE,VENDOR")
	if err != nil {
		return nil, err
	}
	return ParseLSBLKJSON(out)
}

// WriteDisksTable prints disk listing.
func WriteDisksTable(w io.Writer, disks []Disk) {
	fmt.Fprintf(w, "%-12s %-8s %-24s %-6s %s\n", "DEVICE", "SIZE", "MODEL", "TRAN", "TYPE")
	for _, d := range disks {
		fmt.Fprintln(w, FormatDiskLine(d))
		if d.Type == DiskSystem {
			fmt.Fprintf(w, "  ! SYSTEM disk — do NOT write ISOs to %s\n", d.Device)
		}
	}
}

// InspectDevice reads lsblk metadata for a single device.
func InspectDevice(ctx context.Context, runner exec.Runner, device string) (DeviceInfo, error) {
	device = BlockDevicePath(device)
	out, err := runner.RunWithOutput(ctx, "lsblk", "-J", "-o", "NAME,SIZE,MODEL,TRAN,RM,TYPE,MOUNTPOINT", device)
	if err != nil {
		return DeviceInfo{}, err
	}
	var payload lsblkRoot
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		return DeviceInfo{}, err
	}
	if len(payload.BlockDevices) == 0 {
		return DeviceInfo{}, fmt.Errorf("device %s not found", device)
	}
	dev := payload.BlockDevices[0]
	var mps []string
	collectMountpoints(dev, &mps)
	return DeviceInfo{
		Device:      device,
		Size:        dev.Size,
		Model:       strVal(dev.Model),
		Tran:        strVal(dev.Tran),
		RM:          dev.RM,
		Mountpoints: mps,
	}, nil
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
	device = BlockDevicePath(device)
	out, err := runner.RunWithOutput(ctx, "lsblk", "-J", "-o", "NAME,MOUNTPOINT", device)
	if err != nil {
		return err
	}
	var payload lsblkRoot
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		return err
	}
	var targets []string
	for _, d := range payload.BlockDevices {
		collectUmountTargets(d, &targets)
	}
	seen := make(map[string]struct{})
	for _, t := range targets {
		if t == "" {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		tryUmount(ctx, runner, t)
	}
	return nil
}
