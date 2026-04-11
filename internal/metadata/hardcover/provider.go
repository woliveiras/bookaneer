// Package hardcover implements a metadata provider for the Hardcover API.
package hardcover

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/woliveiras/bookaneer/internal/metadata"
)

const (
	baseURL      = "https://api.hardcover.app/v1/graphql"
	providerName = "hardcover"

	defaultTimeout = 30 * time.Second
)

// Provider implements metadata.Provider for Hardcover.
type Provider struct {
	client    *http.Client
	apiToken  string
	userAgent string
}

var _ metadata.Provider = (*Provider)(nil)

// New creates a new Hardcover provider.
func New(client *http.Client, apiToken string, userAgent string) *Provider {
	if client == nil {
		client = &http.Client{Timeout: defaultTimeout}
	}
	if userAgent == "" {
		userAgent = "Bookaneer/1.0"
	}
	return &Provider{
		client:    client,
		apiToken:  apiToken,
		userAgent: userAgent,
	}
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return providerName
}

// SearchAuthors searches for authors by name.
func (p *Provider) SearchAuthors(ctx context.Context, query string) ([]metadata.AuthorResult, error) {
	if p.apiToken == "" {
		return nil, metadata.ErrProviderUnavailable
	}

	gqlQuery := `query SearchAuthors($query: String!) {
		search(query: $query, query_type: "Author", per_page: 20, page: 1) {
			results
		}
	}`

	var resp searchResponse
	if err := p.doGraphQL(ctx, gqlQuery, map[string]interface{}{"query": query}, &resp); err != nil {
		return nil, err
	}

	results := make([]metadata.AuthorResult, 0, len(resp.Data.Search.Results))
	for _, r := range resp.Data.Search.Results {
		result := metadata.AuthorResult{
			ForeignID:  strconv.Itoa(r.ID),
			Name:       r.Name,
			WorksCount: r.BooksCount,
			Provider:   providerName,
		}
		if r.Image.URL != "" {
			result.PhotoURL = r.Image.URL
		}
		results = append(results, result)
	}

	return results, nil
}

// SearchBooks searches for books by title, author, or ISBN.
func (p *Provider) SearchBooks(ctx context.Context, query string) ([]metadata.BookResult, error) {
	if p.apiToken == "" {
		return nil, metadata.ErrProviderUnavailable
	}

	gqlQuery := `query SearchBooks($query: String!) {
		search(query: $query, query_type: "Book", per_page: 20, page: 1) {
			results
		}
	}`

	var resp searchResponse
	if err := p.doGraphQL(ctx, gqlQuery, map[string]interface{}{"query": query}, &resp); err != nil {
		return nil, err
	}

	results := make([]metadata.BookResult, 0, len(resp.Data.Search.Results))
	for _, r := range resp.Data.Search.Results {
		result := metadata.BookResult{
			ForeignID:     strconv.Itoa(r.ID),
			Title:         r.Title,
			Authors:       r.AuthorNames,
			PublishedYear: r.ReleaseYear,
			Provider:      providerName,
		}
		if r.Image.URL != "" {
			result.CoverURL = r.Image.URL
		}
		for _, isbn := range r.ISBNs {
			isbn = strings.ReplaceAll(isbn, "-", "")
			if len(isbn) == 13 && result.ISBN13 == "" {
				result.ISBN13 = isbn
			} else if len(isbn) == 10 && result.ISBN10 == "" {
				result.ISBN10 = isbn
			}
		}
		results = append(results, result)
	}

	return results, nil
}

// GetAuthor fetches detailed author information by Hardcover ID.
func (p *Provider) GetAuthor(ctx context.Context, foreignID string) (*metadata.Author, error) {
	if p.apiToken == "" {
		return nil, metadata.ErrProviderUnavailable
	}

	id, err := strconv.Atoi(foreignID)
	if err != nil {
		return nil, metadata.ErrNotFound
	}

	gqlQuery := `query GetAuthor($id: Int!) {
		authors(where: {id: {_eq: $id}}) {
			id
			name
			bio
			image { url }
			slug
		}
	}`

	var resp authorResponse
	if err := p.doGraphQL(ctx, gqlQuery, map[string]interface{}{"id": id}, &resp); err != nil {
		return nil, err
	}

	if len(resp.Data.Authors) == 0 {
		return nil, metadata.ErrNotFound
	}

	a := resp.Data.Authors[0]
	author := &metadata.Author{
		ForeignID: strconv.Itoa(a.ID),
		Name:      a.Name,
		Bio:       a.Bio,
		Provider:  providerName,
	}

	if a.Image.URL != "" {
		author.PhotoURL = a.Image.URL
	}

	if a.Slug != "" {
		author.Links = append(author.Links, metadata.Link{
			Type: "hardcover",
			URL:  "https://hardcover.app/authors/" + a.Slug,
		})
	}

	return author, nil
}

// GetBook fetches detailed book information by Hardcover ID.
func (p *Provider) GetBook(ctx context.Context, foreignID string) (*metadata.Book, error) {
	if p.apiToken == "" {
		return nil, metadata.ErrProviderUnavailable
	}

	id, err := strconv.Atoi(foreignID)
	if err != nil {
		return nil, metadata.ErrNotFound
	}

	gqlQuery := `query GetBook($id: Int!) {
		books(where: {id: {_eq: $id}}) {
			id
			title
			subtitle
			description
			release_date
			pages
			slug
			cached_image { url }
			contributions {
				author { id name }
			}
			book_series {
				series { name }
				position
			}
			taggings(where: {tag: {tag_type_id: {_eq: 1}}}, limit: 10) {
				tag { tag }
			}
		}
	}`

	var resp bookResponse
	if err := p.doGraphQL(ctx, gqlQuery, map[string]interface{}{"id": id}, &resp); err != nil {
		return nil, err
	}

	if len(resp.Data.Books) == 0 {
		return nil, metadata.ErrNotFound
	}

	b := resp.Data.Books[0]
	book := &metadata.Book{
		ForeignID:   strconv.Itoa(b.ID),
		Title:       b.Title,
		Subtitle:    b.Subtitle,
		Description: b.Description,
		PageCount:   b.Pages,
		Provider:    providerName,
	}

	if b.ReleaseDate != "" {
		book.PublishedDate = parseDate(b.ReleaseDate)
	}

	if b.CachedImage.URL != "" {
		book.CoverURL = b.CachedImage.URL
	}

	for _, c := range b.Contributions {
		book.Authors = append(book.Authors, c.Author.Name)
		book.AuthorIDs = append(book.AuthorIDs, strconv.Itoa(c.Author.ID))
	}

	if len(b.BookSeries) > 0 {
		book.Series = b.BookSeries[0].Series.Name
		book.SeriesPosition = b.BookSeries[0].Position
	}

	for _, t := range b.Taggings {
		book.Genres = append(book.Genres, t.Tag.Tag)
	}

	if b.Slug != "" {
		book.Links = append(book.Links, metadata.Link{
			Type: "hardcover",
			URL:  "https://hardcover.app/books/" + b.Slug,
		})
	}

	return book, nil
}

// GetBookByISBN fetches book information by ISBN.
func (p *Provider) GetBookByISBN(ctx context.Context, isbn string) (*metadata.Book, error) {
	if p.apiToken == "" {
		return nil, metadata.ErrProviderUnavailable
	}

	isbn = strings.ReplaceAll(isbn, "-", "")
	isbn = strings.ReplaceAll(isbn, " ", "")

	gqlQuery := `query GetBookByISBN($isbn: String!) {
		editions(where: {_or: [{isbn_10: {_eq: $isbn}}, {isbn_13: {_eq: $isbn}}]}, limit: 1) {
			book { id }
		}
	}`

	var resp editionResponse
	if err := p.doGraphQL(ctx, gqlQuery, map[string]interface{}{"isbn": isbn}, &resp); err != nil {
		return nil, err
	}

	if len(resp.Data.Editions) == 0 {
		return nil, metadata.ErrNotFound
	}

	return p.GetBook(ctx, strconv.Itoa(resp.Data.Editions[0].Book.ID))
}

func (p *Provider) doGraphQL(ctx context.Context, query string, variables map[string]interface{}, v interface{}) error {
	body := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiToken)
	req.Header.Set("User-Agent", p.userAgent)

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", metadata.ErrProviderUnavailable)
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK:
		// OK
	case http.StatusUnauthorized:
		return metadata.ErrProviderUnavailable
	case http.StatusTooManyRequests:
		return metadata.ErrRateLimited
	case http.StatusNotFound:
		return metadata.ErrNotFound
	default:
		return fmt.Errorf("unexpected status %d: %w", resp.StatusCode, metadata.ErrProviderUnavailable)
	}

	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

func parseDate(s string) time.Time {
	formats := []string{
		"2006-01-02",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
		"2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t
		}
	}

	return time.Time{}
}
