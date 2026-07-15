package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bartrosa/homelab-cli/internal/exec"
	"github.com/bartrosa/homelab-cli/internal/iso"
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
  4. lab iso write      — burn ISO to a USB drive`,
		RunE: func(c *cobra.Command, _ []string) error {
			return c.Help()
		},
	}

	cmd.AddCommand(
		newISOListCmd(),
		newISODownloadCmd(),
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
			return iso.WriteList(stdout(cmd), entries)
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
			iso.WriteDisksTable(stdout(cmd), disks)
			return nil
		},
	}
}

func newISOWriteCmd() *cobra.Command {
	var (
		device    string
		yes       bool
		force     bool
		blockSize string
	)

	cmd := &cobra.Command{
		Use:     "write <iso-path>",
		Short:   "Write an ISO to a block device",
		Example: "  lab iso write ~/.cache/homelab-cli/iso/ubuntu-24.04.3-desktop-amd64.iso --to /dev/sdb",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if device == "" {
				return fmt.Errorf("--to is required (see: lab iso disks)")
			}
			isoPath, err := expandPath(args[0])
			if err != nil {
				return err
			}
			return iso.WriteISO(cmd.Context(), iso.WriteOptions{
				ISOPath:   isoPath,
				Device:    device,
				Yes:       yes,
				Force:     force,
				BlockSize: blockSize,
				Stdout:    stdout(cmd),
				Stderr:    stderr(cmd),
				Runner:    exec.NewOSRunner(stdout(cmd), stderr(cmd)),
			})
		},
	}

	cmd.Flags().StringVar(&device, "to", "", "target block device (e.g. /dev/sdb)")
	cmd.Flags().BoolVar(&yes, "yes", false, "skip confirmation prompt")
	cmd.Flags().BoolVar(&force, "force", false, "skip system-disk safety check (DANGER)")
	cmd.Flags().StringVar(&blockSize, "bs", "4M", "dd block size")

	return cmd
}

func expandPath(p string) (string, error) {
	if len(p) >= 2 && p[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, p[2:]), nil
	}
	return p, nil
}
