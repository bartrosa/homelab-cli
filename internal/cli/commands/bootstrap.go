package commands

import (
	"github.com/bartrosa/homelab-cli/internal/bootstrap"
	"github.com/bartrosa/homelab-cli/internal/platform"
	"github.com/bartrosa/homelab-cli/internal/ui"
	"github.com/spf13/cobra"
)

// NewBootstrapCmd wires bootstrap subcommands (laptop, server, profile).
func NewBootstrapCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bootstrap",
		Short: "Bootstrap machines from zero (laptop or server profiles)",
		Long: `Bootstrap prepares a fresh machine with baseline packages, security posture,
and optional dotfiles. Profiles are defined in the configuration file or built-in YAML.`,
		RunE: func(c *cobra.Command, _ []string) error {
			return c.Help()
		},
	}
	AddDryRunFlag(cmd)

	cmd.AddCommand(
		newBootstrapLaptopCmd(),
		newBootstrapServerCmd(),
		newBootstrapProfileCmd(),
		newBootstrapListCmd(),
		newBootstrapEssentialsCmd(),
	)

	return cmd
}

func newBootstrapLaptopCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "laptop",
		Short:   "Initialize a new developer laptop",
		Long:    "Installs base packages and toolchains using a built-in profile for your OS.",
		Example: "  lab bootstrap laptop\n  lab bootstrap laptop --dry-run",
		RunE: func(cmd *cobra.Command, _ []string) error {
			setDryRun(cmd)
			s := session(cmd)
			name := laptopProfileName()
			profile, err := bootstrap.LoadEmbedded(name)
			if err != nil {
				return err
			}
			runner := &bootstrap.Runner{
				Stdout: stdout(cmd), Stderr: stderr(cmd),
				HomelabRoot: s.HomelabRoot, DryRun: s.DryRun, Styles: s.Styles,
			}
			return runner.RunProfile(cmd.Context(), profile)
		},
	}
	return cmd
}

func newBootstrapServerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "server",
		Short:   "Initialize a new homelab server",
		Long:    "Ubuntu/Debian server baseline: packages and install-server-deps from homelab repo.",
		Example: "  lab bootstrap server",
		RunE: func(cmd *cobra.Command, _ []string) error {
			setDryRun(cmd)
			s := session(cmd)
			profile, err := bootstrap.LoadEmbedded("server-ubuntu")
			if err != nil {
				return err
			}
			runner := &bootstrap.Runner{
				Stdout: stdout(cmd), Stderr: stderr(cmd),
				HomelabRoot: s.HomelabRoot, DryRun: s.DryRun, Styles: s.Styles,
			}
			return runner.RunProfile(cmd.Context(), profile)
		},
	}
	return cmd
}

func newBootstrapProfileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "profile <name>",
		Short:   "Bootstrap using a named profile",
		Long:    "Runs a built-in profile or bootstrap.profiles.<name> from config.",
		Example: "  lab bootstrap profile laptop-macos\n  lab bootstrap profile dgx-spark",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			setDryRun(cmd)
			s := session(cmd)
			name := args[0]

			var profile *bootstrap.Profile
			var err error
			if raw, ok := s.Config.Bootstrap.Profiles[name]; ok {
				profile, err = bootstrap.DecodeProfileMap(raw)
			} else {
				profile, err = bootstrap.LoadEmbedded(name)
			}
			if err != nil {
				return err
			}
			runner := &bootstrap.Runner{
				Stdout: stdout(cmd), Stderr: stderr(cmd),
				HomelabRoot: s.HomelabRoot, DryRun: s.DryRun, Styles: s.Styles,
			}
			return runner.RunProfile(cmd.Context(), profile)
		},
	}
	return cmd
}

func newBootstrapListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List built-in bootstrap profiles",
		RunE: func(cmd *cobra.Command, _ []string) error {
			s := session(cmd)
			names, err := bootstrap.ListEmbedded()
			if err != nil {
				return err
			}
			ui.Section(stdout(cmd), s.Styles, "Built-in profiles", "")
			var summaries []bootstrap.ProfileSummary
			for _, n := range names {
				p, err := bootstrap.LoadEmbedded(n)
				if err != nil {
					continue
				}
				summaries = append(summaries, bootstrap.ProfileSummary{
					Name: p.Name, Description: p.Description, Source: "builtin",
				})
			}
			bootstrap.WriteProfileList(stdout(cmd), summaries)
			return nil
		},
	}
}

func laptopProfileName() string {
	info := platform.Detect()
	switch {
	case info.IsSilverblue:
		return "silverblue-laptop"
	case info.GOOS == platform.OSDarwin:
		return "laptop-macos"
	case info.GOOS == platform.OSLinux:
		return "laptop-linux"
	default:
		return "laptop-macos"
	}
}
