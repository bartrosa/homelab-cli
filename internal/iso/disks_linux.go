//go:build linux

package iso

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/bartrosa/homelab-cli/internal/exec"
)

// ListDisks returns block devices with USB/SYSTEM classification.
func ListDisks(ctx context.Context, runner exec.Runner) ([]Disk, error) {
	if runner == nil {
		runner = exec.NewOSRunner(os.Stdout, os.Stderr)
	}
	out, err := runner.RunWithOutput(ctx, "lsblk", "-J", "-o", "NAME,SIZE,MODEL,TRAN,RM,MOUNTPOINT,VENDOR")
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
			fmt.Fprintf(w, "  ⚠️  SYSTEM disk — do NOT write ISOs to %s\n", d.Device)
		}
	}
}

// InspectDevice reads lsblk metadata for a single device.
func InspectDevice(ctx context.Context, runner exec.Runner, device string) (DeviceInfo, error) {
	name := strings.TrimPrefix(device, "/dev/")
	out, err := runner.RunWithOutput(ctx, "lsblk", "-J", "-n", "-o", "NAME,SIZE,MODEL,TRAN,RM,MOUNTPOINT", name)
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
		Model:       dev.Model,
		Tran:        dev.Tran,
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
	name := strings.TrimPrefix(device, "/dev/")
	out, err := runner.RunWithOutput(ctx, "lsblk", "-J", "-n", "-o", "NAME,MOUNTPOINT", name)
	if err != nil {
		return err
	}
	var payload lsblkRoot
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		return err
	}
	var parts []string
	for _, d := range payload.BlockDevices {
		collectPartNames(d, &parts)
	}
	for _, p := range parts {
		if p == "" {
			continue
		}
		_ = runner.Run(ctx, "umount", "/dev/"+p)
	}
	return nil
}
