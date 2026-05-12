package commands

import "github.com/spf13/cobra"

// NewModelsCmd wires local LLM workflows.
func NewModelsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "models",
		Short: "Local LLMs (Ollama, vLLM, Hugging Face cache, llama.cpp)",
		Long:  "Pull, list, run, and benchmark models on homelab GPUs.",
		RunE: func(c *cobra.Command, _ []string) error {
			return c.Help()
		},
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:     "pull <name>",
			Short:   "Download a model into the configured cache/runtime",
			Example: "  lab models pull llama3",
			Args:    cobra.ExactArgs(1),
			RunE:    StubRunE(),
		},
	)

	return cmd
}
