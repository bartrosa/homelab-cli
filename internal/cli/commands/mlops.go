package commands

import "github.com/spf13/cobra"

// NewMLOpsCmd wires experiment tracking integrations.
func NewMLOpsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mlops",
		Short: "Experiment tracking and model registries",
		Long:  "Integrates with MLflow, Weights & Biases, and similar systems.",
		RunE: func(c *cobra.Command, _ []string) error {
			return c.Help()
		},
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:     "status",
			Short:   "Show configured MLOps endpoints and auth health",
			Example: "  lab mlops status",
			Args:    cobra.NoArgs,
			RunE:    StubRunE(),
		},
	)

	return cmd
}
