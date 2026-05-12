package commands

import "github.com/spf13/cobra"

// NewStorageCmd wires S3-compatible storage helpers.
func NewStorageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "storage",
		Short: "S3-compatible storage management (MinIO, SeaweedFS, ...)",
		Long:  "Bucket lifecycle, replication helpers, and quick smoke tests.",
		RunE: func(c *cobra.Command, _ []string) error {
			return c.Help()
		},
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:     "ls",
			Short:   "List buckets or prefixes for the configured endpoint",
			Example: "  lab storage ls s3://models",
			Args:    cobra.MinimumNArgs(1),
			RunE:    StubRunE(),
		},
	)

	return cmd
}
