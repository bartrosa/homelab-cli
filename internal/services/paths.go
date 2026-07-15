package services

import (
	"os"
	"path/filepath"
)

const appName = "homelab-cli"

// ConfigDir returns XDG config dir for a service id.
func ConfigDir(id string) (string, error) {
	base, err := xdgConfigHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, appName, "services", id), nil
}

// DataDir returns XDG data dir for a service id.
func DataDir(id string) (string, error) {
	base, err := xdgDataHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, appName, "services", id), nil
}

// StateDir returns XDG state dir for a service id (generated compose, .env).
func StateDir(id string) (string, error) {
	base, err := xdgStateHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, appName, "services", id), nil
}

func xdgConfigHome() (string, error) {
	if v := os.Getenv("XDG_CONFIG_HOME"); v != "" {
		return v, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config"), nil
}

func xdgDataHome() (string, error) {
	if v := os.Getenv("XDG_DATA_HOME"); v != "" {
		return v, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share"), nil
}

func xdgStateHome() (string, error) {
	if v := os.Getenv("XDG_STATE_HOME"); v != "" {
		return v, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "state"), nil
}
