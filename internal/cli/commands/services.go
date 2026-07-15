package commands

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/bartrosa/homelab-cli/internal/exec"
	"github.com/bartrosa/homelab-cli/internal/homelabroot"
	"github.com/bartrosa/homelab-cli/internal/mlstack"
	"github.com/bartrosa/homelab-cli/internal/prompt"
	"github.com/bartrosa/homelab-cli/internal/services"
	"github.com/bartrosa/homelab-cli/internal/ui"
	"github.com/spf13/cobra"
)

// NewServicesCmd wires local compose-backed data services.
func NewServicesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "services",
		Short: "Manage local data services (Postgres, Redis, observability, …)",
		Long: `services provisions and runs compose stacks for databases, caches, vector DBs,
observability, and object storage on homelab-net.`,
		RunE: func(c *cobra.Command, _ []string) error {
			return c.Help()
		},
	}
	AddDryRunFlag(cmd)

	cmd.AddCommand(
		newServicesListCmd(),
		newServicesInfoCmd(),
		newServicesInitCmd(),
		newServicesUpCmd(),
		newServicesDownCmd(),
		newServicesRestartCmd(),
		newServicesStatusCmd(),
		newServicesLogsCmd(),
		newServicesConnectCmd(),
		newServicesRMCmd(),
		newServicesPresetListCmd(),
		newServicesPresetShowCmd(),
		newServicesEnsureCmd(),
	)
	return cmd
}

func svcOpts(cmd *cobra.Command) services.InitOptions {
	s := session(cmd)
	yes, _ := cmd.Flags().GetBool("yes")
	force, _ := cmd.Flags().GetBool("force")
	setFlags, _ := cmd.Flags().GetStringSlice("set")
	values := parseSetFlags(setFlags)
	for id, inst := range s.Config.Services.Instances {
		for fk, fv := range inst {
			if _, ok := values[fk]; !ok {
				values[fk] = fv
			}
			_ = id
		}
	}
	return services.InitOptions{
		Runner:         exec.NewOSRunner(stdout(cmd), stderr(cmd)),
		Stdout:         stdout(cmd),
		Stderr:         stderr(cmd),
		Runtime:        s.Config.Services.Runtime,
		DryRun:         s.DryRun,
		Force:          force,
		NonInteractive: yes,
		Values:         values,
		Prompter:       prompt.NewStdinPrompter(),
	}
}

func svcOrchestrator(cmd *cobra.Command) *services.Orchestrator {
	s := session(cmd)
	return &services.Orchestrator{CustomPresets: s.Config.Services.Presets}
}

func parseSetFlags(pairs []string) map[string]any {
	out := map[string]any{}
	for _, p := range pairs {
		k, v, ok := strings.Cut(p, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		if strings.Contains(v, ",") {
			out[k] = strings.Split(v, ",")
		} else {
			out[k] = v
		}
	}
	return out
}

func addServiceFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("yes", false, "non-interactive")
	cmd.Flags().Bool("force", false, "overwrite existing config")
	cmd.Flags().StringSlice("set", nil, "config key=value (repeatable)")
	cmd.Flags().String("preset", "", "service preset name")
}

func newServicesListCmd() *cobra.Command {
	var category string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available services",
		RunE: func(cmd *cobra.Command, _ []string) error {
			s := session(cmd)
			ui.Section(stdout(cmd), s.Styles, "Services", "")
			opts := svcOpts(cmd)
			var rows [][]string
			for _, svc := range services.All() {
				if category != "" && string(svc.Category()) != category {
					continue
				}
				st, _ := svc.Status(cmd.Context(), opts)
				status := "stopped"
				if st.Running {
					status = "running"
				}
				rows = append(rows, []string{string(svc.Category()), svc.ID(), svc.DisplayName(), status})
			}
			ui.Table(stdout(cmd), s.Styles, []string{"CATEGORY", "ID", "NAME", "STATUS"}, rows)
			return nil
		},
	}
	cmd.Flags().StringVar(&category, "category", "", "filter by category")
	return cmd
}

func newServicesInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info <id>",
		Short: "Show service description and config schema",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, ok := services.Lookup(args[0])
			if !ok {
				return fmt.Errorf("unknown service %q", args[0])
			}
			w := stdout(cmd)
			fmt.Fprintf(w, "ID:          %s\n", svc.ID())
			fmt.Fprintf(w, "Name:        %s\n", svc.DisplayName())
			fmt.Fprintf(w, "Category:    %s\n", svc.Category())
			fmt.Fprintf(w, "Description: %s\n", svc.Description())
			fmt.Fprintln(w, "Config fields:")
			for _, f := range svc.Schema().Fields {
				fmt.Fprintf(w, "  - %s (%s) default=%v\n", f.Name, f.Type, f.Default)
			}
			return nil
		},
	}
}

func newServicesInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init <id> [more...]",
		Short: "Initialize service config (interactive or --set)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			preset, _ := cmd.Flags().GetString("preset")
			if preset != "" {
				args = []string{preset}
			}
			return svcOrchestrator(cmd).Init(cmd.Context(), svcOpts(cmd), args...)
		},
	}
	addServiceFlags(cmd)
	return cmd
}

func newServicesUpCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "up <id> [more...]",
		Short:   "Start services",
		Example: "  lab services up postgres\n  lab services up --preset observability --yes",
		RunE: func(cmd *cobra.Command, args []string) error {
			preset, _ := cmd.Flags().GetString("preset")
			if preset != "" {
				args = []string{preset}
			}
			if len(args) == 0 {
				return fmt.Errorf("specify service id(s) or --preset")
			}
			s := session(cmd)
			ui.Section(stdout(cmd), s.Styles, "services up", strings.Join(args, ", "))
			return svcOrchestrator(cmd).Up(cmd.Context(), svcOpts(cmd), args...)
		},
	}
	addServiceFlags(cmd)
	return cmd
}

func newServicesDownCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "down <id> [more...]",
		Short: "Stop services",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("specify service id(s)")
			}
			return svcOrchestrator(cmd).Down(cmd.Context(), svcOpts(cmd), args...)
		},
	}
	return cmd
}

func newServicesRestartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "restart <id>",
		Short: "Restart a service",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			o := svcOrchestrator(cmd)
			opts := svcOpts(cmd)
			if err := o.Down(cmd.Context(), opts, args[0]); err != nil {
				return err
			}
			return o.Up(cmd.Context(), opts, args[0])
		},
	}
}

func newServicesStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status [id]",
		Short: "Show service runtime status",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := svcOpts(cmd)
			ids := args
			if len(ids) == 0 {
				for _, s := range services.All() {
					ids = append(ids, s.ID())
				}
			}
			for _, id := range ids {
				svc, ok := services.Lookup(id)
				if !ok {
					return fmt.Errorf("unknown service %q", id)
				}
				st, err := svc.Status(cmd.Context(), opts)
				if err != nil {
					return err
				}
				state := "stopped"
				if st.Running {
					state = "running"
				}
				fmt.Fprintf(stdout(cmd), "%s: %s %s\n", id, state, st.Detail)
			}
			return nil
		},
	}
}

func newServicesLogsCmd() *cobra.Command {
	var follow bool
	var tail int
	cmd := &cobra.Command{
		Use:   "logs <id>",
		Short: "Tail service logs",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			_ = follow
			_ = tail
			dir, err := services.StateDir(args[0])
			if err != nil {
				return err
			}
			return fmt.Errorf("use compose logs in %s (lab services logs TUI pending)", dir)
		},
	}
	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "follow log output")
	cmd.Flags().IntVar(&tail, "tail", 100, "number of lines")
	return cmd
}

func newServicesConnectCmd() *cobra.Command {
	var interactive bool
	cmd := &cobra.Command{
		Use:   "connect <id>",
		Short: "Print or open a connection to a service",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, ok := services.Lookup(args[0])
			if !ok {
				return fmt.Errorf("unknown service %q", args[0])
			}
			if c, ok := svc.(interface {
				Connect(context.Context, services.InitOptions, bool) error
			}); ok {
				return c.Connect(cmd.Context(), svcOpts(cmd), interactive)
			}
			return fmt.Errorf("connect not implemented for %q", args[0])
		},
	}
	cmd.Flags().BoolVar(&interactive, "interactive", false, "open interactive client session")
	return cmd
}

func newServicesRMCmd() *cobra.Command {
	var withData bool
	cmd := &cobra.Command{
		Use:   "rm <id>",
		Short: "Remove service configuration",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			stateDir, err := services.StateDir(args[0])
			if err != nil {
				return err
			}
			if withData {
				dataDir, err := services.DataDir(args[0])
				if err == nil {
					_ = os.RemoveAll(dataDir)
				}
			}
			return os.RemoveAll(stateDir)
		},
	}
	cmd.Flags().BoolVar(&withData, "data", false, "also remove persistent data")
	return cmd
}

func newServicesPresetListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "preset list",
		Short: "List service presets",
		RunE: func(cmd *cobra.Command, _ []string) error {
			s := session(cmd)
			for _, name := range services.PresetNames(s.Config.Services.Presets) {
				ids, _ := services.ResolvePreset(name, s.Config.Services.Presets)
				fmt.Fprintf(stdout(cmd), "  %-16s %s\n", name, strings.Join(ids, ", "))
			}
			return nil
		},
	}
}

func newServicesPresetShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "preset show <name>",
		Short: "Show services in a preset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := session(cmd)
			ids, err := services.ResolvePreset(args[0], s.Config.Services.Presets)
			if err != nil {
				return err
			}
			for _, id := range ids {
				fmt.Fprintln(stdout(cmd), id)
			}
			return nil
		},
	}
}

func newServicesEnsureCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ensure [ml-stack]",
		Short: "Ensure homelab ml-stack is up (legacy homelab compose)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			setDryRun(cmd)
			s := session(cmd)
			root, err := homelabroot.Resolve(firstCLI(s.HomelabRoot, s.Config.Homelab.Root))
			if err != nil {
				return err
			}
			name := "ml-stack"
			if len(args) == 1 {
				name = args[0]
			}
			if name != "ml-stack" {
				return fmt.Errorf("only ml-stack is supported (got %q)", name)
			}
			ui.Section(stdout(cmd), s.Styles, "services ensure", name)
			mlDir := root + "/ml-stack"
			return mlstack.EnsureUp(cmd.Context(), mlDir, s.Config.Server.Host, s.DryRun, stdout(cmd), stderr(cmd))
		},
	}
}

var _ io.Writer = os.Stdout
