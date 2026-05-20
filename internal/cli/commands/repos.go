package commands

import (
	"github.com/bartrosa/homelab-cli/internal/repos"
	"github.com/bartrosa/homelab-cli/internal/ui"
	"github.com/spf13/cobra"
)

// NewReposCmd wires multi-repo workflows across Git providers.
func NewReposCmd() *cobra.Command {
	var jobs int

	cmd := &cobra.Command{
		Use:   "repos",
		Short: "Clone, mirror, and synchronize repositories",
		Long: `repos integrates with Git hosting providers for bulk workflows.
GitLab backup uses tools/gitlab/backup_account.py from your homelab repo.`,
		RunE: func(c *cobra.Command, _ []string) error {
			return c.Help()
		},
	}
	AddDryRunFlag(cmd)

	cmd.AddCommand(
		&cobra.Command{
			Use:     `clone <pattern>`,
			Short:   "Clone repositories matching a host/org pattern",
			Long:    `Patterns look like "github.com/owner/*" or "gitlab.com/group/**".`,
			Example: `  lab repos clone "github.com/me/*"`,
			Args:    cobra.ExactArgs(1),
			RunE:    StubRunE(),
		},
		&cobra.Command{
			Use:     "backup",
			Short:   "Mirror GitLab projects to local backup dir",
			Example: "  lab repos backup\n  lab repos backup --jobs 4",
			Args:    cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				setDryRun(cmd)
				s := session(cmd)
				ui.Section(stdout(cmd), s.Styles, "repos backup", "GitLab mirror")
				return repos.GitLabBackup(cmd.Context(), s.Config, s.HomelabRoot, s.DryRun, stdout(cmd), stderr(cmd), jobs)
			},
		},
		&cobra.Command{
			Use:     "sync",
			Short:   "Fetch/pull across all local clones under the configured root",
			Example: "  lab repos sync",
			Args:    cobra.NoArgs,
			RunE:    StubRunE(),
		},
		&cobra.Command{
			Use:     "status",
			Short:   "Show dirty branches and ahead/behind summaries",
			Example: "  lab repos status",
			Args:    cobra.NoArgs,
			RunE:    StubRunE(),
		},
		&cobra.Command{
			Use:     "list",
			Short:   "List remote repositories available to configured providers",
			Example: "  lab repos list",
			Args:    cobra.NoArgs,
			RunE:    StubRunE(),
		},
	)

	cmd.PersistentFlags().IntVar(&jobs, "jobs", 2, "parallel git jobs for GitLab backup")

	return cmd
}
