package commands

import (
	"fmt"
	"strings"

	"github.com/bartrosa/homelab-cli/internal/services"
	"github.com/bartrosa/homelab-cli/internal/stack"
	"github.com/bartrosa/homelab-cli/internal/stack/gpu"
	"github.com/bartrosa/homelab-cli/internal/stack/shellrc"
	"github.com/bartrosa/homelab-cli/internal/ui"
	"github.com/spf13/cobra"
)

// NewStackCmd wires lab stack developer environment commands.
func NewStackCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "stack",
		Aliases: []string{"toolchain", "tc"},
		Short:   "Install developer stack components (languages, tools, GPU, embedded DBs)",
		Long: `stack installs and manages developer environment components: language toolchains,
build tools, container runtimes, GPU stacks, and embedded databases.

Alias: lab toolchain (deprecated name, same commands).`,
		RunE: func(c *cobra.Command, _ []string) error {
			return c.Help()
		},
	}
	AddDryRunFlag(cmd)
	cmd.AddCommand(newStackListCmd(), newStackListInstalledCmd(), newStackInfoCmd(), newStackGPUCmd(),
		newStackInstallCmd(), newStackPresetListCmd(), newStackPresetShowCmd(),
		newStackPathCmd(), newStackPathRefreshCmd(), newStackPathRemoveCmd(),
		newStackUseCmd()) // legacy use from toolchain
	return cmd
}

// NewToolchainCmd is an alias for stack (backward compatibility).
func NewToolchainCmd() *cobra.Command { return NewStackCmd() }

func stackEnv(cmd *cobra.Command) (*stack.Env, error) {
	return stack.NewEnv(stdout(cmd), stderr(cmd))
}

func stackOpts(cmd *cobra.Command) stack.InstallOptions {
	s := session(cmd)
	yes, _ := cmd.Flags().GetBool("yes")
	skipPath, _ := cmd.Flags().GetBool("skip-path")
	force, _ := cmd.Flags().GetBool("force")
	version, _ := cmd.Flags().GetBool("version")
	_ = version
	ver, _ := cmd.Flags().GetString("component-version")
	return stack.InstallOptions{
		Force:          force,
		NonInteractive: yes,
		DryRun:         s.DryRun,
		SkipPath:       skipPath,
		Version:        ver,
	}
}

func addStackInstallFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("yes", false, "non-interactive")
	cmd.Flags().Bool("force", false, "force reinstall")
	cmd.Flags().Bool("skip-path", false, "skip shell rc PATH update")
	cmd.Flags().String("component-version", "", "override component version")
	cmd.Flags().String("preset", "", "install a named preset bundle")
	cmd.Flags().Bool("gpu", false, "with --preset ml: add cuda or rocm based on detected GPU")
}

func newStackListCmd() *cobra.Command {
	var category string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available stack components",
		RunE: func(cmd *cobra.Command, _ []string) error {
			s := session(cmd)
			ui.Section(stdout(cmd), s.Styles, "Stack components", "")
			var rows [][]string
			for _, c := range stack.All() {
				if category != "" && string(c.Category()) != category {
					continue
				}
				rows = append(rows, []string{string(c.Category()), c.ID(), c.DisplayName(), c.DefaultVersion()})
			}
			ui.Table(stdout(cmd), s.Styles, []string{"CATEGORY", "ID", "NAME", "DEFAULT"}, rows)
			return nil
		},
	}
	cmd.Flags().StringVar(&category, "category", "", "filter by category")
	return cmd
}

func newStackListInstalledCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list-installed",
		Short: "List installed components and versions",
		RunE: func(cmd *cobra.Command, _ []string) error {
			env, err := stackEnv(cmd)
			if err != nil {
				return err
			}
			list, err := stack.ListInstalled(cmd.Context(), env)
			if err != nil {
				return err
			}
			s := session(cmd)
			var rows [][]string
			for _, item := range list {
				rows = append(rows, []string{item.ID, item.Version})
			}
			ui.Table(stdout(cmd), s.Styles, []string{"ID", "VERSION"}, rows)
			return nil
		},
	}
}

func newStackInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info <component>",
		Short: "Show component details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ok := stack.Lookup(args[0])
			if !ok {
				return fmt.Errorf("unknown component %q", args[0])
			}
			s := session(cmd)
			_, _ = fmt.Fprintf(stdout(cmd), "ID:          %s\n", c.ID())
			_, _ = fmt.Fprintf(stdout(cmd), "Name:        %s\n", c.DisplayName())
			_, _ = fmt.Fprintf(stdout(cmd), "Category:    %s\n", c.Category())
			_, _ = fmt.Fprintf(stdout(cmd), "Description: %s\n", c.Description())
			_, _ = fmt.Fprintf(stdout(cmd), "Requires:    %s\n", strings.Join(c.Requires(), ", "))
			_ = s
			return nil
		},
	}
}

func newStackGPUCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "gpu",
		Short: "Show detected GPUs and available compute stacks",
		RunE: func(cmd *cobra.Command, _ []string) error {
			env, err := stackEnv(cmd)
			if err != nil {
				return err
			}
			gpus, err := gpu.Detect(cmd.Context(), env.Runner)
			if err != nil {
				return err
			}
			w := stdout(cmd)
			if len(gpus) == 0 {
				fmt.Fprintln(w, "No GPUs detected.")
				return nil
			}
			fmt.Fprintln(w, "Detected GPUs:")
			for i, g := range gpus {
				fmt.Fprintf(w, "  [%d] %s (vendor: %s)\n", i, g.Model, g.Vendor)
			}
			fmt.Fprintln(w, "\nCompute stacks:")
			fmt.Fprintln(w, "  cuda (NVIDIA)  → lab stack install cuda")
			fmt.Fprintln(w, "  rocm (AMD)     → lab stack install rocm")
			return nil
		},
	}
}

func newStackInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "install [component...]",
		Short:   "Install stack components or a preset",
		Example: "  lab stack install --preset ml --yes\n  lab stack install python uv rust --yes",
		RunE: func(cmd *cobra.Command, args []string) error {
			env, err := stackEnv(cmd)
			if err != nil {
				return err
			}
			s := session(cmd)
			opts := stackOpts(cmd)
			preset, _ := cmd.Flags().GetString("preset")
			withGPU, _ := cmd.Flags().GetBool("gpu")
			var ids []string
			if preset != "" {
				ids, err = stack.ResolvePreset(preset, s.Config.Stack.Presets)
				if err != nil {
					return err
				}
				if withGPU {
					ok, _ := gpu.DetectNvidia(cmd.Context(), env.Runner)
					if ok {
						ids = append(ids, "cuda")
					} else if ok, _ := gpu.DetectAmd(cmd.Context(), env.Runner); ok {
						ids = append(ids, "rocm")
					}
				}
			} else {
				ids = args
			}
			if len(ids) == 0 {
				return fmt.Errorf("specify components or --preset")
			}
			ui.Section(stdout(cmd), s.Styles, "stack install", strings.Join(ids, ", "))
			return stack.InstallAll(cmd.Context(), env, ids, opts)
		},
	}
	addStackInstallFlags(cmd)
	return cmd
}

func newStackPresetListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "preset list",
		Short: "List stack presets",
		RunE: func(cmd *cobra.Command, _ []string) error {
			s := session(cmd)
			for _, name := range stack.PresetNames(s.Config.Stack.Presets) {
				ids, _ := stack.ResolvePreset(name, s.Config.Stack.Presets)
				fmt.Fprintf(stdout(cmd), "  %-14s %s\n", name, strings.Join(ids, ", "))
			}
			return nil
		},
	}
}

func newStackPresetShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "preset show <name>",
		Short: "Show components in a preset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := session(cmd)
			ids, err := stack.ResolvePreset(args[0], s.Config.Stack.Presets)
			if err != nil {
				return err
			}
			for _, id := range ids {
				fmt.Fprintln(stdout(cmd), id)
			}
			_ = s
			return nil
		},
	}
}

func newStackPathCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Show managed shell PATH block",
		RunE: func(cmd *cobra.Command, _ []string) error {
			shells, err := shellrc.Detect()
			if err != nil {
				return err
			}
			for _, sh := range shells {
				block, err := shellrc.ReadBlock(sh)
				if err != nil {
					return err
				}
				if block != "" {
					fmt.Fprintln(stdout(cmd), block)
				}
			}
			return nil
		},
	}
}

func newStackPathRefreshCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "path refresh",
		Short: "Regenerate shell PATH block from installed components",
		RunE: func(cmd *cobra.Command, _ []string) error {
			env, err := stackEnv(cmd)
			if err != nil {
				return err
			}
			return stack.RefreshPath(cmd.Context(), env)
		},
	}
}

func newStackPathRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "path remove",
		Short: "Remove managed shell PATH block",
		RunE: func(_ *cobra.Command, _ []string) error {
			shells, err := shellrc.Detect()
			if err != nil {
				return err
			}
			for _, sh := range shells {
				if err := shellrc.RemoveBlock(sh); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

func newStackUseCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "use <lang> <version>",
		Short:   "Switch active mise toolchain version (legacy)",
		Example: "  lab stack use go 1.25.0",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			env, err := stackEnv(cmd)
			if err != nil {
				return err
			}
			return env.Runner.Run(cmd.Context(), "mise", "use", "-g", args[0]+"@"+args[1])
		},
	}
}

// unused import guard for services orchestrator wiring in install preset
var _ = services.Orchestrator{}
