package commands

import (
	"github.com/bartrosa/homelab-cli/internal/postgres"
	"github.com/bartrosa/homelab-cli/internal/ui"
	"github.com/spf13/cobra"
)

// NewPostgresCmd wires PostgreSQL provisioning from YAML.
func NewPostgresCmd() *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:   "postgres",
		Short: "PostgreSQL provisioning (YAML desired state)",
		Long:  "Idempotent apply of databases/users from instances.yaml. Requires POSTGRES_ADMIN_PASSWORD or PGPASSWORD.",
		RunE: func(c *cobra.Command, _ []string) error {
			return c.Help()
		},
	}
	AddDryRunFlag(cmd)

	apply := &cobra.Command{
		Use:     "apply",
		Short:   "Apply instances.yaml to PostgreSQL",
		Example: `  lab postgres apply --config ~/homelab/postgres/config/instances.yaml`,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			setDryRun(cmd)
			s := session(cmd)
			ui.Section(stdout(cmd), s.Styles, "postgres apply", configPath)
			cfg, err := postgres.LoadConfig(configPath)
			if err != nil {
				return err
			}
			return postgres.Apply(cmd.Context(), cfg, s.DryRun)
		},
	}
	apply.Flags().StringVar(&configPath, "config", "", "path to instances.yaml (required)")
	_ = apply.MarkFlagRequired("config")

	cmd.AddCommand(apply)
	return cmd
}
