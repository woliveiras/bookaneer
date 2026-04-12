package metadata

import (
	"context"
	"log/slog"
	"strings"
)

// Aggregator wraps multiple providers and merges results from all of them.
// Providers that fail are skipped; results are deduplicated across providers.
type Aggregator struct {
	providers []Provider
	logger    *slog.Logger
}

// NewAggregator creates an aggregator with the given providers.
// Providers are tried in the order given (first = highest priority).
func NewAggregator(logger *slog.Logger, providers ...Provider) *Aggregator {
	if logger == nil {
		logger = slog.Default()
	}
	return &Aggregator{
		providers: providers,
		logger:    logger,
	}
}

// Providers returns the list of configured providers.
func (a *Aggregator) Providers() []Provider {
	return a.providers
}

// SearchAuthors searches all providers and merges results.
// Authors are deduplicated by normalized name.
func (a *Aggregator) SearchAuthors(ctx context.Context, query string) ([]AuthorResult, error) {
	if len(a.providers) == 0 {
		return nil, ErrNoProviders
	}

	var lastErr error
	var anySucceeded bool
	var allResults []AuthorResult
	seen := make(map[string]bool)

	for _, p := range a.providers {
		results, err := p.SearchAuthors(ctx, query)
		if err != nil {
			a.logger.Warn("provider search authors failed",
				"provider", p.Name(),
				"query", query,
				"error", err,
			)
			lastErr = err
			continue
		}
		anySucceeded = true
		for _, r := range results {
			key := strings.ToLower(r.Name)
			if seen[key] {
				continue
			}
			seen[key] = true
			allResults = append(allResults, r)
		}
	}

	if !anySucceeded && lastErr != nil {
		return nil, lastErr
	}
	if allResults == nil {
		allResults = []AuthorResult{}
	}
	return allResults, nil
}

// SearchBooks searches all providers and merges results.
// Books are deduplicated by ISBN13, ISBN10, or provider+foreignID.
func (a *Aggregator) SearchBooks(ctx context.Context, query string) ([]BookResult, error) {
	if len(a.providers) == 0 {
		return nil, ErrNoProviders
	}

	var lastErr error
	var anySucceeded bool
	var allResults []BookResult
	seen := make(map[string]bool)

	for _, p := range a.providers {
		results, err := p.SearchBooks(ctx, query)
		if err != nil {
			a.logger.Warn("provider search books failed",
				"provider", p.Name(),
				"query", query,
				"error", err,
			)
			lastErr = err
			continue
		}
		anySucceeded = true
		for _, r := range results {
			key := bookDedupeKey(r)
			if seen[key] {
				continue
			}
			seen[key] = true
			allResults = append(allResults, r)
		}
	}

	if !anySucceeded && lastErr != nil {
		return nil, lastErr
	}
	if allResults == nil {
		allResults = []BookResult{}
	}
	return allResults, nil
}

// bookDedupeKey returns a deduplication key for a book result.
// Prefers ISBN13, then ISBN10, then provider+foreignID.
func bookDedupeKey(r BookResult) string {
	if r.ISBN13 != "" {
		return "isbn13:" + r.ISBN13
	}
	if r.ISBN10 != "" {
		return "isbn10:" + r.ISBN10
	}
	return r.Provider + ":" + r.ForeignID
}

// GetAuthor fetches author details, trying each provider until one succeeds.
func (a *Aggregator) GetAuthor(ctx context.Context, provider, foreignID string) (*Author, error) {
	if len(a.providers) == 0 {
		return nil, ErrNoProviders
	}

	if provider != "" {
		p, ok := a.findProvider(provider)
		if !ok {
			return nil, ErrNotFound
		}
		return p.GetAuthor(ctx, foreignID)
	}

	var lastErr error
	for _, p := range a.providers {
		author, err := p.GetAuthor(ctx, foreignID)
		if err != nil {
			a.logger.Debug("provider get author failed",
				"provider", p.Name(),
				"foreignId", foreignID,
				"error", err,
			)
			lastErr = err
			continue
		}
		return author, nil
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, ErrNotFound
}

// GetBook fetches book details, trying each provider until one succeeds.
func (a *Aggregator) GetBook(ctx context.Context, provider, foreignID string) (*Book, error) {
	if len(a.providers) == 0 {
		return nil, ErrNoProviders
	}

	if provider != "" {
		p, ok := a.findProvider(provider)
		if !ok {
			return nil, ErrNotFound
		}
		return p.GetBook(ctx, foreignID)
	}

	var lastErr error
	for _, p := range a.providers {
		book, err := p.GetBook(ctx, foreignID)
		if err != nil {
			a.logger.Debug("provider get book failed",
				"provider", p.Name(),
				"foreignId", foreignID,
				"error", err,
			)
			lastErr = err
			continue
		}
		return book, nil
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, ErrNotFound
}

// findProvider returns the provider with the given name, or false if not found.
func (a *Aggregator) findProvider(name string) (Provider, bool) {
	for _, p := range a.providers {
		if p.Name() == name {
			return p, true
		}
	}
	return nil, false
}

// GetBookByISBN fetches book details by ISBN, trying each provider.
func (a *Aggregator) GetBookByISBN(ctx context.Context, isbn string) (*Book, error) {
	if len(a.providers) == 0 {
		return nil, ErrNoProviders
	}

	var lastErr error
	for _, p := range a.providers {
		book, err := p.GetBookByISBN(ctx, isbn)
		if err != nil {
			a.logger.Debug("provider get book by ISBN failed",
				"provider", p.Name(),
				"isbn", isbn,
				"error", err,
			)
			lastErr = err
			continue
		}
		return book, nil
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, ErrNotFound
}
