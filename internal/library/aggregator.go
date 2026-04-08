package library

import (
	"context"
	"log/slog"
	"sync"
)

// Aggregator searches multiple library providers in parallel.
type Aggregator struct {
	providers []Provider
}

// NewAggregator creates a new aggregator with the given providers.
func NewAggregator(providers ...Provider) *Aggregator {
	return &Aggregator{providers: providers}
}

// Search searches all providers in parallel and combines results.
func (a *Aggregator) Search(ctx context.Context, query string) ([]SearchResult, error) {
	if len(a.providers) == 0 {
		return []SearchResult{}, nil
	}

	type providerResult struct {
		name    string
		results []SearchResult
		err     error
	}

	resultsChan := make(chan providerResult, len(a.providers))
	var wg sync.WaitGroup

	for _, provider := range a.providers {
		wg.Add(1)
		go func(p Provider) {
			defer wg.Done()
			results, err := p.Search(ctx, query)
			resultsChan <- providerResult{name: p.Name(), results: results, err: err}
		}(provider)
	}

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	var allResults []SearchResult
	for pr := range resultsChan {
		if pr.err != nil {
			slog.Debug("library provider search failed", "provider", pr.name, "error", pr.err)
			continue
		}
		slog.Debug("library provider search completed", "provider", pr.name, "results", len(pr.results))
		allResults = append(allResults, pr.results...)
	}

	return allResults, nil
}

// Providers returns the list of configured providers.
func (a *Aggregator) Providers() []string {
	names := make([]string, len(a.providers))
	for i, p := range a.providers {
		names[i] = p.Name()
	}
	return names
}
