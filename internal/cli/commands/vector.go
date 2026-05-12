package commands

import "github.com/spf13/cobra"

// NewVectorCmd wires vector database helpers.
func NewVectorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vector",
		Short: "Vector database lifecycle (Qdrant, Weaviate, Milvus, Chroma)",
		Long:  "Provision, snapshot, and validate vector stores used by RAG stacks.",
		RunE: func(c *cobra.Command, _ []string) error {
			return c.Help()
		},
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:     "list",
			Short:   "List configured vector stores and connection health",
			Example: "  lab vector list",
			Args:    cobra.NoArgs,
			RunE:    StubRunE(),
		},
	)

	return cmd
}
