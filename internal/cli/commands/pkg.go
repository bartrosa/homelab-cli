package commands

import (
	"fmt"

	"github.com/bartrosa/homelab-cli/internal/packager"
	"github.com/bartrosa/homelab-cli/internal/ui"
	"github.com/spf13/cobra"
)

// NewPkgCmd wires package manager abstractions (brew, apt, dnf).
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
	AddDryRunFlag(cmd)

	cmd.AddCommand(
		&cobra.Command{
			Use:     "install <name>",
			Short:   "Install a package using the native package manager",
			Example: "  lab pkg install ripgrep",
			Args:    cobra.ExactArgs(1),
			RunE:    pkgInstallRunE(false),
		},
		&cobra.Command{
			Use:     "ensure <name>",
			Short:   "Idempotently ensure a package is present",
			Example: "  lab pkg ensure jq",
			Args:    cobra.ExactArgs(1),
			RunE:    pkgInstallRunE(true),
		},
		&cobra.Command{
			Use:     "list",
			Short:   "List common lab packages and detected backend",
			Example: "  lab pkg list",
			Args:    cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				setDryRun(cmd)
				s := session(cmd)
				mgr := packager.New(stdout(cmd), stderr(cmd), s.DryRun)
				ui.Section(stdout(cmd), s.Styles, "Package manager", mgr.Info.PackagerLabel())
				var rows [][]string
				for _, p := range mgr.ListTracked() {
					rows = append(rows, []string{p, "lab pkg ensure " + p})
				}
				ui.Table(stdout(cmd), s.Styles, []string{"PACKAGE", "COMMAND"}, rows)
				return nil
			},
		},
	)

	return cmd
}

func pkgInstallRunE(ensure bool) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		setDryRun(cmd)
		s := session(cmd)
		mgr := packager.New(stdout(cmd), stderr(cmd), s.DryRun)
		name := args[0]
		ui.Section(stdout(cmd), s.Styles, "pkg", mgr.Info.PackagerLabel())
		var err error
		if ensure {
			err = mgr.Ensure(cmd.Context(), name)
		} else {
			err = mgr.Install(cmd.Context(), name)
		}
		if err != nil {
			return err
		}
		ui.OK(stdout(cmd), s.Styles, fmt.Sprintf("%s ok", name))
		return nil
	}
}
