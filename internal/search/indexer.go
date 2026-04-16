package search

import (
	"context"
	"time"
)

// Indexer defines the interface for search indexers (Newznab, Torznab).
type Indexer interface {
	Name() string
	Type() string
	Search(ctx context.Context, query SearchQuery) ([]Result, error)
	Caps(ctx context.Context) (*Capabilities, error)
	Test(ctx context.Context) error
}

// SearchQuery represents a search request.
type SearchQuery struct {
	Query    string
	Author   string
	Title    string
	ISBN     string
	Category []string
	Limit    int
	Offset   int
}

// Result represents a single search result from an indexer.
type Result struct {
	GUID        string    `json:"guid"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Size        int64     `json:"size"`
	PubDate     time.Time `json:"pubDate"`
	Category    string    `json:"category,omitempty"`
	CategoryID  string    `json:"categoryId,omitempty"`
	DownloadURL string    `json:"downloadUrl"`
	InfoURL     string    `json:"infoUrl,omitempty"`
	Comments    int       `json:"comments,omitempty"`
	Seeders     int       `json:"seeders,omitempty"`
	Leechers    int       `json:"leechers,omitempty"`
	Grabs       int       `json:"grabs,omitempty"`
	Quality     string    `json:"quality,omitempty"`
	QualityRank int       `json:"qualityRank,omitempty"`
	IndexerID   int64     `json:"indexerId"`
	IndexerName string    `json:"indexerName"`
}

// Capabilities describes what an indexer can do.
type Capabilities struct {
	Searching struct {
		Search     bool `json:"search"`
		BookSearch bool `json:"bookSearch"`
	} `json:"searching"`
	Categories []Category `json:"categories"`
}

// Category represents a category supported by an indexer.
type Category struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	SubCategory []Category `json:"subCategory,omitempty"`
}

// IndexerConfig holds the configuration for an indexer.
type IndexerConfig struct {
	ID                      int64    `json:"id"`
	Name                    string   `json:"name"`
	Type                    string   `json:"type"`
	BaseURL                 string   `json:"baseUrl"`
	APIPath                 string   `json:"apiPath"`
	APIKey                  string   `json:"apiKey"`
	Categories              string   `json:"categories"`
	Priority                int      `json:"priority"`
	Enabled                 bool     `json:"enabled"`
	EnableRSS               bool     `json:"enableRss"`
	EnableInteractiveSearch bool     `json:"enableInteractiveSearch"`
	AdditionalParameters    string   `json:"additionalParameters"`
	MinimumSeeders          int      `json:"minimumSeeders"`      // Torznab only
	SeedRatio               *float64 `json:"seedRatio,omitempty"` // Torznab only, nil = use client default
	SeedTime                *int     `json:"seedTime,omitempty"`  // Torznab only, minutes, nil = use client default
	CreatedAt               string   `json:"createdAt"`
	UpdatedAt               string   `json:"updatedAt"`
}

// IndexerOptions holds global indexer settings.
type IndexerOptions struct {
	MinimumAge         int    `json:"minimumAge"`         // Minutes (Usenet: min age before grab)
	Retention          int    `json:"retention"`          // Days (Usenet: 0 = unlimited)
	MaximumSize        int    `json:"maximumSize"`        // MB (0 = unlimited)
	RSSSyncInterval    int    `json:"rssSyncInterval"`    // Minutes (0 = disabled)
	PreferIndexerFlags bool   `json:"preferIndexerFlags"` // Prioritize releases with special flags
	AvailabilityDelay  int    `json:"availabilityDelay"`  // Days
	UpdatedAt          string `json:"updatedAt"`
}
