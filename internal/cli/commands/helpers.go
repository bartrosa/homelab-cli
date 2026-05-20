package commands

import (
	"io"

	"github.com/bartrosa/homelab-cli/internal/cli/appctx"
	"github.com/spf13/cobra"
)

func session(cmd *cobra.Command) *appctx.Session {
	return appctx.MustSession(cmd.Context())
}

func setDryRun(cmd *cobra.Command) {
	s := session(cmd)
	if cmd.Flags().Changed("dry-run") {
		v, _ := cmd.Flags().GetBool("dry-run")
		s.DryRun = v
	}
}

func stdout(cmd *cobra.Command) io.Writer {
	return cmd.OutOrStdout()
}

func stderr(cmd *cobra.Command) io.Writer {
	return cmd.ErrOrStderr()
}

// AddDryRunFlag registers --dry-run on a command.
func AddDryRunFlag(cmd *cobra.Command) {
	cmd.Flags().Bool("dry-run", false, "print planned actions without executing")
}
