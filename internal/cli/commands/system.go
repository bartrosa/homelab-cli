package commands

import (
	"github.com/bartrosa/homelab-cli/internal/system"
	"github.com/bartrosa/homelab-cli/internal/ui"
	"github.com/spf13/cobra"
)

// NewSystemCmd wires host-level utilities.
func NewSystemCmd() *cobra.Command {
	var (
		workDir string
		device  string
		distro  string
		isoURL  string
	)

	cmd := &cobra.Command{
		Use:   "system",
		Short: "Host utilities (bootable USB, …)",
		RunE: func(c *cobra.Command, _ []string) error {
			return c.Help()
		},
	}
	AddDryRunFlag(cmd)

	usb := &cobra.Command{
		Use:   "usb",
		Short: "Create a bootable USB from a Linux ISO",
		Long: `Pobiera wersje z upstreamu:
  Ubuntu — changelogs.ubuntu.com (meta-release) + releases.ubuntu.com (SHA256SUMS)
  Fedora Silverblue — download.fedoraproject.org (najnowszy release)

Domyślnie: ubuntu-latest. Użyj "lab system usb list" aby zobaczyć ID.`,
		Example: `  lab system usb list
  lab system usb --distro ubuntu-lts-24.04 --device /dev/sdb
  lab system usb --distro fedora-silverblue --workdir ~/Downloads --device /dev/sdb`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			setDryRun(cmd)
			s := session(cmd)
			ui.Section(stdout(cmd), s.Styles, "system usb", distro)
			return system.CreateBootableUSB(cmd.Context(), system.USBOptions{
				WorkDir: workDir,
				Device:  device,
				Distro:  distro,
				ISOURL:  isoURL,
				DryRun:  s.DryRun,
			}, stdout(cmd), stderr(cmd))
		},
	}
	usb.Flags().StringVar(&workDir, "workdir", "", "directory for ISO download")
	usb.Flags().StringVar(&device, "device", "", "block device (e.g. /dev/sdb)")
	usb.Flags().StringVar(&distro, "distro", "ubuntu-latest", "image ID from 'lab system usb list' or alias: ubuntu-latest, ubuntu-lts, fedora-silverblue")
	usb.Flags().StringVar(&isoURL, "iso-url", "", "custom ISO URL (skips discovery)")

	list := &cobra.Command{
		Use:   "list",
		Short: "List bootable images discovered from Ubuntu and Fedora",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			s := session(cmd)
			ui.Section(stdout(cmd), s.Styles, "system usb list", "querying upstream")
			return system.PrintBootImageList(cmd.Context(), stdout(cmd))
		},
	}

	usb.AddCommand(list)
	cmd.AddCommand(usb)
	return cmd
}
