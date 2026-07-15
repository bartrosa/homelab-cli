package stack

import (
	"io"

	"github.com/bartrosa/homelab-cli/internal/exec"
	"github.com/bartrosa/homelab-cli/internal/pkgmgr"
	"github.com/bartrosa/homelab-cli/internal/platform"
)

// Env carries runtime dependencies for stack components.
type Env struct {
	Runner  exec.Runner
	Stdout  io.Writer
	Stderr  io.Writer
	Info    platform.Info
	PkgMgr  pkgmgr.Manager
	HomeDir string
}
