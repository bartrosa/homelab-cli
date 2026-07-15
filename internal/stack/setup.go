package stack

import (
	"io"
	"os"

	"github.com/bartrosa/homelab-cli/internal/exec"
	"github.com/bartrosa/homelab-cli/internal/pkgmgr"
	"github.com/bartrosa/homelab-cli/internal/platform"
)

// NewEnv builds a component environment.
func NewEnv(stdout, stderr io.Writer) (*Env, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	info := platform.Detect()
	var mgr pkgmgr.Manager
	switch {
	case info.IsSilverblue:
		mgr = &pkgmgr.RPMOstree{Runner: exec.NewOSRunner(stdout, stderr)}
	case info.Packager == platform.PackagerAPT:
		mgr = &pkgmgr.APT{Runner: exec.NewOSRunner(stdout, stderr), Sudo: true}
	default:
		mgr = nil
	}
	return &Env{
		Runner:  exec.NewOSRunner(stdout, stderr),
		Stdout:  stdout,
		Stderr:  stderr,
		Info:    info,
		PkgMgr:  mgr,
		HomeDir: home,
	}, nil
}
