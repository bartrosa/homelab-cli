package commands

import (
	"fmt"

	"github.com/bartrosa/homelab-cli/internal/services"
	"github.com/bartrosa/homelab-cli/internal/ui"
	"github.com/spf13/cobra"
)

// NewVectorCmd wires vector database helpers.
func NewVectorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vector",
		Short: "Vector database lifecycle (Qdrant, Weaviate)",
		Long:  "Provision vector stores via lab services.",
		RunE: func(c *cobra.Command, _ []string) error {
			return c.Help()
		},
	}
	AddDryRunFlag(cmd)

	cmd.AddCommand(
		&cobra.Command{
			Use:     "list",
			Short:   "List vector services",
			Example: "  lab vector list",
			Args:    cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				for _, id := range []string{"qdrant", "weaviate"} {
					svc, ok := services.Lookup(id)
					if !ok {
						continue
					}
					st, _ := svc.Status(cmd.Context(), svcOpts(cmd))
					fmt.Fprintf(stdout(cmd), "  %s: running=%v %s\n", id, st.Running, st.Detail)
				}
				return nil
			},
		},
		&cobra.Command{
			Use:     "up <qdrant|weaviate>",
			Short:   "Start a vector database service",
			Example: "  lab vector up qdrant",
			Args:    cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				setDryRun(cmd)
				s := session(cmd)
				ui.Section(stdout(cmd), s.Styles, "vector up", args[0])
				o := &services.Orchestrator{CustomPresets: s.Config.Services.Presets}
				opts := svcOpts(cmd)
				opts.NonInteractive = true
				return o.Up(cmd.Context(), opts, args[0])
			},
		},
	)

	return cmd
}
