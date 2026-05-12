package commands

import "github.com/spf13/cobra"

// NewNetCmd wires Tailscale, WireGuard, and DNS helpers.
func NewNetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "net",
		Short: "Homelab networking (Tailscale, WireGuard, DNS, mDNS)",
		Long:  "Manage overlays, DNS records, and quick diagnostics for lab networks.",
		RunE: func(c *cobra.Command, _ []string) error {
			return c.Help()
		},
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:     "status",
			Short:   "Show VPN interfaces, peers, and DNS resolver health",
			Example: "  lab net status",
			Args:    cobra.NoArgs,
			RunE:    StubRunE(),
		},
	)

	return cmd
}
