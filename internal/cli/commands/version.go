package commands

import (
	"encoding/json"
	"strings"

	"github.com/bartrosa/homelab-cli/internal/buildinfo"
	"github.com/bartrosa/homelab-cli/internal/logging"
	"github.com/bartrosa/homelab-cli/internal/ui"

	"github.com/spf13/cobra"
)

// NewVersionCmd prints build metadata (text or JSON).
func NewVersionCmd() *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:     "version",
		Short:   "Print build information",
		Long:    "Shows version, git commit, build date, and the Go toolchain used to compile the binary.",
		Example: "  lab version\n  lab version --output json",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			logging.LoggerFromContext(cmd.Context()).Debug("version command invoked")

			info := buildinfo.Get()
			switch strings.ToLower(strings.TrimSpace(output)) {
			case "json":
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(info)
			default:
				return printVersionText(cmd, info)
			}
		},
	}

	cmd.Flags().StringVar(&output, "output", "text", "output format (text|json)")

	return cmd
}

func printVersionText(cmd *cobra.Command, info buildinfo.Info) error {
	s := session(cmd)
	ui.KeyValue(stdout(cmd), s.Styles, "lab", [][2]string{
		{"Version", info.Version},
		{"Commit", info.Commit},
		{"Date", info.Date},
		{"GoVersion", info.GoVersion},
	})
	return nil
}
