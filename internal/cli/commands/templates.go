package commands

import "github.com/spf13/cobra"

// NewTemplatesCmd wires project scaffolding.
func NewTemplatesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "templates",
		Short: "Generate projects from curated templates",
		Long:  "Scaffold ML services, APIs, CLIs, and Terraform modules using Go templates.",
		RunE: func(c *cobra.Command, _ []string) error {
			return c.Help()
		},
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:     "new <name>",
			Short:   "Create a new project from the selected template",
			Example: "  lab templates new fraud-detector",
			Args:    cobra.ExactArgs(1),
			RunE:    StubRunE(),
		},
	)

	return cmd
}
