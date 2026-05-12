package commands

import "github.com/spf13/cobra"

// NewReposCmd wires multi-repo workflows across Git providers.
func NewReposCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repos",
		Short: "Clone, mirror, and synchronize repositories",
		Long: `repos integrates with GitHub, GitLab, Gitea, and Codeberg using REST APIs plus
local git operations for bulk workflows across organizations and patterns.`,
		RunE: func(c *cobra.Command, _ []string) error {
			return c.Help()
		},
	}

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
			Short:   "Mirror configured repositories to a local path or object storage",
			Example: "  lab repos backup",
			Args:    cobra.NoArgs,
			RunE:    StubRunE(),
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

	return cmd
}
