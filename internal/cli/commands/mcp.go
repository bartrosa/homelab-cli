package commands

import "github.com/spf13/cobra"

// NewMCPCmd exposes the future stdio MCP server used by IDEs.
func NewMCPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Built-in MCP server for IDE integrations",
		Long: `mcp will expose a curated subset of lab commands as Model Context Protocol tools
so assistants in Cursor or VS Code can operate safely against your homelab.`,
		RunE: func(c *cobra.Command, _ []string) error {
			return c.Help()
		},
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:     "serve",
			Short:   "Start the MCP server on stdio",
			Long:    "Streams MCP messages over stdin/stdout for editor integrations.",
			Example: "  lab mcp serve",
			Args:    cobra.NoArgs,
			RunE:    StubRunE(),
		},
	)

	return cmd
}
