//go:build linux

package iso

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/bartrosa/homelab-cli/internal/exec"
	"github.com/bartrosa/homelab-cli/internal/ui"
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

// WriteISO burns an ISO to a block device with safety checks.
func WriteISO(ctx context.Context, opts WriteOptions) error {
	if opts.Runner == nil {
		opts.Runner = exec.NewOSRunner(opts.Stdout, opts.Stderr)
	}
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}
	if opts.BlockSize == "" {
		opts.BlockSize = "4M"
	}

	device := BlockDevicePath(opts.Device)
	if device == "" {
		return fmt.Errorf("device is required")
	}

	st, err := os.Stat(device)
	if err != nil {
		return fmt.Errorf("device %s: %w", device, err)
	}
	if st.Mode()&os.ModeDevice == 0 {
		return fmt.Errorf("%s is not a block device", device)
	}

	info, err := InspectDevice(ctx, opts.Runner, device)
	if err != nil {
		return err
	}

	if err := ValidateDeviceCapacity(info.Size); err != nil && !opts.Force {
		return err
	}

	diskType := ClassifyDisk(info.RM, info.Tran)
	if !opts.Force && diskType == DiskSystem {
		return fmt.Errorf("refusing to write to system disk %s (TRAN=%s RM=%v); use --force to override", device, info.Tran, info.RM)
	}
	if len(info.Mountpoints) > 0 {
		fmt.Fprintf(opts.Stdout, "Unmounting partitions on %s: %v\n", device, info.Mountpoints)
		if err := unmountDevice(ctx, opts.Runner, device); err != nil {
			return err
		}
		info, err = InspectDevice(ctx, opts.Runner, device)
		if err != nil {
			return err
		}
		if len(info.Mountpoints) > 0 && !opts.Force {
			return fmt.Errorf("device %s still has mounted partitions: %v; unmount manually or use --force", device, info.Mountpoints)
		}
	}

	isoStat, err := os.Stat(opts.ISOPath)
	if err != nil {
		return fmt.Errorf("iso file: %w", err)
	}
	isoSize := formatBytes(isoStat.Size())

	fmt.Fprintf(opts.Stdout, "About to write to:\n")
	fmt.Fprintf(opts.Stdout, "  Device:  %s\n", device)
	fmt.Fprintf(opts.Stdout, "  Model:   %s\n", info.Model)
	fmt.Fprintf(opts.Stdout, "  Size:    %s\n", info.Size)
	fmt.Fprintf(opts.Stdout, "  Source:  %s (%s)\n\n", opts.ISOPath, isoSize)
	ui.WarnLine(opts.Stdout, opts.NoColor, fmt.Sprintf("ALL DATA ON %s WILL BE DESTROYED", device))
	fmt.Fprintln(opts.Stdout)

	if !opts.Yes {
		prompt := fmt.Sprintf("Type %q to confirm: ", device)
		var answer string
		if opts.Confirm != nil {
			answer, err = opts.Confirm(prompt)
		} else {
			fmt.Fprint(opts.Stdout, prompt)
			_, err = fmt.Scanln(&answer)
		}
		if err != nil {
			return err
		}
		if answer != device {
			return fmt.Errorf("confirmation mismatch: expected %q", device)
		}
	}

	if os.Geteuid() != 0 {
		fmt.Fprintln(opts.Stdout, "Elevating with sudo for disk write (password may be required)...")
	}

	args := []string{
		"if=" + opts.ISOPath,
		"of=" + device,
		"bs=" + opts.BlockSize,
		"status=progress",
		"conv=fdatasync",
		"oflag=direct",
	}
	if err := runDDWithProgress(ctx, opts.Stdout, opts.NoColor, isoStat.Size(), device, args); err != nil {
		return err
	}
	if err := runPrivileged(ctx, opts.Runner, "sync"); err != nil {
		return err
	}

	ui.OKLine(opts.Stdout, opts.NoColor, "Done — unplug the USB drive and boot from it")
	return nil
}
