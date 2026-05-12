package commands

import "github.com/spf13/cobra"

// NewPipelinesCmd wires local training/inference pipelines.
func NewPipelinesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pipelines",
		Short: "Local pipeline runner for training, inference, and ETL",
		Long:  "Executes declarative pipeline specs on homelab hardware without a remote CI engine.",
		RunE: func(c *cobra.Command, _ []string) error {
			return c.Help()
		},
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:     "run <name>",
			Short:   "Run a named pipeline definition from disk",
			Example: "  lab pipelines run nightly-embeddings",
			Args:    cobra.ExactArgs(1),
			RunE:    StubRunE(),
		},
	)

	return cmd
}
