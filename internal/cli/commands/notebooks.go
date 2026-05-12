package commands

import "github.com/spf13/cobra"

// NewNotebooksCmd wires notebook servers.
func NewNotebooksCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "notebooks",
		Short: "Notebook servers (Jupyter, JupyterLab, marimo)",
		Long:  "Launch and manage local notebook environments with homelab defaults.",
		RunE: func(c *cobra.Command, _ []string) error {
			return c.Help()
		},
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:     "up",
			Short:   "Start the configured notebook server",
			Example: "  lab notebooks up",
			Args:    cobra.NoArgs,
			RunE:    StubRunE(),
		},
	)

	return cmd
}
