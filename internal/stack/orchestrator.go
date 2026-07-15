package stack

import (
	"context"
	"fmt"
	"strings"

	"github.com/bartrosa/homelab-cli/internal/stack/shellrc"
)

// InstallAll resolves dependencies and installs components in order.
func InstallAll(ctx context.Context, env *Env, ids []string, opts InstallOptions) error {
	order, err := resolveOrder(ids)
	if err != nil {
		return err
	}
	total := len(order)
	for i, id := range order {
		c, ok := Lookup(id)
		if !ok {
			return fmt.Errorf("unknown component %q", id)
		}
		installed, ver, err := c.IsInstalled(ctx, env)
		if err != nil {
			return fmt.Errorf("%s: check installed: %w", id, err)
		}
		if installed && !opts.Force {
			fmt.Fprintf(env.Stdout, "[%d/%d] %s ... already installed (%s), skipping\n", i+1, total, id, ver)
			continue
		}
		if opts.DryRun {
			fmt.Fprintf(env.Stdout, "[%d/%d] %s ... would install\n", i+1, total, id)
			continue
		}
		fmt.Fprintf(env.Stdout, "[%d/%d] %s ... installing\n", i+1, total, id)
		if err := c.Install(ctx, env, opts); err != nil {
			return fmt.Errorf("%s: install: %w", id, err)
		}
		fmt.Fprintf(env.Stdout, "    ✅ Installed %s\n", id)
	}
	if !opts.DryRun && !opts.SkipPath {
		if err := refreshPath(ctx, env); err != nil {
			return err
		}
	}
	return nil
}

// Plan returns dry-run steps for ids.
func Plan(ctx context.Context, env *Env, ids []string, opts InstallOptions) ([]PlanStep, error) {
	order, err := resolveOrder(ids)
	if err != nil {
		return nil, err
	}
	var steps []PlanStep
	for _, id := range order {
		c, ok := Lookup(id)
		if !ok {
			return nil, fmt.Errorf("unknown component %q", id)
		}
		step := PlanStep{ID: id, Requires: c.Requires(), Version: c.DefaultVersion()}
		installed, ver, err := c.IsInstalled(ctx, env)
		if err != nil {
			return nil, err
		}
		if installed && !opts.Force {
			step.Action = "skip"
			step.Reason = "already installed (" + ver + ")"
		} else {
			step.Action = "install"
		}
		steps = append(steps, step)
	}
	return steps, nil
}

func resolveOrder(ids []string) ([]string, error) {
	seen := map[string]struct{}{}
	var order []string
	var visit func(string) error
	visit = func(id string) error {
		if _, ok := seen[id]; ok {
			return nil
		}
		c, ok := Lookup(id)
		if !ok {
			return fmt.Errorf("unknown component %q", id)
		}
		for _, req := range c.Requires() {
			if err := visit(req); err != nil {
				return err
			}
		}
		seen[id] = struct{}{}
		order = append(order, id)
		return nil
	}
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if err := visit(id); err != nil {
			return nil, err
		}
	}
	return order, nil
}

// RefreshPath regenerates shell rc blocks from registered component PathEntries.
func RefreshPath(ctx context.Context, env *Env) error {
	return refreshPath(ctx, env)
}

func refreshPath(_ context.Context, env *Env) error {
	entries := collectPathEntries()
	if len(entries) == 0 {
		return nil
	}
	shells, err := shellrc.Detect()
	if err != nil {
		return err
	}
	var updated []string
	for _, sh := range shells {
		rcEntries := make([]shellrc.Entry, len(entries))
		for i, e := range entries {
			rcEntries[i] = shellrc.Entry{Shell: e.Shell, Content: e.Content, Marker: e.Marker}
		}
		if err := shellrc.UpdateBlock(sh, rcEntries); err != nil {
			return err
		}
		p, _ := shellrc.RCPath(sh)
		updated = append(updated, p)
	}
	if len(updated) > 0 {
		fmt.Fprintf(env.Stdout, "Updated shell PATH in: %s\n", strings.Join(updated, ", "))
		fmt.Fprintln(env.Stdout, "👉 Run `source ~/.bashrc` or restart terminal to activate.")
	}
	return nil
}

func collectPathEntries() []PathEntry {
	seen := map[string]struct{}{}
	var out []PathEntry
	for _, c := range All() {
		for _, e := range c.PathEntries() {
			key := e.Marker + "|" + e.Shell
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, e)
		}
	}
	return out
}

// ListInstalled returns installed components with versions.
func ListInstalled(ctx context.Context, env *Env) ([]struct {
	ID, Version string
}, error,
) {
	var out []struct{ ID, Version string }
	for _, c := range All() {
		ok, ver, err := c.IsInstalled(ctx, env)
		if err != nil {
			return nil, err
		}
		if ok {
			out = append(out, struct{ ID, Version string }{c.ID(), ver})
		}
	}
	return out, nil
}
