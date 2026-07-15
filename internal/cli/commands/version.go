package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bartrosa/homelab-cli/internal/buildinfo"
	"github.com/bartrosa/homelab-cli/internal/logging"

	"github.com/charmbracelet/lipgloss"
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
	out := cmd.OutOrStdout()

	noColor, err := cmd.Root().PersistentFlags().GetBool("no-color")
	if err != nil {
		noColor = false
	}

	title := "lab"
	if !noColor {
		title = lipgloss.NewStyle().Bold(true).Render("lab")
	}

	_, werr := fmt.Fprintf(out, "%s version\n", title)
	if werr != nil {
		return werr
	}

	lines := []string{
		fmt.Sprintf("Version:   %s", info.Version),
		fmt.Sprintf("Commit:    %s", info.Commit),
		fmt.Sprintf("Date:      %s", info.Date),
		fmt.Sprintf("GoVersion: %s", info.GoVersion),
	}

	for _, line := range lines {
		if _, err := fmt.Fprintln(out, line); err != nil {
			return err
		}
	}

	return nil
}
