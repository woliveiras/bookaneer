package download

import (
	"fmt"
	"sync"
)

// ClientFactory is a function that creates a Client from configuration.
type ClientFactory func(cfg ClientConfig) (Client, error)

var (
	factoryMu sync.RWMutex
	factories = make(map[string]ClientFactory)
)

// RegisterFactory registers a client factory for a given type.
func RegisterFactory(clientType string, factory ClientFactory) {
	factoryMu.Lock()
	defer factoryMu.Unlock()
	factories[clientType] = factory
}

// NewClient creates a download client from the given configuration.
func NewClient(cfg ClientConfig) (Client, error) {
	factoryMu.RLock()
	factory, ok := factories[cfg.Type]
	factoryMu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrInvalidType, cfg.Type)
	}

	return factory(cfg)
}
