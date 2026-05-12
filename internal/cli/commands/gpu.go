package commands

import "github.com/spf13/cobra"

// NewGPUCmd wires GPU diagnostics and driver helpers.
func NewGPUCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gpu",
		Short: "GPU, CUDA, and driver diagnostics",
		Long:  "Inspect GPUs, driver versions, and container runtime integrations.",
		RunE: func(c *cobra.Command, _ []string) error {
			return c.Help()
		},
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:     "info",
			Short:   "Print GPU inventory and driver capabilities",
			Example: "  lab gpu info",
			Args:    cobra.NoArgs,
			RunE:    StubRunE(),
		},
	)

	return cmd
}
