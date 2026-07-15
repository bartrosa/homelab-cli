package commands

import (
	"github.com/bartrosa/homelab-cli/internal/buildinfo"
	"github.com/bartrosa/homelab-cli/internal/clierrors"
	"github.com/bartrosa/homelab-cli/internal/updater"
	"github.com/spf13/cobra"
)

// NewSelfUpdateCmd wires lab self-update.
func NewSelfUpdateCmd() *cobra.Command {
	var (
		checkOnly  bool
		version    string
		preRelease bool
		yes        bool
	)

	cmd := &cobra.Command{
		Use:   "self-update",
		Short: "Download and install the latest lab release from GitHub",
		Long: `Checks GitHub releases for a newer lab binary, verifies SHA256 checksums,
and atomically replaces the running executable.`,
		Example: `  lab self-update
  lab self-update --check
  lab self-update --version v0.2.0 --yes`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			code, err := updater.PerformUpdate(cmd.Context(), buildinfo.Version, updater.UpdateOptions{
				ForceVersion:      version,
				IncludePrerelease: preRelease,
				Yes:               yes,
				CheckOnly:         checkOnly,
				Stdout:            stdout(cmd),
				Stderr:            stderr(cmd),
			})
			if err != nil {
				return err
			}
			if checkOnly && code == 3 {
				return &clierrors.ExitError{Code: 3}
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&checkOnly, "check", false, "only check for updates (exit 0 if up to date, 3 if update available)")
	cmd.Flags().StringVar(&version, "version", "", "install a specific release tag")
	cmd.Flags().BoolVar(&preRelease, "pre-release", false, "include pre-releases when checking latest")
	cmd.Flags().BoolVar(&yes, "yes", false, "skip confirmation prompts")

	return cmd
}
