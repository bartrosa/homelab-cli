package commands

import "github.com/spf13/cobra"

// NewLogsCmd wires aggregated log access.
func NewLogsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Aggregate logs from homelab services",
		Long:  "Streams logs from local compose stacks and remote Loki endpoints.",
		RunE: func(c *cobra.Command, _ []string) error {
			return c.Help()
		},
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:     "tail <selector>",
			Short:   "Tail logs matching a selector (service, host, or query)",
			Example: "  lab logs tail postgres",
			Args:    cobra.ExactArgs(1),
			RunE:    StubRunE(),
		},
	)

	return cmd
}
