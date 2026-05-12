package commands

import "github.com/spf13/cobra"

// NewContainersCmd wires docker/podman housekeeping.
func NewContainersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "containers",
		Short: "Container runtime utilities (Docker and Podman)",
		Long:  "Housekeeping, registry helpers, and cross-runtime diagnostics.",
		RunE: func(c *cobra.Command, _ []string) error {
			return c.Help()
		},
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:     "ps",
			Short:   "List running containers across configured runtimes",
			Example: "  lab containers ps",
			Args:    cobra.NoArgs,
			RunE:    StubRunE(),
		},
	)

	return cmd
}
