package commands

import "github.com/spf13/cobra"

// NewBootstrapCmd wires bootstrap subcommands (laptop, server, profile).
func NewBootstrapCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bootstrap",
		Short: "Bootstrap machines from zero (laptop or server profiles)",
		Long: `Bootstrap prepares a fresh machine with baseline packages, security posture,
and optional dotfiles. Profiles are defined in the configuration file.`,
		RunE: func(c *cobra.Command, _ []string) error {
			return c.Help()
		},
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:     "laptop",
			Short:   "Initialize a new developer laptop",
			Long:    "Installs base packages, shells, fonts, dotfiles, git, and container tooling.",
			Example: "  lab bootstrap laptop",
			RunE:    StubRunE(),
		},
		&cobra.Command{
			Use:     "server",
			Short:   "Initialize a new homelab server",
			Long:    "SSH hardening, container runtime, monitoring agents, and fail2ban-style baselines.",
			Example: "  lab bootstrap server",
			RunE:    StubRunE(),
		},
		&cobra.Command{
			Use:     "profile <name>",
			Short:   "Bootstrap using a named profile from config",
			Long:    "Runs the bootstrap graph defined under bootstrap.profiles.<name> in the config file.",
			Example: "  lab bootstrap profile dgx-spark",
			Args:    cobra.ExactArgs(1),
			RunE:    StubRunE(),
		},
	)

	return cmd
}
