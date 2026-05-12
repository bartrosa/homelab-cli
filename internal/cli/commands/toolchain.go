package commands

import "github.com/spf13/cobra"

// NewToolchainCmd wires language toolchain operations (mise-backed).
func NewToolchainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "toolchain",
		Short: "Install and switch language toolchains",
		Long: `Toolchain commands wrap mise (or compatible shims) to install and activate
Go, Node, Bun, Deno, Python, Rust, Erlang, Elixir, Zig, Java, Ruby and more.`,
		RunE: func(c *cobra.Command, _ []string) error {
			return c.Help()
		},
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:     "install <lang> [more...]",
			Short:   "Install one or more language toolchains",
			Example: "  lab toolchain install go bun rust",
			Args:    cobra.MinimumNArgs(1),
			RunE:    StubRunE(),
		},
		&cobra.Command{
			Use:     "list",
			Short:   "List installed toolchains and active versions",
			Example: "  lab toolchain list",
			Args:    cobra.NoArgs,
			RunE:    StubRunE(),
		},
		&cobra.Command{
			Use:     "use <lang> <version>",
			Short:   "Switch the active toolchain version for a language",
			Example: "  lab toolchain use go 1.25.0",
			Args:    cobra.ExactArgs(2),
			RunE:    StubRunE(),
		},
	)

	return cmd
}
