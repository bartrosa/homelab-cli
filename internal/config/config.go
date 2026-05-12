// Package config loads lab configuration from YAML and environment variables.
package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config is the top-level lab configuration (file + env); flags override in the CLI layer.
type Config struct {
	LogLevel  string `mapstructure:"log_level"`
	LogFormat string `mapstructure:"log_format"`

	Bootstrap BootstrapConfig `mapstructure:"bootstrap"`
	Repos     ReposConfig     `mapstructure:"repos"`
	Services  ServicesConfig  `mapstructure:"services"`
	Cluster   ClusterConfig   `mapstructure:"cluster"`
	Storage   StorageConfig   `mapstructure:"storage"`
}

// BootstrapConfig describes bootstrap profiles and defaults.
type BootstrapConfig struct {
	DefaultProfile string         `mapstructure:"default_profile"`
	Profiles       map[string]any `mapstructure:"profiles"`
}

// ReposConfig describes multi-repo roots and providers.
type ReposConfig struct {
	Providers []RepoProvider `mapstructure:"providers"`
	Root      string         `mapstructure:"root"`
	BackupDir string         `mapstructure:"backup_dir"`
}

// RepoProvider is a single Git hosting integration.
type RepoProvider struct {
	Name     string `mapstructure:"name"`
	Kind     string `mapstructure:"kind"`
	Host     string `mapstructure:"host"`
	TokenEnv string `mapstructure:"token_env"`
}

// ServicesConfig describes local compose stacks and runtime.
type ServicesConfig struct {
	StacksDir string `mapstructure:"stacks_dir"`
	Runtime   string `mapstructure:"runtime"`
}

// ClusterConfig describes Kubernetes client defaults.
type ClusterConfig struct {
	Kubeconfig string `mapstructure:"kubeconfig"`
	Context    string `mapstructure:"context"`
}

// StorageConfig describes S3-compatible endpoints for backups and artifacts.
type StorageConfig struct {
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"access_key"`
}

// Default returns baseline configuration used when no file exists.
func Default() *Config {
	return &Config{
		LogLevel:  "info",
		LogFormat: "text",
		Bootstrap: BootstrapConfig{
			DefaultProfile: "default",
			Profiles:       map[string]any{},
		},
		Repos: ReposConfig{
			Providers: nil,
			Root:      "~/src",
			BackupDir: "~/backups/repos",
		},
		Services: ServicesConfig{
			StacksDir: "~/.config/homelab-cli/stacks",
			Runtime:   "podman",
		},
		Cluster: ClusterConfig{
			Kubeconfig: defaultKubeconfigPath(),
			Context:    "",
		},
		Storage: StorageConfig{
			Endpoint:  "",
			AccessKey: "",
		},
	}
}

func defaultKubeconfigPath() string {
	if v := os.Getenv("KUBECONFIG"); v != "" {
		return v
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%s/.kube/config", home)
}

// Load reads YAML from path when the file exists, applies env (LAB_*), and unmarshals into Config.
// Precedence for overlapping keys is handled by the caller (CLI flags > env > file > defaults).
func Load(configPath string) (*Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")

	if strings.TrimSpace(configPath) == "" {
		return nil, errors.New("config path is empty")
	}

	v.SetEnvPrefix("LAB")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	bindDefaults(v)

	if _, err := os.Stat(configPath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("stat config: %w", err)
		}
	} else {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	var out Config
	if err := v.Unmarshal(&out); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	normalize(&out)
	return &out, nil
}

func bindDefaults(v *viper.Viper) {
	d := Default()
	v.SetDefault("log_level", d.LogLevel)
	v.SetDefault("log_format", d.LogFormat)
	v.SetDefault("bootstrap.default_profile", d.Bootstrap.DefaultProfile)
	v.SetDefault("repos.root", d.Repos.Root)
	v.SetDefault("repos.backup_dir", d.Repos.BackupDir)
	v.SetDefault("services.stacks_dir", d.Services.StacksDir)
	v.SetDefault("services.runtime", d.Services.Runtime)
	v.SetDefault("cluster.kubeconfig", d.Cluster.Kubeconfig)
	v.SetDefault("cluster.context", d.Cluster.Context)
	v.SetDefault("storage.endpoint", d.Storage.Endpoint)
	v.SetDefault("storage.access_key", d.Storage.AccessKey)
}

func normalize(c *Config) {
	if c.LogLevel == "" {
		c.LogLevel = Default().LogLevel
	}
	if c.LogFormat == "" {
		c.LogFormat = Default().LogFormat
	}
	if c.Services.Runtime == "" {
		c.Services.Runtime = Default().Services.Runtime
	}
	if c.Cluster.Kubeconfig == "" {
		c.Cluster.Kubeconfig = Default().Cluster.Kubeconfig
	}
}

// DefaultConfigPath returns ~/.config/homelab-cli/config.yaml when $HOME is available.
func DefaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/.config/homelab-cli/config.yaml", home), nil
}
