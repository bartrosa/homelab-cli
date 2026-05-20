// Package system provides host utilities (bootable USB, etc.).
package system

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/bartrosa/homelab-cli/internal/executil"
)

// USBOptions configures bootable USB creation.
type USBOptions struct {
	WorkDir string
	Device  string
	Distro  string
	ISOURL  string // bypass discovery
	DryRun  bool
}

// CreateBootableUSB downloads an ISO, verifies checksum when available, writes to a block device.
func CreateBootableUSB(ctx context.Context, opts USBOptions, stdout, stderr io.Writer) error {
	for _, bin := range []string{"wget", "sha256sum", "dd", "sudo"} {
		if !executil.CommandExists(bin) {
			return fmt.Errorf("%s not found in PATH", bin)
		}
	}

	var spec isoSpec
	var label string
	if opts.ISOURL != "" {
		spec = isoSpec{
			isoURL:  opts.ISOURL,
			isoFile: filepath.Base(strings.Split(opts.ISOURL, "?")[0]),
		}
		label = spec.isoFile
	} else {
		var err error
		spec, label, err = ResolveBootImage(ctx, opts.Distro)
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "Wybrane: %s\n", label)
	}

	work := opts.WorkDir
	if work == "" {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		work = wd
	}
	if err := os.MkdirAll(work, 0o755); err != nil {
		return err
	}

	ex := executil.NewRunner(stdout, stderr)
	ex.DryRun = opts.DryRun
	ex.WorkDir = work

	isoPath := filepath.Join(work, spec.isoFile)
	fmt.Fprintf(stdout, "Pobieranie ISO: %s\n", spec.isoFile)
	if err := ex.Run(ctx, "wget", "-c", "-O", spec.isoFile, spec.isoURL); err != nil {
		return fmt.Errorf("download iso: %w", err)
	}

	if spec.checksumURL != "" {
		fmt.Fprintln(stdout, "Pobieranie checksumów...")
		_ = ex.Run(ctx, "wget", "-q", "-O", spec.checksumFile, spec.checksumURL)
		if spec.fedoraStyle {
			_ = ex.Run(ctx, "bash", "-c", fmt.Sprintf(
				`base=$(basename %q); hash=$(grep -F "($base)" %q | sed -n 's/.*= \\([a-fA-F0-9]*\\).*/\\1/p' | tr -d ' ')
test -n "$hash" && test "$(sha256sum %q | awk '{print $1}')" = "$hash"`, spec.isoFile, spec.checksumFile, spec.isoFile))
		} else {
			_ = ex.Run(ctx, "sha256sum", "-c", spec.checksumFile, "--ignore-missing")
		}
	}

	device := strings.TrimSpace(opts.Device)
	if device == "" {
		return fmt.Errorf("podaj --device (np. /dev/sdb); podgląd: lsblk -d -o NAME,SIZE,MODEL,TRAN")
	}
	fmt.Fprintf(stdout, "Zapis ISO → %s\n", device)
	if opts.DryRun {
		fmt.Fprintf(stdout, "[dry-run] sudo dd if=%s of=%s bs=4M status=progress oflag=sync\n", isoPath, device)
		return nil
	}
	script := fmt.Sprintf(`set -euo pipefail
[[ -b %q ]] || { echo "not a block device"; exit 1; }
read -r -p "Zapisać %q na %q? Wpisz yes: " confirm
[[ "$confirm" == "yes" ]] || exit 0
sudo dd if=%q of=%q bs=4M status=progress oflag=sync
sync
echo "Gotowe."
`, device, spec.isoFile, device, isoPath, device)
	return ex.Run(ctx, "bash", "-c", script)
}

type isoSpec struct {
	isoURL       string
	checksumURL  string
	isoFile      string
	checksumFile string
	fedoraStyle  bool
}
