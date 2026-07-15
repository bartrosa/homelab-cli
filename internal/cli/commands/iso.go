package commands

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/bartrosa/homelab-cli/internal/exec"
	"github.com/bartrosa/homelab-cli/internal/iso"
	"github.com/bartrosa/homelab-cli/internal/ui"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

// NewISOCmd wires lab iso subcommands.
func NewISOCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "iso",
		Short: "Download OS ISOs and create bootable USB drives",
		Long: `Provisioning flow:
  1. lab iso list       — see supported distributions
  2. lab iso download   — fetch and verify an ISO into cache
  3. lab iso disks      — list block devices (USB vs system)
  4. lab iso write      — burn a cached ISO to USB (interactive or by name)

Quick burn:
  lab iso write                  pick image + USB drive interactively
  lab iso write ubuntu-desktop --usb`,
		RunE: func(c *cobra.Command, _ []string) error {
			return c.Help()
		},
	}

	cmd.AddCommand(
		newISOListCmd(),
		newISODownloadCmd(),
		newISOImagesCmd(),
		newISODisksCmd(),
		newISOWriteCmd(),
	)

	return cmd
}

func newISOListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List supported distributions and current versions",
		RunE: func(cmd *cobra.Command, _ []string) error {
			entries, err := iso.ListDistros()
			if err != nil {
				return err
			}
			s := session(cmd)
			w := stdout(cmd)
			ui.Section(w, s.Styles, "ISO catalog", "Installer images from upstream mirrors")
			headers := []string{"DISTRO", "VERSION", "SIZE", "ARCHITECTURES"}
			rows := make([][]string, len(entries))
			for i, e := range entries {
				ver := e.Version
				if ver == "planned" {
					ver = s.Styles.Dim.Render("planned")
				}
				rows[i] = []string{e.ID, ver, e.ApproxSize, e.Architectures}
			}
			ui.Table(w, s.Styles, headers, rows)
			return nil
		},
	}
}

func newISODownloadCmd() *cobra.Command {
	var (
		arch     string
		version  string
		output   string
		noVerify bool
		force    bool
	)

	cmd := &cobra.Command{
		Use:     "download <distro>",
		Short:   "Download and verify an ISO",
		Example: "  lab iso download ubuntu-desktop\n  lab iso download fedora-silverblue --arch amd64",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			d, ok := iso.LookupDistro(args[0])
			if !ok {
				return fmt.Errorf("unknown distro %q (run: lab iso list)", args[0])
			}
			if output == "" {
				home, _ := os.UserHomeDir()
				output = filepath.Join(home, ".cache", "homelab-cli", "iso")
			}
			_, err := iso.DownloadISO(cmd.Context(), d, iso.DownloadOptions{
				Arch:      arch,
				Version:   version,
				OutputDir: output,
				NoVerify:  noVerify,
				Force:     force,
				NoColor:   session(cmd).NoColor,
				Stdout:    stdout(cmd),
			})
			return err
		},
	}

	cmd.Flags().StringVar(&arch, "arch", "amd64", "target architecture (amd64|arm64)")
	cmd.Flags().StringVar(&version, "version", "", "pin version (default: latest resolver)")
	cmd.Flags().StringVar(&output, "output", "", "cache directory (default: ~/.cache/homelab-cli/iso/)")
	cmd.Flags().BoolVar(&noVerify, "no-verify", false, "skip GPG verification of checksums")
	cmd.Flags().BoolVar(&force, "force", false, "redownload even if cached file exists")

	return cmd
}

func newISODisksCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "disks",
		Short: "List block devices and mark USB vs system disks",
		RunE: func(cmd *cobra.Command, _ []string) error {
			runner := exec.NewOSRunner(stdout(cmd), stderr(cmd))
			disks, err := iso.ListDisks(cmd.Context(), runner)
			if err != nil {
				return err
			}
			s := session(cmd)
			w := stdout(cmd)
			ui.Section(w, s.Styles, "Block devices", "Write only to USB — never to SYSTEM disks")
			headers := []string{"DEVICE", "SIZE", "MODEL", "TRAN", "TYPE"}
			rows := make([][]string, len(disks))
			for i, d := range disks {
				typ := string(d.Type)
				switch d.Type {
				case iso.DiskUSB:
					typ = s.Styles.OK.Render("USB")
				case iso.DiskSystem:
					typ = s.Styles.Warn.Render("SYSTEM")
				}
				rows[i] = []string{d.Device, d.Size, d.Model, d.Tran, typ}
			}
			ui.Table(w, s.Styles, headers, rows)
			if len(disks) == 0 {
				ui.Warn(w, s.Styles, "no block devices found (is a USB drive connected?)")
			}
			for _, d := range disks {
				if d.Type == iso.DiskSystem {
					ui.Warn(w, s.Styles, "do not write ISOs to "+d.Device)
				}
			}
			return nil
		},
	}
}

func newISOImagesCmd() *cobra.Command {
	var cacheDir string
	cmd := &cobra.Command{
		Use:   "images",
		Short: "List ISO files in the local cache",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if cacheDir == "" {
				var err error
				cacheDir, err = iso.DefaultCacheDir()
				if err != nil {
					return err
				}
			}
			images, err := iso.ListCachedImages(cacheDir)
			if err != nil {
				return err
			}
			s := session(cmd)
			w := stdout(cmd)
			ui.Section(w, s.Styles, "Cached ISO images", cacheDir)
			if len(images) == 0 {
				ui.Warn(w, s.Styles, "no images yet — run: lab iso download ubuntu-desktop")
				return nil
			}
			headers := []string{"FILE", "SIZE"}
			rows := make([][]string, len(images))
			for i, img := range images {
				rows[i] = []string{img.Name, img.Size}
			}
			ui.Table(w, s.Styles, headers, rows)
			return nil
		},
	}
	cmd.Flags().StringVar(&cacheDir, "cache", "", "ISO cache directory (default: ~/.cache/homelab-cli/iso/)")
	return cmd
}

func newISOWriteCmd() *cobra.Command {
	var (
		device    string
		usb       bool
		yes       bool
		force     bool
		blockSize string
		cacheDir  string
	)

	cmd := &cobra.Command{
		Use:   "write [distro|iso-file]",
		Short: "Write a cached ISO to a USB drive",
		Long: `Burn an installer ISO to a USB stick.

Without arguments, pick interactively from cached images and USB drives.

Examples:
  lab iso write
  lab iso write ubuntu-desktop --usb
  lab iso write ubuntu-desktop --to /dev/sda
  lab iso write ~/.cache/homelab-cli/iso/ubuntu-24.04.4-desktop-amd64.iso --to sda`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if cacheDir == "" {
				var err error
				cacheDir, err = iso.DefaultCacheDir()
				if err != nil {
					return err
				}
			}

			runner := exec.NewOSRunner(stdout(cmd), stderr(cmd))
			in := cmd.InOrStdin()
			interactive := terminalInteractive(cmd)

			isoPath, err := resolveWriteISO(args, cacheDir, in, stdout(cmd), interactive)
			if err != nil {
				return err
			}

			target, err := resolveWriteDevice(cmd.Context(), runner, device, usb, in, stdout(cmd), interactive)
			if err != nil {
				return err
			}

			return iso.WriteISO(cmd.Context(), iso.WriteOptions{
				ISOPath:   isoPath,
				Device:    target,
				Yes:       yes,
				Force:     force,
				NoColor:   session(cmd).NoColor,
				BlockSize: blockSize,
				Stdout:    stdout(cmd),
				Stderr:    stderr(cmd),
				Runner:    runner,
			})
		},
	}

	cmd.Flags().StringVar(&device, "to", "", "target block device (e.g. /dev/sda or sda)")
	cmd.Flags().StringVar(&device, "device", "", "alias for --to")
	cmd.Flags().BoolVar(&usb, "usb", false, "use the only connected USB drive (or pick if several)")
	cmd.Flags().BoolVar(&yes, "yes", false, "skip confirmation prompt")
	cmd.Flags().BoolVar(&force, "force", false, "skip system-disk safety check (DANGER)")
	cmd.Flags().StringVar(&blockSize, "bs", "4M", "dd block size")
	cmd.Flags().StringVar(&cacheDir, "cache", "", "ISO cache directory (default: ~/.cache/homelab-cli/iso/)")

	return cmd
}

func resolveWriteISO(args []string, cacheDir string, in io.Reader, w io.Writer, interactive bool) (string, error) {
	if len(args) == 1 {
		return iso.ResolveISORef(args[0], cacheDir)
	}

	images, err := iso.ListCachedImages(cacheDir)
	if err != nil {
		return "", err
	}
	if len(images) == 0 {
		return "", fmt.Errorf("no cached ISO images in %s (run: lab iso download ubuntu-desktop)", cacheDir)
	}
	if !interactive {
		return "", fmt.Errorf("specify an image: lab iso write ubuntu-desktop (or: lab iso images)")
	}

	opts := make([]string, len(images))
	for i, img := range images {
		opts[i] = fmt.Sprintf("%s  (%s)", img.Name, img.Size)
	}
	idx, err := promptChoice(in, w, "Select ISO to write:", opts, 0)
	if err != nil {
		return "", err
	}
	return images[idx].Path, nil
}

func resolveWriteDevice(ctx context.Context, runner exec.Runner, device string, usbAuto bool, in io.Reader, w io.Writer, interactive bool) (string, error) {
	device = normalizeDevice(device)
	if device != "" {
		return device, nil
	}

	disks, err := iso.ListDisks(ctx, runner)
	if err != nil {
		return "", err
	}
	var usbs []iso.Disk
	for _, d := range disks {
		if d.Type == iso.DiskUSB {
			usbs = append(usbs, d)
		}
	}

	if usbAuto {
		switch len(usbs) {
		case 0:
			return "", fmt.Errorf("no USB drives found (plug in a stick and run: lab iso disks)")
		case 1:
			fmt.Fprintf(w, "Using USB drive %s (%s)\n", usbs[0].Device, usbs[0].Model)
			return usbs[0].Device, nil
		}
	}

	pickFrom := usbs
	title := "Select USB drive:"
	if len(usbs) == 0 {
		if !interactive {
			return "", fmt.Errorf("--to is required (no USB drives detected; run: lab iso disks)")
		}
		pickFrom = disks
		title = "No USB drives detected — select device (CAUTION):"
	}

	if !interactive {
		if usbAuto && len(usbs) > 1 {
			return "", fmt.Errorf("multiple USB drives — pick one with --to (run: lab iso disks)")
		}
		return "", fmt.Errorf("--to or --usb is required (run: lab iso disks)")
	}

	opts := make([]string, len(pickFrom))
	for i, d := range pickFrom {
		opts[i] = fmt.Sprintf("%s  %s  %s  [%s]", d.Device, d.Size, d.Model, d.Type)
	}
	idx, err := promptChoice(in, w, title, opts, 0)
	if err != nil {
		return "", err
	}
	return pickFrom[idx].Device, nil
}

func terminalInteractive(cmd *cobra.Command) bool {
	in, okIn := cmd.InOrStdin().(*os.File)
	out, okOut := cmd.OutOrStdout().(*os.File)
	return okIn && okOut && isatty.IsTerminal(in.Fd()) && isatty.IsTerminal(out.Fd())
}

func normalizeDevice(d string) string {
	return iso.BlockDevicePath(d)
}
