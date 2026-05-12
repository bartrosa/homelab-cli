package commands

import "github.com/spf13/cobra"

// NewDataCmd wires dataset utilities.
func NewDataCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "data",
		Short: "Datasets and lightweight data pipelines",
		Long:  "Helpers for DVC, lakeFS, Parquet conversion, and ad-hoc data prep.",
		RunE: func(c *cobra.Command, _ []string) error {
			return c.Help()
		},
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:     "sync",
			Short:   "Synchronize tracked datasets according to config",
			Example: "  lab data sync",
			Args:    cobra.NoArgs,
			RunE:    StubRunE(),
		},
	)

	return cmd
}
