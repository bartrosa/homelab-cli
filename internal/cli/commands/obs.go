package commands

import (
	"github.com/bartrosa/homelab-cli/internal/services"
	"github.com/bartrosa/homelab-cli/internal/ui"
	"github.com/spf13/cobra"
)

// NewObsCmd wires observability stacks (wrapper on lab services).
func NewObsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "obs",
		Short: "Observability stacks (Prometheus, Grafana, Loki, Tempo)",
		Long:  "Launch curated observability bundles via lab services presets.",
		RunE: func(c *cobra.Command, _ []string) error {
			return c.Help()
		},
	}
	AddDryRunFlag(cmd)

	cmd.AddCommand(
		&cobra.Command{
			Use:     "up",
			Short:   "Start the observability stack",
			Example: "  lab obs up",
			Args:    cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				setDryRun(cmd)
				s := session(cmd)
				ui.Section(stdout(cmd), s.Styles, "obs up", "preset observability")
				o := &services.Orchestrator{CustomPresets: s.Config.Services.Presets}
				opts := svcOpts(cmd)
				opts.NonInteractive = true
				return o.Up(cmd.Context(), opts, "observability")
			},
		},
		&cobra.Command{
			Use:   "down",
			Short: "Stop observability services",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				o := &services.Orchestrator{CustomPresets: session(cmd).Config.Services.Presets}
				return o.Down(cmd.Context(), svcOpts(cmd), "observability")
			},
		},
	)

	return cmd
}
