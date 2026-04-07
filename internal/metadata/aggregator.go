package metadata

import (
	"context"
	"errors"
	"log/slog"
)

// Aggregator wraps multiple providers and implements fallback logic.
// It tries each provider in order until one succeeds.
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

// SearchAuthors searches all providers and aggregates results.
// Returns results from the first provider that succeeds.
func (a *Aggregator) SearchAuthors(ctx context.Context, query string) ([]AuthorResult, error) {
	if len(a.providers) == 0 {
		return nil, ErrNoProviders
	}

	var lastErr error
	for _, p := range a.providers {
		results, err := p.SearchAuthors(ctx, query)
		if err != nil {
			a.logger.Debug("provider search authors failed",
				"provider", p.Name(),
				"query", query,
				"error", err,
			)
			lastErr = err
			if errors.Is(err, ErrRateLimited) || errors.Is(err, ErrProviderUnavailable) {
				continue
			}
			continue
		}
		if len(results) > 0 {
			return results, nil
		}
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return []AuthorResult{}, nil
}

// SearchBooks searches all providers and aggregates results.
// Returns results from the first provider that succeeds.
func (a *Aggregator) SearchBooks(ctx context.Context, query string) ([]BookResult, error) {
	if len(a.providers) == 0 {
		return nil, ErrNoProviders
	}

	var lastErr error
	for _, p := range a.providers {
		results, err := p.SearchBooks(ctx, query)
		if err != nil {
			a.logger.Debug("provider search books failed",
				"provider", p.Name(),
				"query", query,
				"error", err,
			)
			lastErr = err
			continue
		}
		if len(results) > 0 {
			return results, nil
		}
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return []BookResult{}, nil
}

// GetAuthor fetches author details, trying each provider until one succeeds.
func (a *Aggregator) GetAuthor(ctx context.Context, provider, foreignID string) (*Author, error) {
	if len(a.providers) == 0 {
		return nil, ErrNoProviders
	}

	// If a specific provider is requested, use only that one
	if provider != "" {
		for _, p := range a.providers {
			if p.Name() == provider {
				return p.GetAuthor(ctx, foreignID)
			}
		}
		return nil, ErrNotFound
	}

	// Otherwise, try all providers
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
			if errors.Is(err, ErrNotFound) {
				continue
			}
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

	// If a specific provider is requested, use only that one
	if provider != "" {
		for _, p := range a.providers {
			if p.Name() == provider {
				return p.GetBook(ctx, foreignID)
			}
		}
		return nil, ErrNotFound
	}

	// Otherwise, try all providers
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
