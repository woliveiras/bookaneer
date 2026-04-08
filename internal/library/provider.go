// Package library provides interfaces for searching digital library sources.
package library

import "context"

// Provider searches a digital library for downloadable ebooks.
type Provider interface {
	Name() string
	Search(ctx context.Context, query string) ([]SearchResult, error)
	GetDownloadLink(ctx context.Context, id string) (string, error)
}

// SearchResult represents a downloadable ebook from a library.
type SearchResult struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Authors     []string `json:"authors,omitempty"`
	Publisher   string   `json:"publisher,omitempty"`
	Year        int      `json:"year,omitempty"`
	Language    string   `json:"language,omitempty"`
	Format      string   `json:"format"`
	Size        int64    `json:"size"`
	ISBN        string   `json:"isbn,omitempty"`
	CoverURL    string   `json:"coverUrl,omitempty"`
	DownloadURL string   `json:"downloadUrl,omitempty"`
	InfoURL     string   `json:"infoUrl,omitempty"`
	Provider    string   `json:"provider"`
	Score       int      `json:"score,omitempty"` // Quality score for ranking
}
