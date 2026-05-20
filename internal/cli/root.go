package cli

import (
	"context"
	"github.com/bartrosa/homelab-cli/internal/cli/appctx"
	"github.com/bartrosa/homelab-cli/internal/cli/commands"
	"github.com/bartrosa/homelab-cli/internal/config"
	"github.com/bartrosa/homelab-cli/internal/logging"
	"github.com/bartrosa/homelab-cli/internal/ui"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// Execute builds the root command and runs it with ctx.
func Execute(ctx context.Context) error {
	return NewRootCmd().ExecuteContext(ctx)
}

// NewRootCmd wires the full command tree for the lab CLI.
func NewRootCmd() *cobra.Command {
	var (
		configPath  string
		logLevel    string
		logFormat   string
		noColor     bool
		dryRun      bool
		homelabRoot string
	)

	root := &cobra.Command{
		Use:   "lab",
		Short: "Homelab automation from bare metal to GPU-served LLMs",
		Long: `lab is a single entry point for bootstrapping machines, managing language toolchains,
running local data services, mirroring repositories, operating clusters, and supporting ML/AI workflows.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			path, err := resolveConfigPath(cmd, configPath)
			if err != nil {
				return err
			}

			cfg, err := config.Load(path)
			if err != nil {
				return err
			}

			level := cfg.LogLevel
			if cmd.Root().PersistentFlags().Lookup("log-level").Changed {
				level = logLevel
			}

			format := cfg.LogFormat
			if cmd.Root().PersistentFlags().Lookup("log-format").Changed {
				format = logFormat
			}

			logger := logging.New(cmd.ErrOrStderr(), level, format, noColor)
			ctx := logging.WithLogger(cmd.Context(), logger)
			session := &appctx.Session{
				Config:      cfg,
				ConfigPath:  path,
				DryRun:      dryRun,
				NoColor:     noColor,
				Styles:      ui.NewStyles(cmd.OutOrStdout(), noColor),
				HomelabRoot: homelabRoot,
			}
			cmd.SetContext(appctx.WithSession(ctx, session))
			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	root.PersistentFlags().StringVar(&configPath, "config", "", "path to config file (default: ~/.config/homelab-cli/config.yaml)")
	root.PersistentFlags().StringVar(&logLevel, "log-level", "info", "log level (debug|info|warn|error)")
	root.PersistentFlags().StringVar(&logFormat, "log-format", "text", "log format (text|json)")
	root.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colorized output")
	root.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "print planned actions without executing external commands")
	root.PersistentFlags().StringVar(&homelabRoot, "homelab-root", "", "path to homelab repo (overrides config homelab.root and LAB_HOMELAB_ROOT)")

	root.AddGroup(&cobra.Group{ID: "foundation", Title: "Foundation — bootstrap & install"})
	root.AddGroup(&cobra.Group{ID: "repos", Title: "Repos — multi-repo management"})
	root.AddGroup(&cobra.Group{ID: "infra", Title: "Infra & networking"})
	root.AddGroup(&cobra.Group{ID: "data", Title: "Data / AI / ML / MLOps"})
	root.AddGroup(&cobra.Group{ID: "workflow", Title: "Workflow & observability"})
	root.AddGroup(&cobra.Group{ID: "meta", Title: "Meta"})

	add := func(cmd *cobra.Command, group string) {
		cmd.GroupID = group
		root.AddCommand(cmd)
	}

	add(commands.NewBootstrapCmd(), "foundation")
	add(commands.NewPkgCmd(), "foundation")
	add(commands.NewToolchainCmd(), "foundation")
	add(commands.NewServicesCmd(), "foundation")

	add(commands.NewReposCmd(), "repos")

	add(commands.NewClusterCmd(), "infra")
	add(commands.NewGPUCmd(), "infra")
	add(commands.NewServerCmd(), "infra")
	add(commands.NewSSHCmd(), "infra")
	add(commands.NewPostgresCmd(), "infra")
	add(commands.NewBaremetalCmd(), "infra")
	add(commands.NewSystemCmd(), "infra")
	add(commands.NewContainersCmd(), "infra")
	add(commands.NewNetCmd(), "infra")
	add(commands.NewStorageCmd(), "infra")

	add(commands.NewModelsCmd(), "data")
	add(commands.NewDataCmd(), "data")
	add(commands.NewNotebooksCmd(), "data")
	add(commands.NewMLOpsCmd(), "data")
	add(commands.NewVectorCmd(), "data")
	add(commands.NewPipelinesCmd(), "data")
	add(commands.NewAgentsCmd(), "data")

	add(commands.NewObsCmd(), "workflow")
	add(commands.NewLogsCmd(), "workflow")
	add(commands.NewTemplatesCmd(), "workflow")
	add(commands.NewMediaCmd(), "workflow")
	add(commands.NewMCPCmd(), "workflow")

	add(commands.NewVersionCmd(), "meta")

	return root
}

func resolveConfigPath(cmd *cobra.Command, flagPath string) (string, error) {
	if cmd.Root().PersistentFlags().Lookup("config").Changed && strings.TrimSpace(flagPath) != "" {
		return flagPath, nil
	}

	if strings.TrimSpace(flagPath) != "" {
		// Defensive: flag set without Changed (should not happen) still honors explicit path.
		return flagPath, nil
	}

	path, err := config.DefaultConfigPath()
	if err != nil {
		return "", fmt.Errorf("resolve default config path: %w", err)
	}

	// When --config not passed, flag default is empty string; use default path even if file missing.
	if !cmd.Root().PersistentFlags().Lookup("config").Changed {
		return path, nil
	}

	if strings.TrimSpace(flagPath) == "" {
		return "", errors.New("--config was set to an empty path")
	}

	return flagPath, nil
}

// PrintCommandError writes errors to stderr when SilenceErrors is enabled on the root command.
func PrintCommandError(err error) {
	fmt.Fprintln(os.Stderr, err)
}
