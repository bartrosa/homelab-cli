package commands

import (
	"fmt"

	"github.com/bartrosa/homelab-cli/internal/templates"
	"github.com/bartrosa/homelab-cli/internal/ui"
	"github.com/spf13/cobra"
)

// NewTemplatesCmd wires project scaffolding from homelab project-initiators.
func NewTemplatesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "templates",
		Short: "Generate projects from homelab templates",
		Long:  `templates copies project-initiators from your homelab repo into a new directory.`,
		RunE: func(c *cobra.Command, _ []string) error {
			return c.Help()
		},
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:     "list",
			Short:   "List available template kinds",
			Example: "  lab templates list",
			Args:    cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				s := session(cmd)
				ui.Section(stdout(cmd), s.Styles, "Templates", "")
				for _, k := range templates.ListKinds() {
					_, _ = fmt.Fprintf(stdout(cmd), "  %s\n", k)
				}
				return nil
			},
		},
		&cobra.Command{
			Use:     "new <kind> <directory>",
			Short:   "Create a project from a template",
			Example: "  lab templates new go ./my-service",
			Args:    cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				s := session(cmd)
				ui.Section(stdout(cmd), s.Styles, "templates new", args[0]+" → "+args[1])
				if err := templates.NewProject(s.HomelabRoot, args[0], args[1], stdout(cmd)); err != nil {
					return err
				}
				ui.OK(stdout(cmd), s.Styles, "scaffolded "+args[1])
				return nil
			},
		},
	)

	return cmd
}
