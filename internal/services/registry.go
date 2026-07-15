package services

import (
	"sort"
	"sync"
)

var (
	regMu sync.RWMutex
	reg   = map[string]Service{}
)

// Register adds a service to the global registry.
func Register(s Service) {
	if s == nil {
		return
	}
	regMu.Lock()
	defer regMu.Unlock()
	reg[s.ID()] = s
}

// Lookup returns a service by id.
func Lookup(id string) (Service, bool) {
	regMu.RLock()
	defer regMu.RUnlock()
	s, ok := reg[id]
	return s, ok
}

// All returns registered services sorted by id.
func All() []Service {
	regMu.RLock()
	defer regMu.RUnlock()
	out := make([]Service, 0, len(reg))
	for _, s := range reg {
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID() < out[j].ID() })
	return out
}

// ByCategory returns services in a category.
func ByCategory(cat Category) []Service {
	var out []Service
	for _, s := range All() {
		if s.Category() == cat {
			out = append(out, s)
		}
	}
	return out
}
