package agent

import (
	"fmt"
	"sort"
	"sync"
)

// Registry holds the enabled agents and the global default.
type Registry struct {
	mu       sync.RWMutex
	agents   map[string]Agent
	fallback string // default agent name
}

// NewRegistry creates an empty registry.
func NewRegistry() *Registry {
	return &Registry{agents: make(map[string]Agent)}
}

// Register adds an agent. The first registered agent becomes the default until
// SetDefault is called.
func (r *Registry) Register(a Agent) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.agents[a.Name()] = a
	if r.fallback == "" {
		r.fallback = a.Name()
	}
}

// SetDefault sets the global default agent. It errors if the name is unknown.
func (r *Registry) SetDefault(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.agents[name]; !ok {
		return fmt.Errorf("default agent %q is not enabled", name)
	}
	r.fallback = name
	return nil
}

// Default returns the current default agent, or nil if none are registered.
func (r *Registry) Default() Agent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.agents[r.fallback]
}

// DefaultName returns the default agent's name.
func (r *Registry) DefaultName() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.fallback
}

// Get returns the named agent and whether it exists.
func (r *Registry) Get(name string) (Agent, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.agents[name]
	return a, ok
}

// Names returns the registered agent names, sorted.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.agents))
	for n := range r.agents {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// List returns the registered agents, ordered by name.
func (r *Registry) List() []Agent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.agents))
	for n := range r.agents {
		names = append(names, n)
	}
	sort.Strings(names)
	out := make([]Agent, 0, len(names))
	for _, n := range names {
		out = append(out, r.agents[n])
	}
	return out
}
