// Package pkgmgr abstracts native package managers for bootstrap.
package pkgmgr

import (
	"context"
)

// Manager installs packages on a target OS.
type Manager interface {
	Name() string
	Available() bool
	UpdateCache(ctx context.Context) error
	Install(ctx context.Context, packages ...string) error
	IsInstalled(ctx context.Context, pkg string) (bool, error)
}
