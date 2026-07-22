package commands

import (
	"fmt"

	"github.com/bartrosa/homelab-cli/internal/exec"
	"github.com/bartrosa/homelab-cli/internal/platform"
	"github.com/bartrosa/homelab-cli/internal/podman"
	"github.com/bartrosa/homelab-cli/internal/ui"
	"github.com/spf13/cobra"
)

// NewPodmanCmd wires Podman install, configure, doctor, upgrade and remove.
func NewPodmanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "podman",
		Short: "Install, configure, and verify the Podman container runtime",
		Long: `Install Podman via the OS package manager, then configure rootless prerequisites
(subuid/subgid, linger, Quadlet dirs, registries). Prefer Quadlet + systemd --user
over podman-compose so services restart after power loss.`,
		RunE: func(c *cobra.Command, _ []string) error {
			return c.Help()
		},
	}
	AddDryRunFlag(cmd)

	var (
		skipConfigure bool
		allowUserNS   bool
		enableSocket  bool
	)

	install := &cobra.Command{
		Use:   "install",
		Short: "Install Podman and run post-install configuration",
		Example: `  lab podman install
  lab podman install --skip-configure
  lab podman install --allow-userns --enable-socket
  lab podman install --dry-run`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			setDryRun(cmd)
			s := session(cmd)
			info := platform.Detect()
			method, err := podman.DetectMethod(info)
			if err != nil {
				return err
			}
			title := "podman install"
			if s.DryRun {
				title += " (dry-run)"
			}
			ui.Section(stdout(cmd), s.Styles, title, method.Kind+" — "+method.Reason)

			runner := exec.NewOSRunner(stdout(cmd), stderr(cmd))
			if err := podman.Install(cmd.Context(), runner, podman.InstallOpts{
				Info: info, DryRun: s.DryRun, Stdout: stdout(cmd), Stderr: stderr(cmd),
			}); err != nil {
				return err
			}
			if skipConfigure {
				ui.OK(stdout(cmd), s.Styles, "skipped configure (--skip-configure)")
				return nil
			}
			return runPodmanConfigure(cmd, info, allowUserNS, enableSocket)
		},
	}
	install.Flags().BoolVar(&skipConfigure, "skip-configure", false, "install packages only; skip post-install configure")
	install.Flags().BoolVar(&allowUserNS, "allow-userns", false, "on Ubuntu >= 23.10, allow disabling AppArmor unprivileged userns restriction (security trade-off)")
	install.Flags().BoolVar(&enableSocket, "enable-socket", false, "enable systemd --user podman.socket (Docker API)")

	configure := &cobra.Command{
		Use:   "configure",
		Short: "Run post-install configuration (rootless, Quadlet, registries)",
		Example: `  lab podman configure
  lab podman configure --allow-userns --enable-socket`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			setDryRun(cmd)
			s := session(cmd)
			info := platform.Detect()
			title := "podman configure"
			if s.DryRun {
				title += " (dry-run)"
			}
			ui.Section(stdout(cmd), s.Styles, title, info.PackagerLabel())
			return runPodmanConfigure(cmd, info, allowUserNS, enableSocket)
		},
	}
	configure.Flags().BoolVar(&allowUserNS, "allow-userns", false, "on Ubuntu >= 23.10, allow disabling AppArmor unprivileged userns restriction (security trade-off)")
	configure.Flags().BoolVar(&enableSocket, "enable-socket", false, "enable systemd --user podman.socket (Docker API)")

	doctor := &cobra.Command{
		Use:   "doctor",
		Short: "Verify rootless Podman and Quadlet prerequisites",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			setDryRun(cmd)
			s := session(cmd)
			ui.Section(stdout(cmd), s.Styles, "podman doctor", "")
			runner := exec.NewOSRunner(stdout(cmd), stderr(cmd))
			report, err := podman.Doctor(cmd.Context(), runner, podman.DoctorOpts{Info: platform.Detect()})
			if err != nil {
				return err
			}
			rows := make([][]string, 0, len(report.Checks))
			fails := 0
			for _, c := range report.Checks {
				rows = append(rows, []string{c.Status, c.Name, c.Detail})
				if c.Status == podman.StatusFail {
					fails++
					if c.Fix != "" {
						ui.Warn(stdout(cmd), s.Styles, c.Name+": "+c.Fix)
					}
				}
			}
			ui.Table(stdout(cmd), s.Styles, []string{"STATUS", "CHECK", "DETAIL"}, rows)
			if fails > 0 {
				return fmt.Errorf("podman doctor: %d check(s) failed", fails)
			}
			ui.OK(stdout(cmd), s.Styles, "all checks passed")
			return nil
		},
	}

	upgrade := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade Podman via the OS package manager",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			setDryRun(cmd)
			s := session(cmd)
			info := platform.Detect()
			method, err := podman.DetectMethod(info)
			if err != nil {
				return err
			}
			title := "podman upgrade"
			if s.DryRun {
				title += " (dry-run)"
			}
			ui.Section(stdout(cmd), s.Styles, title, method.Kind+" — "+method.Reason)
			runner := exec.NewOSRunner(stdout(cmd), stderr(cmd))
			return podman.Upgrade(cmd.Context(), runner, podman.InstallOpts{
				Info: info, DryRun: s.DryRun, Stdout: stdout(cmd), Stderr: stderr(cmd),
			})
		},
	}

	var (
		purge bool
		yes   bool
	)
	remove := &cobra.Command{
		Use:   "remove",
		Short: "Remove Podman packages",
		Long:  "Removes Podman packages. With --purge, also deletes ~/.local/share/containers and ~/.config/containers (all images and volumes).",
		Example: `  lab podman remove
  lab podman remove --purge --yes`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			setDryRun(cmd)
			s := session(cmd)
			if purge && !yes && !s.DryRun {
				return fmt.Errorf("--purge deletes all local images and volumes; re-run with --yes to confirm")
			}
			info := platform.Detect()
			method, err := podman.DetectMethod(info)
			if err != nil {
				return err
			}
			title := "podman remove"
			if s.DryRun {
				title += " (dry-run)"
			}
			ui.Section(stdout(cmd), s.Styles, title, method.Kind)
			runner := exec.NewOSRunner(stdout(cmd), stderr(cmd))
			if err := podman.Remove(cmd.Context(), runner, podman.InstallOpts{
				Info: info, DryRun: s.DryRun, Purge: purge, Stdout: stdout(cmd), Stderr: stderr(cmd),
			}); err != nil {
				return err
			}
			if purge {
				ui.Warn(stdout(cmd), s.Styles, "purging user container storage and config")
				return podman.PurgeUserData(podman.ConfigureOpts{DryRun: s.DryRun, Stdout: stdout(cmd)})
			}
			return nil
		},
	}
	remove.Flags().BoolVar(&purge, "purge", false, "also remove ~/.local/share/containers and ~/.config/containers (destructive)")
	remove.Flags().BoolVar(&yes, "yes", false, "confirm --purge (required when not --dry-run)")

	version := &cobra.Command{
		Use:   "version",
		Short: "Print installed Podman version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			setDryRun(cmd)
			s := session(cmd)
			runner := exec.NewOSRunner(stdout(cmd), stderr(cmd))
			ver, err := podman.Version(cmd.Context(), runner)
			if err != nil {
				ui.Warn(stdout(cmd), s.Styles, "podman not found — run: lab podman install")
				return fmt.Errorf("podman not installed: %w", err)
			}
			ui.Section(stdout(cmd), s.Styles, "podman version", ver)
			return nil
		},
	}

	cmd.AddCommand(install, configure, doctor, upgrade, remove, version)
	return cmd
}

func runPodmanConfigure(cmd *cobra.Command, info platform.Info, allowUserNS, enableSocket bool) error {
	s := session(cmd)
	runner := exec.NewOSRunner(stdout(cmd), stderr(cmd))
	results, err := podman.Configure(cmd.Context(), runner, podman.ConfigureOpts{
		Info:         info,
		DryRun:       s.DryRun,
		AllowUserNS:  allowUserNS,
		EnableSocket: enableSocket,
		Stdout:       stdout(cmd),
		Stderr:       stderr(cmd),
	})
	if err != nil {
		return err
	}
	for _, r := range results {
		line := r.Name + ": " + r.Detail
		switch {
		case r.Skipped:
			ui.Warn(stdout(cmd), s.Styles, line)
		case r.Changed:
			ui.OK(stdout(cmd), s.Styles, line)
		default:
			_, _ = fmt.Fprintln(stdout(cmd), s.Styles.Dim.Render("· "+line))
		}
	}
	if info.GOOS == platform.OSDarwin {
		ui.Warn(stdout(cmd), s.Styles, "macOS: use podman machine; Quadlet/rootless Linux steps do not apply")
	}
	return nil
}
