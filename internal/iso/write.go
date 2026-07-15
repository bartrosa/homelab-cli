//go:build linux

package iso

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/bartrosa/homelab-cli/internal/exec"
)

// WriteOptions configures ISO write to block device.
type WriteOptions struct {
	ISOPath   string
	Device    string
	Yes       bool
	Force     bool
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

	device := opts.Device
	if !strings.HasPrefix(device, "/dev/") {
		device = "/dev/" + strings.TrimPrefix(device, "/dev/")
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
	if len(info.Mountpoints) > 0 && !opts.Force {
		return fmt.Errorf("device %s has mounted partitions: %v; unmount first or use --force", device, info.Mountpoints)
	}

	isoSize, err := ReadISOSize(opts.ISOPath)
	if err != nil {
		return err
	}

	fmt.Fprintf(opts.Stdout, "About to write to:\n")
	fmt.Fprintf(opts.Stdout, "  Device:  %s\n", device)
	fmt.Fprintf(opts.Stdout, "  Model:   %s\n", info.Model)
	fmt.Fprintf(opts.Stdout, "  Size:    %s\n", info.Size)
	fmt.Fprintf(opts.Stdout, "  Source:  %s (%s)\n\n", opts.ISOPath, isoSize)
	fmt.Fprintf(opts.Stdout, "⚠️  ALL DATA ON %s WILL BE DESTROYED.\n\n", device)

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

	if err := unmountDevice(ctx, opts.Runner, device); err != nil {
		return err
	}

	args := []string{
		"if=" + opts.ISOPath,
		"of=" + device,
		"bs=" + opts.BlockSize,
		"status=progress",
		"conv=fdatasync",
		"oflag=direct",
	}
	if err := opts.Runner.Run(ctx, "dd", args...); err != nil {
		return fmt.Errorf("dd: %w", err)
	}
	if err := opts.Runner.Run(ctx, "sync"); err != nil {
		return err
	}

	fmt.Fprintln(opts.Stdout, "Done. You can now unplug the USB drive and boot from it.")
	return nil
}
