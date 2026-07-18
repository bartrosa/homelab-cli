package commands

import (
	"fmt"
	"strings"

	"github.com/bartrosa/homelab-cli/internal/exec"
	"github.com/bartrosa/homelab-cli/internal/platform"
	"github.com/bartrosa/homelab-cli/internal/tofu"
	"github.com/bartrosa/homelab-cli/internal/ui"
	"github.com/spf13/cobra"
)

// NewTofuCmd wires OpenTofu install and stack orchestration.
func NewTofuCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tofu",
		Short: "Install OpenTofu and orchestrate IaC stacks",
		Long: `Install the OpenTofu binary via OS-native methods (brew, official installer script, snap, winget),
then run tofu init/plan/apply/validate/fmt against a stack directory.

Secrets for apply/plan (TF_VAR_*, passwords) should be injected by wrapping lab with
op run (1Password op:// refs) — never written to files.`,
		RunE: func(c *cobra.Command, _ []string) error {
			return c.Help()
		},
	}
	AddDryRunFlag(cmd)

	var upgrade bool
	install := &cobra.Command{
		Use:   "install",
		Short: "Install OpenTofu binary (detects OS)",
		Long:  "Delegates to brew, the official get.opentofu.org script, snap, or winget — does not manage repo keys itself.",
		Example: `  lab tofu install
  lab tofu install --upgrade
  lab tofu install --dry-run`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			setDryRun(cmd)
			s := session(cmd)
			info := platform.Detect()
			method, err := tofu.DetectMethod(info)
			if err != nil {
				return err
			}
			action := "install"
			if upgrade {
				action = "upgrade"
			}
			detail := fmt.Sprintf("%s via %s — %s", action, method.Kind, method.Reason)
			title := "tofu install"
			if s.DryRun {
				title = "tofu install (dry-run)"
			}
			ui.Section(stdout(cmd), s.Styles, title, detail)

			opts := tofu.InstallOpts{
				Info:   info,
				DryRun: s.DryRun,
				Stdout: stdout(cmd),
				Stderr: stderr(cmd),
			}
			runner := exec.NewOSRunner(stdout(cmd), stderr(cmd))
			if upgrade {
				return tofu.Upgrade(cmd.Context(), runner, opts)
			}
			return tofu.Install(cmd.Context(), runner, opts)
		},
	}
	install.Flags().BoolVar(&upgrade, "upgrade", false, "upgrade existing OpenTofu install")

	version := &cobra.Command{
		Use:   "version",
		Short: "Print installed tofu version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			setDryRun(cmd)
			s := session(cmd)
			runner := exec.NewOSRunner(stdout(cmd), stderr(cmd))
			ok, ver, err := tofu.IsInstalled(cmd.Context(), runner)
			if err != nil {
				return err
			}
			if !ok {
				ui.Warn(stdout(cmd), s.Styles, "tofu not found on PATH — run: lab tofu install")
				return fmt.Errorf("tofu not installed")
			}
			ui.Section(stdout(cmd), s.Styles, "tofu version", ver)
			return nil
		},
	}

	var yes bool
	apply := &cobra.Command{
		Use:   "apply <stack>",
		Short: "Apply stack changes (requires --yes for -auto-approve)",
		Long: `Runs tofu apply in the stack directory.
Without --yes, tofu prompts interactively. Pass --yes to add -auto-approve.

TODO: resolve stack names relative to homelab.root/hq-infra from config when a bare name is given.`,
		Example: `  lab tofu apply ./hq-infra/tofu/stacks/rb5009-core --yes
  lab tofu apply ./stacks/rb5009-core --dry-run`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			setDryRun(cmd)
			s := session(cmd)
			stack := resolveStackArg(args[0])
			extra := ""
			if yes {
				extra = " (-auto-approve)"
			}
			title := "tofu apply"
			if s.DryRun {
				title = "tofu apply (dry-run)"
			}
			ui.Section(stdout(cmd), s.Styles, title, stack.Dir+extra)
			opts := tofu.RunOpts{AutoApprove: yes, DryRun: s.DryRun}
			runner := exec.NewOSRunner(stdout(cmd), stderr(cmd))
			return tofu.Apply(cmd.Context(), runner, stack, opts)
		},
	}
	apply.Flags().BoolVar(&yes, "yes", false, "pass -auto-approve to tofu apply")

	cmd.AddCommand(
		install,
		version,
		newTofuStackCmd("init", "Initialize a stack working directory"),
		newTofuStackCmd("validate", "Validate stack configuration"),
		newTofuStackCmd("plan", "Show execution plan for a stack"),
		apply,
		newTofuStackCmd("fmt", "Format stack configuration"),
	)

	return cmd
}

func newTofuStackCmd(name, short string) *cobra.Command {
	return &cobra.Command{
		Use:   name + " <stack>",
		Short: short,
		Long: `Runs tofu ` + name + ` with -chdir=<stack>.

TODO: resolve stack names relative to homelab.root/hq-infra from config when a bare name is given.`,
		Example: fmt.Sprintf("  lab tofu %s ./hq-infra/tofu/stacks/rb5009-core", name),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			setDryRun(cmd)
			s := session(cmd)
			stack := resolveStackArg(args[0])
			title := "tofu " + name
			if s.DryRun {
				title += " (dry-run)"
				ui.Section(stdout(cmd), s.Styles, title, "tofu "+strings.Join(tofu.CommandLine(stack, name), " "))
			} else {
				ui.Section(stdout(cmd), s.Styles, title, stack.Dir)
			}
			opts := tofu.RunOpts{DryRun: s.DryRun}
			runner := exec.NewOSRunner(stdout(cmd), stderr(cmd))
			switch name {
			case "init":
				return tofu.Init(cmd.Context(), runner, stack, opts)
			case "validate":
				return tofu.Validate(cmd.Context(), runner, stack, opts)
			case "plan":
				return tofu.Plan(cmd.Context(), runner, stack, opts)
			case "fmt":
				return tofu.Fmt(cmd.Context(), runner, stack, opts)
			default:
				return fmt.Errorf("unknown tofu subcommand %q", name)
			}
		},
	}
}

// resolveStackArg treats the argument as a filesystem path to the stack directory.
// TODO: when arg is a bare name (no path separator), resolve under homelab.root/hq-infra
// (e.g. stacks/<name> or tofu/stacks/<name>) from config.
func resolveStackArg(arg string) tofu.Stack {
	return tofu.Stack{Dir: strings.TrimSpace(arg)}
}
