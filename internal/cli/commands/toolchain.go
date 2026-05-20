package commands

import (
	"strings"

	"github.com/bartrosa/homelab-cli/internal/toolchain"
	"github.com/bartrosa/homelab-cli/internal/ui"
	"github.com/spf13/cobra"
)

// NewToolchainCmd wires language toolchain operations (mise-backed).
func NewToolchainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "toolchain",
		Short: "Install and switch language toolchains",
		Long: `Toolchain commands wrap mise to install and activate
Go, Node, Bun, Deno, Python, Rust, Erlang, Elixir, Zig, Java, Ruby and more.`,
		RunE: func(c *cobra.Command, _ []string) error {
			return c.Help()
		},
	}
	AddDryRunFlag(cmd)

	cmd.AddCommand(
		&cobra.Command{
			Use:     "install <lang> [more...]",
			Short:   "Install one or more language toolchains",
			Example: "  lab toolchain install go bun rust",
			Args:    cobra.MinimumNArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				setDryRun(cmd)
				s := session(cmd)
				tc := toolchain.New(stdout(cmd), stderr(cmd), s.DryRun)
				ui.Section(stdout(cmd), s.Styles, "toolchain", "mise")
				if err := tc.Install(cmd.Context(), args...); err != nil {
					return err
				}
				ui.OK(stdout(cmd), s.Styles, strings.Join(args, ", "))
				return nil
			},
		},
		&cobra.Command{
			Use:     "list",
			Short:   "List installed toolchains and active versions",
			Example: "  lab toolchain list",
			Args:    cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				setDryRun(cmd)
				s := session(cmd)
				tc := toolchain.New(stdout(cmd), stderr(cmd), s.DryRun)
				return tc.List(cmd.Context())
			},
		},
		&cobra.Command{
			Use:     "use <lang> <version>",
			Short:   "Switch the active toolchain version for a language",
			Example: "  lab toolchain use go 1.25.0",
			Args:    cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				setDryRun(cmd)
				s := session(cmd)
				tc := toolchain.New(stdout(cmd), stderr(cmd), s.DryRun)
				return tc.Use(cmd.Context(), args[0], args[1])
			},
		},
	)

	return cmd
}
