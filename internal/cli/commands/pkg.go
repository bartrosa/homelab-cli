package commands

import "github.com/spf13/cobra"

// NewPkgCmd wires package manager abstractions (brew, apt, dnf, pacman).
func NewPkgCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pkg",
		Short: "Cross-platform package installation helpers",
		Long: `pkg abstracts native package managers so playbooks can stay declarative.
Detection picks the right backend for the host OS.`,
		RunE: func(c *cobra.Command, _ []string) error {
			return c.Help()
		},
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:     "install <name>",
			Short:   "Install a package using the native package manager",
			Example: "  lab pkg install ripgrep",
			Args:    cobra.ExactArgs(1),
			RunE:    StubRunE(),
		},
		&cobra.Command{
			Use:     "ensure <name>",
			Short:   "Idempotently ensure a package is present",
			Example: "  lab pkg ensure jq",
			Args:    cobra.ExactArgs(1),
			RunE:    StubRunE(),
		},
		&cobra.Command{
			Use:     "list",
			Short:   "List packages tracked by lab",
			Example: "  lab pkg list",
			Args:    cobra.NoArgs,
			RunE:    StubRunE(),
		},
	)

	return cmd
}
