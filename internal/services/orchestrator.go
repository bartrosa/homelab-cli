package services

import (
	"context"
	"fmt"
	"strings"
)

// Orchestrator coordinates init/up/down across services with dependency ordering.
type Orchestrator struct {
	CustomPresets map[string][]string
}

// Init initializes one or more services or presets.
func (o *Orchestrator) Init(ctx context.Context, opts InitOptions, names ...string) error {
	ids, err := o.expand(names...)
	if err != nil {
		return err
	}
	order, err := resolveServiceOrder(ids)
	if err != nil {
		return err
	}
	for _, id := range order {
		s, ok := Lookup(id)
		if !ok {
			return fmtUnknownService(id)
		}
		if err := s.Init(ctx, opts); err != nil {
			return fmt.Errorf("%s: init: %w", id, err)
		}
	}
	return nil
}

// Up starts services in dependency order (grafana last among observability).
func (o *Orchestrator) Up(ctx context.Context, opts InitOptions, names ...string) error {
	ids, err := o.expand(names...)
	if err != nil {
		return err
	}
	order, err := resolveServiceOrder(ids)
	if err != nil {
		return err
	}
	for _, id := range order {
		s, ok := Lookup(id)
		if !ok {
			return fmtUnknownService(id)
		}
		if err := s.Up(ctx, opts); err != nil {
			return fmt.Errorf("%s: up: %w", id, err)
		}
	}
	return nil
}

// Down stops services in reverse dependency order.
func (o *Orchestrator) Down(ctx context.Context, opts InitOptions, names ...string) error {
	ids, err := o.expand(names...)
	if err != nil {
		return err
	}
	order, err := resolveServiceOrder(ids)
	if err != nil {
		return err
	}
	for i := len(order) - 1; i >= 0; i-- {
		id := order[i]
		s, ok := Lookup(id)
		if !ok {
			return fmtUnknownService(id)
		}
		if err := s.Down(ctx, opts); err != nil {
			return fmt.Errorf("%s: down: %w", id, err)
		}
	}
	return nil
}

func (o *Orchestrator) expand(names ...string) ([]string, error) {
	return ExpandNames(names, o.CustomPresets)
}

func resolveServiceOrder(ids []string) ([]string, error) {
	seen := map[string]struct{}{}
	var order []string
	var visit func(string) error
	visit = func(id string) error {
		if _, ok := seen[id]; ok {
			return nil
		}
		s, ok := Lookup(id)
		if !ok {
			return fmtUnknownService(id)
		}
		deps := s.DependsOn()
		// Grafana starts after other observability backends when present.
		if id == "grafana" {
			deps = appendUnique(deps, "prometheus", "loki", "tempo")
		}
		for _, dep := range deps {
			dep = strings.TrimSpace(dep)
			if dep == "" {
				continue
			}
			if contains(ids, dep) || serviceRegistered(dep) {
				if err := visit(dep); err != nil {
					return err
				}
			}
		}
		seen[id] = struct{}{}
		order = append(order, id)
		return nil
	}
	for _, id := range ids {
		if err := visit(id); err != nil {
			return nil, err
		}
	}
	return order, nil
}

func serviceRegistered(id string) bool {
	_, ok := Lookup(id)
	return ok
}

func contains(list []string, item string) bool {
	for _, v := range list {
		if v == item {
			return true
		}
	}
	return false
}

func appendUnique(base []string, items ...string) []string {
	seen := map[string]struct{}{}
	for _, b := range base {
		seen[b] = struct{}{}
	}
	for _, item := range items {
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		base = append(base, item)
	}
	return base
}
