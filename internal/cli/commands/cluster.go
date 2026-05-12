package commands

import "github.com/spf13/cobra"

// NewClusterCmd wires Kubernetes/k3s helpers.
func NewClusterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Manage k3s/Kubernetes clusters in the homelab",
		Long:  "Inspect nodes, kubeconfig contexts, and deploy curated addons.",
		RunE: func(c *cobra.Command, _ []string) error {
			return c.Help()
		},
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:     "status",
			Short:   "Show cluster connectivity and node readiness",
			Example: "  lab cluster status",
			Args:    cobra.NoArgs,
			RunE:    StubRunE(),
		},
		&cobra.Command{
			Use:     "kubeconfig",
			Short:   "Print or merge kubeconfig snippets for homelab contexts",
			Example: "  lab cluster kubeconfig",
			Args:    cobra.NoArgs,
			RunE:    StubRunE(),
		},
	)

	return cmd
}
