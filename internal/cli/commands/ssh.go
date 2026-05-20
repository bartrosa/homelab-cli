package commands

import (
	"github.com/bartrosa/homelab-cli/internal/ssh"
	"github.com/bartrosa/homelab-cli/internal/ui"
	"github.com/spf13/cobra"
)

// NewSSHCmd wires SSH convenience helpers for homelab hosts.
func NewSSHCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ssh",
		Short: "SSH into lab machines with managed host definitions",
		Long:  "Wraps ssh with curated inventory from config (ssh.hosts) and homelab sync helpers.",
		RunE: func(c *cobra.Command, _ []string) error {
			return c.Help()
		},
	}
	AddDryRunFlag(cmd)

	cmd.AddCommand(
		&cobra.Command{
			Use:     "connect <host>",
			Short:   "Open an interactive SSH session to a known host alias",
			Example: "  lab ssh connect gpu-01",
			Args:    cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				s := session(cmd)
				return ssh.Connect(cmd.Context(), s.Config, args[0])
			},
		},
		&cobra.Command{
			Use:     "sync",
			Short:   "Rsync homelab repo to configured server",
			Example: "  lab ssh sync",
			Args:    cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				setDryRun(cmd)
				s := session(cmd)
				ui.Section(stdout(cmd), s.Styles, "ssh sync", "homelab → server")
				return ssh.SyncRepo(cmd.Context(), s.Config, s.HomelabRoot, s.DryRun, stdout(cmd), stderr(cmd))
			},
		},
	)

	return cmd
}
