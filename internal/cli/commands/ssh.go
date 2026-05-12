package commands

import "github.com/spf13/cobra"

// NewSSHCmd wires SSH convenience helpers for homelab hosts.
func NewSSHCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ssh",
		Short: "SSH into lab machines with managed host definitions",
		Long:  "Wraps ssh/scp with curated inventory, jump hosts, and key management.",
		RunE: func(c *cobra.Command, _ []string) error {
			return c.Help()
		},
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:     "connect <host>",
			Short:   "Open an interactive SSH session to a known host alias",
			Example: "  lab ssh connect gpu-01",
			Args:    cobra.ExactArgs(1),
			RunE:    StubRunE(),
		},
	)

	return cmd
}
