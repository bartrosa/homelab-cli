package commands

import "github.com/spf13/cobra"

// NewAgentsCmd wires local AI agent runtimes.
func NewAgentsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agents",
		Short: "Local AI agent runtimes (LangGraph, CrewAI, MCP clients)",
		Long:  "Manage long-running agent processes and their dependencies.",
		RunE: func(c *cobra.Command, _ []string) error {
			return c.Help()
		},
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:     "list",
			Short:   "List registered agents and their schedules",
			Example: "  lab agents list",
			Args:    cobra.NoArgs,
			RunE:    StubRunE(),
		},
	)

	return cmd
}
