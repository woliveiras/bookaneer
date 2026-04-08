package search

import "sync"

// IndexerFactory creates an Indexer from a config.
type IndexerFactory func(cfg IndexerConfig) (Indexer, error)

var (
	factoryMu  sync.RWMutex
	factoryReg = make(map[string]IndexerFactory)
)

// RegisterFactory registers an indexer factory by type name.
func RegisterFactory(typeName string, factory IndexerFactory) {
	factoryMu.Lock()
	defer factoryMu.Unlock()
	factoryReg[typeName] = factory
}

// GetFactory returns a registered factory by type name.
func GetFactory(typeName string) (IndexerFactory, bool) {
	factoryMu.RLock()
	defer factoryMu.RUnlock()
	f, ok := factoryReg[typeName]
	return f, ok
}
