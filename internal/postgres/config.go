package postgres

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config is the root YAML structure (instances.yaml).
type Config struct {
	Instances []Instance `yaml:"instances"`
}

// Instance is one PostgreSQL server.
type Instance struct {
	Host      string     `yaml:"host"`
	Port      int        `yaml:"port"`
	Databases []Database `yaml:"databases"`
	Users     []User     `yaml:"users"`
}

// Database desired state.
type Database struct {
	Name  string `yaml:"name"`
	Owner string `yaml:"owner"`
}

// User desired state.
type User struct {
	Name      string   `yaml:"name"`
	Password  string   `yaml:"password"`
	Databases []string `yaml:"databases"`
}

// LoadConfig reads instances YAML.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cfg, nil
}
