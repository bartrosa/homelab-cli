package commands

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bartrosa/homelab-cli/internal/homelabroot"
	"github.com/bartrosa/homelab-cli/internal/mlstack"
	"github.com/bartrosa/homelab-cli/internal/services"
	"github.com/bartrosa/homelab-cli/internal/ui"
	"github.com/spf13/cobra"
)

// NewServicesCmd wires local compose stacks for databases and brokers.
func NewServicesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "services",
		Short: "Run homelab data services via compose (docker/podman)",
		Long: `services manages compose stacks from your homelab repo (e.g. ml-stack).
Set homelab.root in config or LAB_HOMELAB_ROOT.`,
		RunE: func(c *cobra.Command, _ []string) error {
			return c.Help()
		},
	}
	AddDryRunFlag(cmd)

	cmd.AddCommand(
		&cobra.Command{
			Use:     "up <name> [more...]",
			Short:   "Start one or more service stacks",
			Example: "  lab services up ml-stack",
			Args:    cobra.MinimumNArgs(1),
			RunE:    servicesRunE("up"),
		},
		&cobra.Command{
			Use:     "down <name> [more...]",
			Short:   "Stop one or more service stacks",
			Example: "  lab services down ml-stack",
			Args:    cobra.MinimumNArgs(1),
			RunE:    servicesRunE("down"),
		},
		&cobra.Command{
			Use:     "list",
			Short:   "List available stacks and their runtime status",
			Example: "  lab services list",
			Args:    cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				s := session(cmd)
				sr := services.NewRunner(stdout(cmd), stderr(cmd), s.HomelabRoot, s.Config.Services.Runtime, s.DryRun)
				ui.Section(stdout(cmd), s.Styles, "Stacks", fmt.Sprintf("runtime: %s", s.Config.Services.Runtime))
				var rows [][]string
				for _, st := range sr.List() {
					rows = append(rows, []string{st.Name, st.ComposeFile})
				}
				ui.Table(stdout(cmd), s.Styles, []string{"NAME", "COMPOSE"}, rows)
				return nil
			},
		},
		&cobra.Command{
			Use:     "ensure [ml-stack]",
			Short:   "Ensure ml-stack is up (podman-compose up -d)",
			Example: "  lab services ensure",
			Args:    cobra.MaximumNArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				setDryRun(cmd)
				s := session(cmd)
				root, err := homelabroot.Resolve(firstCLI(s.HomelabRoot, s.Config.Homelab.Root))
				if err != nil {
					return err
				}
				name := "ml-stack"
				if len(args) == 1 {
					name = args[0]
				}
				if name != "ml-stack" {
					return fmt.Errorf("only ml-stack is supported (got %q)", name)
				}
				ui.Section(stdout(cmd), s.Styles, "services ensure", name)
				ip := s.Config.Server.Host
				mlDir := filepath.Join(root, "ml-stack")
				return mlstack.EnsureUp(cmd.Context(), mlDir, ip, s.DryRun, stdout(cmd), stderr(cmd))
			},
		},
		&cobra.Command{
			Use:     "logs <name>",
			Short:   "Tail logs for a running stack",
			Example: "  lab services logs ml-stack",
			Args:    cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				setDryRun(cmd)
				s := session(cmd)
				sr := services.NewRunner(stdout(cmd), stderr(cmd), s.HomelabRoot, s.Config.Services.Runtime, s.DryRun)
				return sr.Logs(cmd.Context(), args[0])
			},
		},
	)

	return cmd
}

func servicesRunE(action string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		setDryRun(cmd)
		s := session(cmd)
		sr := services.NewRunner(stdout(cmd), stderr(cmd), s.HomelabRoot, s.Config.Services.Runtime, s.DryRun)
		ui.Section(stdout(cmd), s.Styles, "services "+action, strings.Join(args, ", "))
		switch action {
		case "up":
			return sr.Up(cmd.Context(), args...)
		case "down":
			return sr.Down(cmd.Context(), args...)
		default:
			return fmt.Errorf("unknown action %s", action)
		}
	}
}
