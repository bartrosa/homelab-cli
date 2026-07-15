package commands

import (
	"fmt"

	"github.com/bartrosa/homelab-cli/internal/services"
	"github.com/bartrosa/homelab-cli/internal/ui"
	"github.com/spf13/cobra"
)

// NewDataCmd wires dataset / storage service helpers.
func NewDataCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "data",
		Short: "Data services (Postgres, ClickHouse, MinIO)",
		Long:  "Start data platform services via lab services.",
		RunE: func(c *cobra.Command, _ []string) error {
			return c.Help()
		},
	}
	AddDryRunFlag(cmd)

	cmd.AddCommand(
		&cobra.Command{
			Use:     "up <postgres|clickhouse|minio>",
			Short:   "Start a data service",
			Example: "  lab data up postgres",
			Args:    cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				setDryRun(cmd)
				s := session(cmd)
				ui.Section(stdout(cmd), s.Styles, "data up", args[0])
				o := &services.Orchestrator{CustomPresets: s.Config.Services.Presets}
				opts := svcOpts(cmd)
				opts.NonInteractive = true
				return o.Up(cmd.Context(), opts, args[0])
			},
		},
		&cobra.Command{
			Use:     "sync",
			Short:   "List data services status",
			Example: "  lab data sync",
			Args:    cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				for _, id := range []string{"postgres", "clickhouse", "minio"} {
					svc, ok := services.Lookup(id)
					if !ok {
						continue
					}
					st, _ := svc.Status(cmd.Context(), svcOpts(cmd))
					fmt.Fprintf(stdout(cmd), "  %s: running=%v\n", id, st.Running)
				}
				return nil
			},
		},
	)

	return cmd
}
