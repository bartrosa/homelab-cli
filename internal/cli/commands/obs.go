package commands

import "github.com/spf13/cobra"

// NewObsCmd wires observability stacks.
func NewObsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "obs",
		Short: "Observability stacks (Prometheus, Grafana, Loki, Tempo, OTel)",
		Long:  "Launch curated observability bundles tailored for homelab services.",
		RunE: func(c *cobra.Command, _ []string) error {
			return c.Help()
		},
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:     "up",
			Short:   "Start the observability stack",
			Example: "  lab obs up",
			Args:    cobra.NoArgs,
			RunE:    StubRunE(),
		},
	)

	return cmd
}
