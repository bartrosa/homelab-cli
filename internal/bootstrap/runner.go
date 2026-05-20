// Package bootstrap executes setup profiles.
package bootstrap

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/bartrosa/homelab-cli/internal/executil"
	"github.com/bartrosa/homelab-cli/internal/homelabroot"
	"github.com/bartrosa/homelab-cli/internal/packager"
	"github.com/bartrosa/homelab-cli/internal/toolchain"
	"github.com/bartrosa/homelab-cli/internal/ui"
)

// Runner executes bootstrap profiles.
type Runner struct {
	Stdout      io.Writer
	Stderr      io.Writer
	HomelabRoot string
	DryRun      bool
	Styles      ui.Styles
}

// RunProfile applies all steps in order.
func (r *Runner) RunProfile(ctx context.Context, profile *Profile) error {
	if profile == nil {
		return fmt.Errorf("nil profile")
	}
	ui.Section(r.Stdout, r.Styles, "bootstrap "+profile.Name, profile.Description)

	pkgMgr := packager.New(r.Stdout, r.Stderr, r.DryRun)
	tc := toolchain.New(r.Stdout, r.Stderr, r.DryRun)

	total := len(profile.Steps)
	for i, step := range profile.Steps {
		ui.Step(r.Stdout, r.Styles, i+1, total, string(step.Type))
		switch step.Type {
		case StepPkg:
			for _, pkg := range step.Packages {
				if err := pkgMgr.Ensure(ctx, pkg); err != nil {
					return fmt.Errorf("pkg %s: %w", pkg, err)
				}
				ui.OK(r.Stdout, r.Styles, pkg)
			}
		case StepToolchain:
			if err := tc.Install(ctx, step.Languages...); err != nil {
				return err
			}
			for _, lang := range step.Languages {
				ui.OK(r.Stdout, r.Styles, "toolchain "+lang)
			}
		case StepScript:
			if err := r.runScript(ctx, step.Script); err != nil {
				return err
			}
			ui.OK(r.Stdout, r.Styles, step.Script)
		default:
			return fmt.Errorf("unknown step type %q", step.Type)
		}
	}
	ui.OK(r.Stdout, r.Styles, "profile complete")
	return nil
}

func (r *Runner) runScript(ctx context.Context, rel string) error {
	root, err := homelabroot.Resolve(r.HomelabRoot)
	if err != nil {
		return err
	}
	path := filepath.Join(root, rel)
	ex := executil.NewRunner(r.Stdout, r.Stderr)
	ex.DryRun = r.DryRun
	ex.WorkDir = root
	return ex.Run(ctx, "bash", path)
}

// LaptopProfile picks the best built-in profile for the current OS.
func LaptopProfile() (string, error) {
	return "laptop-macos", nil // caller may override via platform
}
