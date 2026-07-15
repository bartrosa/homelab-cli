package commands

import (
	"github.com/bartrosa/homelab-cli/internal/bootstrap"
	"github.com/bartrosa/homelab-cli/internal/exec"
	"github.com/spf13/cobra"
)

func newBootstrapEssentialsCmd() *cobra.Command {
	var (
		target string
		yes    bool
		skip   string
		only   string
	)

	cmd := &cobra.Command{
		Use:     "essentials",
		Short:   "Install baseline packages for Ubuntu or Fedora Silverblue",
		Long:    "Idempotent bootstrap of CLI tools, build deps, containers, mise, and Silverblue-specific layers.",
		Example: "  lab bootstrap essentials\n  lab bootstrap essentials --dry-run --target silverblue\n  lab bootstrap essentials --only cli-basics --yes",
		RunE: func(cmd *cobra.Command, _ []string) error {
			setDryRun(cmd)
			s := session(cmd)
			return bootstrap.RunEssentials(cmd.Context(), bootstrap.EssentialsOptions{
				Target: target,
				Yes:    yes,
				Skip:   bootstrap.ParseCSV(skip),
				Only:   bootstrap.ParseCSV(only),
				DryRun: s.DryRun,
				Stdout: stdout(cmd),
				Stderr: stderr(cmd),
				Runner: exec.NewOSRunner(stdout(cmd), stderr(cmd)),
			})
		},
	}

	cmd.Flags().StringVar(&target, "target", "auto", "ubuntu|silverblue|auto")
	cmd.Flags().BoolVar(&yes, "yes", false, "accept defaults without prompts")
	cmd.Flags().StringVar(&skip, "skip", "", "comma-separated sections to skip")
	cmd.Flags().StringVar(&only, "only", "", "comma-separated sections to run")
	AddDryRunFlag(cmd)

	return cmd
}
