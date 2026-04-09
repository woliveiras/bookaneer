package library

import (
	"context"
	"log/slog"
	"regexp"
	"sort"
	"strings"
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

// sanitizeQuery removes special characters that can break search APIs.
// Keeps only alphanumeric characters, spaces, and basic punctuation.
func sanitizeQuery(query string) string {
	// Remove problematic characters for search APIs: ; ( ) ' " [ ] { }
	re := regexp.MustCompile(`[;()\[\]{}'"]+`)
	query = re.ReplaceAllString(query, " ")

	// Collapse multiple spaces
	query = regexp.MustCompile(`\s+`).ReplaceAllString(query, " ")

	// Trim and return
	return strings.TrimSpace(query)
}

// Search searches all providers in parallel and combines results.
func (a *Aggregator) Search(ctx context.Context, query string) ([]SearchResult, error) {
	if len(a.providers) == 0 {
		return []SearchResult{}, nil
	}

	// Sanitize query to remove problematic characters
	query = sanitizeQuery(query)
	if query == "" {
		return []SearchResult{}, nil
	}

	slog.Debug("library search", "query", query)

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

	// Calculate scores and sort by quality
	rankResults(allResults)

	return allResults, nil
}

// rankResults calculates quality scores and sorts results by score descending.
func rankResults(results []SearchResult) {
	for i := range results {
		results[i].Score = calculateScore(&results[i])
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
}

// calculateScore assigns a quality score to a search result.
// Higher scores indicate better quality sources.
func calculateScore(r *SearchResult) int {
	score := 0

	// Format preference: EPUB > PDF > MOBI > others
	format := strings.ToLower(r.Format)
	switch format {
	case "epub":
		score += 100
	case "pdf":
		score += 80
	case "mobi", "azw", "azw3":
		score += 60
	default:
		score += 20
	}

	// Provider reliability bonus
	provider := strings.ToLower(r.Provider)
	switch provider {
	case "internet-archive":
		score += 50 // Most reliable, legal
	case "libgen":
		score += 40
	case "annas-archive":
		score += 30
	default:
		score += 10
	}

	// Prefer newer editions (if year is available)
	if r.Year > 0 {
		if r.Year >= 2020 {
			score += 30
		} else if r.Year >= 2010 {
			score += 20
		} else if r.Year >= 2000 {
			score += 10
		}
	}

	// Has metadata bonuses
	if len(r.Authors) > 0 {
		score += 15
	}
	if r.ISBN != "" {
		score += 10
	}
	if r.CoverURL != "" {
		score += 5
	}

	// Language preference (English gets slight boost for discoverability)
	lang := strings.ToLower(r.Language)
	if lang == "eng" || lang == "en" || lang == "english" {
		score += 5
	}

	return score
}

// Providers returns the list of configured providers.
func (a *Aggregator) Providers() []string {
	names := make([]string, len(a.providers))
	for i, p := range a.providers {
		names[i] = p.Name()
	}
	return names
}
