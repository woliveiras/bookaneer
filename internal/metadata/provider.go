// Package metadata defines interfaces and types for external metadata providers.
// Providers fetch author/book information from services like OpenLibrary, GoogleBooks, etc.
package metadata

import (
	"context"
	"time"
)

// Provider fetches metadata from an external source.
type Provider interface {
	// Name returns the provider identifier (e.g., "openlibrary", "googlebooks").
	Name() string

	// SearchAuthors searches for authors by name.
	SearchAuthors(ctx context.Context, query string) ([]AuthorResult, error)

	// SearchBooks searches for books by title, author, or ISBN.
	SearchBooks(ctx context.Context, query string) ([]BookResult, error)

	// GetAuthor fetches detailed author information by provider-specific ID.
	GetAuthor(ctx context.Context, foreignID string) (*Author, error)

	// GetBook fetches detailed book information by provider-specific ID.
	GetBook(ctx context.Context, foreignID string) (*Book, error)

	// GetBookByISBN fetches book information by ISBN (ISBN-10 or ISBN-13).
	GetBookByISBN(ctx context.Context, isbn string) (*Book, error)
}

// AuthorResult represents a search result for an author.
type AuthorResult struct {
	ForeignID  string `json:"foreignId"`
	Name       string `json:"name"`
	BirthYear  int    `json:"birthYear,omitempty"`
	DeathYear  int    `json:"deathYear,omitempty"`
	PhotoURL   string `json:"photoUrl,omitempty"`
	WorksCount int    `json:"worksCount,omitempty"`
	Provider   string `json:"provider"`
}

// BookResult represents a search result for a book.
type BookResult struct {
	ForeignID     string   `json:"foreignId"`
	Title         string   `json:"title"`
	Authors       []string `json:"authors,omitempty"`
	PublishedYear int      `json:"publishedYear,omitempty"`
	CoverURL      string   `json:"coverUrl,omitempty"`
	ISBN10        string   `json:"isbn10,omitempty"`
	ISBN13        string   `json:"isbn13,omitempty"`
	Provider      string   `json:"provider"`
}

// Author represents detailed author metadata.
type Author struct {
	ForeignID   string    `json:"foreignId"`
	Name        string    `json:"name"`
	SortName    string    `json:"sortName,omitempty"`
	Bio         string    `json:"bio,omitempty"`
	BirthDate   time.Time `json:"birthDate,omitempty"`
	DeathDate   time.Time `json:"deathDate,omitempty"`
	PhotoURL    string    `json:"photoUrl,omitempty"`
	Website     string    `json:"website,omitempty"`
	Wikipedia   string    `json:"wikipedia,omitempty"`
	Nationality string    `json:"nationality,omitempty"`
	Provider    string    `json:"provider"`
	Links       []Link    `json:"links,omitempty"`
}

// Book represents detailed book metadata.
type Book struct {
	ForeignID      string    `json:"foreignId"`
	Title          string    `json:"title"`
	Subtitle       string    `json:"subtitle,omitempty"`
	Authors        []string  `json:"authors,omitempty"`
	AuthorIDs      []string  `json:"authorIds,omitempty"`
	Description    string    `json:"description,omitempty"`
	PublishedDate  time.Time `json:"publishedDate,omitempty"`
	Publisher      string    `json:"publisher,omitempty"`
	PageCount      int       `json:"pageCount,omitempty"`
	Language       string    `json:"language,omitempty"`
	ISBN10         string    `json:"isbn10,omitempty"`
	ISBN13         string    `json:"isbn13,omitempty"`
	ASIN           string    `json:"asin,omitempty"`
	CoverURL       string    `json:"coverUrl,omitempty"`
	Genres         []string  `json:"genres,omitempty"`
	Subjects       []string  `json:"subjects,omitempty"`
	Series         string    `json:"series,omitempty"`
	SeriesPosition float64   `json:"seriesPosition,omitempty"`
	AverageRating  float64   `json:"averageRating,omitempty"`
	RatingsCount   int       `json:"ratingsCount,omitempty"`
	Provider       string    `json:"provider"`
	Links          []Link    `json:"links,omitempty"`
	Editions       []Edition `json:"editions,omitempty"`
}

// Edition represents a specific edition of a book.
type Edition struct {
	ForeignID     string    `json:"foreignId"`
	Title         string    `json:"title"`
	Publisher     string    `json:"publisher,omitempty"`
	PublishedDate time.Time `json:"publishedDate,omitempty"`
	Format        string    `json:"format,omitempty"`
	PageCount     int       `json:"pageCount,omitempty"`
	Language      string    `json:"language,omitempty"`
	ISBN10        string    `json:"isbn10,omitempty"`
	ISBN13        string    `json:"isbn13,omitempty"`
	CoverURL      string    `json:"coverUrl,omitempty"`
}

// Link represents an external reference link.
type Link struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}
