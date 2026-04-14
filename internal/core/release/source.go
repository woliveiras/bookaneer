// Package release defines the unified ReleaseSource abstraction that maps
// both digital-library providers and torrent/usenet indexers into a single
// interface.
package release

import "context"

// SourceType distinguishes between release source backends.
type SourceType string

const (
	SourceLibrary SourceType = "library"
	SourceIndexer SourceType = "indexer"
)

// Source is the unified interface for any release source.
// It abstracts library.Provider and search.Indexer into one search contract.
type Source interface {
	// Name returns a human-readable name for this source.
	Name() string
	// Type returns whether this is a library or indexer source.
	Type() SourceType
	// Search returns releases matching the query.
	Search(ctx context.Context, query Query) ([]Release, error)
}

// Query is a provider-agnostic search request.
type Query struct {
	Text   string // free-text (title, author, or combined)
	Author string
	Title  string
	ISBN   string
	Limit  int
}

// Release is the normalised result shared across all source types.
type Release struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Authors     []string   `json:"authors,omitempty"`
	Format      string     `json:"format,omitempty"`
	Size        int64      `json:"size"`
	DownloadURL string     `json:"downloadUrl"`
	InfoURL     string     `json:"infoUrl,omitempty"`
	Provider    string     `json:"provider"`
	SourceType  SourceType `json:"sourceType"`

	// Library-specific
	Language string `json:"language,omitempty"`
	Year     int    `json:"year,omitempty"`
	ISBN     string `json:"isbn,omitempty"`
	CoverURL string `json:"coverUrl,omitempty"`
	Score    int    `json:"score,omitempty"`

	// Indexer-specific
	Seeders     int    `json:"seeders,omitempty"`
	Leechers    int    `json:"leechers,omitempty"`
	Grabs       int    `json:"grabs,omitempty"`
	IndexerID   int64  `json:"indexerId,omitempty"`
	IndexerName string `json:"indexerName,omitempty"`
	Quality     string `json:"quality,omitempty"`
}
