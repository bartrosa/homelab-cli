package commands

import (
	"fmt"

	"github.com/bartrosa/homelab-cli/internal/media"
	"github.com/bartrosa/homelab-cli/internal/ui"
	"github.com/spf13/cobra"
)

// NewMediaCmd wires media utilities (HEIC conversion).
func NewMediaCmd() *cobra.Command {
	var (
		heicQuality int
		heicForce   bool
	)

	cmd := &cobra.Command{
		Use:   "media",
		Short: "Media conversion helpers",
		Long:  `media: HEIC→JPEG via heif-convert.`,
		RunE: func(c *cobra.Command, _ []string) error {
			return c.Help()
		},
	}
	AddDryRunFlag(cmd)

	heic := &cobra.Command{
		Use:     "heic [directory]",
		Short:   "Convert .HEIC photos to JPEG in a directory",
		Long:    `Converts each .HEIC/.heic file to .jpg using heif-convert (quality 1–100). Skips when .jpg already exists unless --force.`,
		Example: `  lab media heic .
  lab media heic ~/Pictures/import --quality 95`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			setDryRun(cmd)
			s := session(cmd)
			dir := "."
			if len(args) == 1 {
				dir = args[0]
			}
			ui.Section(stdout(cmd), s.Styles, "media heic", dir)
			n, err := media.ConvertHEIC(cmd.Context(), stdout(cmd), stderr(cmd), media.HEICOptions{
				Dir: dir, Quality: heicQuality, Force: heicForce, DryRun: s.DryRun,
			})
			if err != nil {
				return err
			}
			ui.OK(stdout(cmd), s.Styles, fmt.Sprintf("converted %d file(s)", n))
			return nil
		},
	}
	heic.Flags().IntVar(&heicQuality, "quality", 100, "JPEG quality (1–100)")
	heic.Flags().BoolVar(&heicForce, "force", false, "reconvert even when output .jpg exists")

	cmd.AddCommand(heic)

	return cmd
}
