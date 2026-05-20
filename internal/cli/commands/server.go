package commands

import (
	"os"
	"strings"

	"github.com/bartrosa/homelab-cli/internal/homelabroot"
	"github.com/bartrosa/homelab-cli/internal/server"
	"github.com/bartrosa/homelab-cli/internal/ui"
	"github.com/spf13/cobra"
)

// NewServerCmd wires remote homelab server operations.
func NewServerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Remote homelab server (rsync, ssh run, deploy)",
		Long:  "Uses server.* from config (host, user, port, path). No homelab shell scripts.",
		RunE: func(c *cobra.Command, _ []string) error {
			return c.Help()
		},
	}
	AddDryRunFlag(cmd)

	cmd.AddCommand(
		&cobra.Command{
			Use:     "run <command>",
			Short:   "Run a shell command on the server in server.path",
			Example: `  lab server run 'cd ml-stack && podman-compose ps'`,
			Args:    cobra.MinimumNArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				setDryRun(cmd)
				s := session(cmd)
				if s.DryRun {
					ui.Section(stdout(cmd), s.Styles, "server run (dry-run)", strings.Join(args, " "))
					return nil
				}
				t, err := server.TargetFromConfig(s.Config)
				if err != nil {
					return err
				}
				remote := strings.Join(args, " ")
				ui.Section(stdout(cmd), s.Styles, "server run", remote)
				return server.Run(cmd.Context(), t, remote, os.Stdin, stdout(cmd), stderr(cmd))
			},
		},
		&cobra.Command{
			Use:   "deploy [sync|provision|compose|full]",
			Short: "Rsync homelab to server; optionally provision PG or start ml-stack",
			Long: `Modes:
  (none)     sync only
  provision  sync + lab postgres apply (local, against PG in instances.yaml)
  compose    sync + podman-compose up -d on server
  full       provision + compose`,
			Example: `  lab server deploy
  lab server deploy full`,
			Args: cobra.MaximumNArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				setDryRun(cmd)
				s := session(cmd)
				root, err := homelabroot.Resolve(firstCLI(s.HomelabRoot, s.Config.Homelab.Root))
				if err != nil {
					return err
				}
				mode := server.DeploySync
				if len(args) == 1 {
					mode = server.DeployMode(args[0])
				}
				ui.Section(stdout(cmd), s.Styles, "server deploy", string(mode))
				return server.Deploy(cmd.Context(), s.Config, root, mode, s.DryRun, stdout(cmd), stderr(cmd))
			},
		},
	)

	return cmd
}

func firstCLI(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
