package commands

import (
	"fmt"

	"github.com/bartrosa/homelab-cli/internal/baremetal"
	"github.com/bartrosa/homelab-cli/internal/ui"
	"github.com/spf13/cobra"
)

// NewBaremetalCmd installs vector/OLAP databases on Linux hosts.
func NewBaremetalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "baremetal",
		Short: "Install databases on bare metal (no containers)",
		Long:  "Run on the target Linux server. Uses curl/apt/sudo — see docs/external-binaries.md.",
		RunE: func(c *cobra.Command, _ []string) error {
			return c.Help()
		},
	}
	AddDryRunFlag(cmd)

	install := &cobra.Command{
		Use:   "install <qdrant|milvus|clickhouse>",
		Short: "Install qdrant, milvus, or clickhouse",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			setDryRun(cmd)
			s := session(cmd)
			ui.Section(stdout(cmd), s.Styles, "baremetal install", args[0])
			ctx := cmd.Context()
			switch args[0] {
			case "qdrant":
				return baremetal.InstallQdrant(ctx, s.DryRun, stdout(cmd), stderr(cmd))
			case "milvus":
				return baremetal.InstallMilvus(ctx, s.DryRun, stdout(cmd), stderr(cmd))
			case "clickhouse":
				return baremetal.InstallClickHouse(ctx, s.DryRun, stdout(cmd), stderr(cmd))
			default:
				return fmt.Errorf("unknown target %q (use: qdrant, milvus, clickhouse)", args[0])
			}
		},
	}
	install.ValidArgs = []string{"qdrant", "milvus", "clickhouse"}

	cmd.AddCommand(install)
	return cmd
}
