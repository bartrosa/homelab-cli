package commands

import "github.com/spf13/cobra"

// NewServicesCmd wires local compose stacks for databases and brokers.
func NewServicesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "services",
		Short: "Run homelab data services via compose (docker/podman)",
		Long: `services manages opinionated compose stacks for Postgres, Redis, MongoDB, Kafka,
RabbitMQ, MinIO, ClickHouse, etcd, NATS, and similar dependencies.`,
		RunE: func(c *cobra.Command, _ []string) error {
			return c.Help()
		},
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:     "up <name> [more...]",
			Short:   "Start one or more service stacks",
			Example: "  lab services up postgres redis",
			Args:    cobra.MinimumNArgs(1),
			RunE:    StubRunE(),
		},
		&cobra.Command{
			Use:     "down <name> [more...]",
			Short:   "Stop one or more service stacks",
			Example: "  lab services down postgres",
			Args:    cobra.MinimumNArgs(1),
			RunE:    StubRunE(),
		},
		&cobra.Command{
			Use:     "list",
			Short:   "List available stacks and their runtime status",
			Example: "  lab services list",
			Args:    cobra.NoArgs,
			RunE:    StubRunE(),
		},
		&cobra.Command{
			Use:     "logs <name>",
			Short:   "Tail logs for a running stack",
			Example: "  lab services logs postgres",
			Args:    cobra.ExactArgs(1),
			RunE:    StubRunE(),
		},
	)

	return cmd
}
