package stack

import (
	"fmt"
	"sort"
	"sync"
)

var (
	mu       sync.RWMutex
	registry = map[string]Component{}
)

// Register adds a component to the global registry.
func Register(c Component) {
	if c == nil {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	registry[c.ID()] = c
}

// Lookup returns a component by id.
func Lookup(id string) (Component, bool) {
	mu.RLock()
	defer mu.RUnlock()
	c, ok := registry[id]
	return c, ok
}

// All returns all registered components sorted by id.
func All() []Component {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]Component, 0, len(registry))
	for _, c := range registry {
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID() < out[j].ID() })
	return out
}

// ByCategory returns components in a category.
func ByCategory(cat Category) []Component {
	var out []Component
	for _, c := range All() {
		if c.Category() == cat {
			out = append(out, c)
		}
	}
	return out
}

// IDs validates and returns component ids.
func IDs(names ...string) ([]Component, error) {
	var out []Component
	for _, name := range names {
		c, ok := Lookup(name)
		if !ok {
			return nil, fmt.Errorf("unknown stack component %q", name)
		}
		out = append(out, c)
	}
	return out, nil
}
