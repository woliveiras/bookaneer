package release

import "sync"

// Registry holds all registered release sources and allows lookup by name or type.
type Registry struct {
	mu      sync.RWMutex
	sources []Source
}

// NewRegistry creates an empty source registry.
func NewRegistry() *Registry {
	return &Registry{}
}

// Register adds a source to the registry.
func (r *Registry) Register(s Source) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sources = append(r.sources, s)
}

// All returns every registered source.
func (r *Registry) All() []Source {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Source, len(r.sources))
	copy(out, r.sources)
	return out
}

// ByType returns sources matching the given type.
func (r *Registry) ByType(t SourceType) []Source {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []Source
	for _, s := range r.sources {
		if s.Type() == t {
			out = append(out, s)
		}
	}
	return out
}

// ByName returns the first source with the given name, or nil.
func (r *Registry) ByName(name string) Source {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, s := range r.sources {
		if s.Name() == name {
			return s
		}
	}
	return nil
}
