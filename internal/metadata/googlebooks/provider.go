// Package googlebooks implements a metadata provider for the Google Books API.
package googlebooks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/woliveiras/bookaneer/internal/metadata"
)

const (
	baseURL      = "https://www.googleapis.com/books/v1"
	providerName = "googlebooks"

	defaultTimeout = 30 * time.Second
)

// Provider implements metadata.Provider for Google Books.
type Provider struct {
	client *http.Client
	apiKey string
}

var _ metadata.Provider = (*Provider)(nil)

// New creates a new Google Books provider.
func New(client *http.Client, apiKey string) *Provider {
	if client == nil {
		client = &http.Client{Timeout: defaultTimeout}
	}
	return &Provider{
		client: client,
		apiKey: apiKey,
	}
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return providerName
}

// SearchAuthors searches for authors by name.
func (p *Provider) SearchAuthors(ctx context.Context, query string) ([]metadata.AuthorResult, error) {
	u := p.buildURL("/volumes", map[string]string{
		"q":          "inauthor:" + query,
		"maxResults": "40",
		"printType":  "books",
	})

	var resp volumesResponse
	if err := p.doRequest(ctx, u, &resp); err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	results := make([]metadata.AuthorResult, 0)

	for _, item := range resp.Items {
		for _, author := range item.VolumeInfo.Authors {
			authorLower := strings.ToLower(author)
			if seen[authorLower] {
				continue
			}
			seen[authorLower] = true

			results = append(results, metadata.AuthorResult{
				ForeignID: url.QueryEscape(author),
				Name:      author,
				Provider:  providerName,
			})
		}
	}

	return results, nil
}

// SearchBooks searches for books by title, author, or ISBN.
func (p *Provider) SearchBooks(ctx context.Context, query string) ([]metadata.BookResult, error) {
	u := p.buildURL("/volumes", map[string]string{
		"q":          query,
		"maxResults": "20",
		"printType":  "books",
	})

	var resp volumesResponse
	if err := p.doRequest(ctx, u, &resp); err != nil {
		return nil, err
	}

	results := make([]metadata.BookResult, 0, len(resp.Items))
	for _, item := range resp.Items {
		result := metadata.BookResult{
			ForeignID:     item.ID,
			Title:         item.VolumeInfo.Title,
			Authors:       item.VolumeInfo.Authors,
			PublishedYear: parseYear(item.VolumeInfo.PublishedDate),
			Provider:      providerName,
		}

		if item.VolumeInfo.ImageLinks.Thumbnail != "" {
			result.CoverURL = item.VolumeInfo.ImageLinks.Thumbnail
		}

		for _, id := range item.VolumeInfo.IndustryIdentifiers {
			switch id.Type {
			case "ISBN_10":
				result.ISBN10 = id.Identifier
			case "ISBN_13":
				result.ISBN13 = id.Identifier
			}
		}

		results = append(results, result)
	}

	return results, nil
}

// GetAuthor fetches detailed author information.
func (p *Provider) GetAuthor(ctx context.Context, foreignID string) (*metadata.Author, error) {
	authorName, err := url.QueryUnescape(foreignID)
	if err != nil {
		authorName = foreignID
	}

	u := p.buildURL("/volumes", map[string]string{
		"q":          "inauthor:\"" + authorName + "\"",
		"maxResults": "1",
		"printType":  "books",
	})

	var resp volumesResponse
	if err := p.doRequest(ctx, u, &resp); err != nil {
		return nil, err
	}

	if len(resp.Items) == 0 {
		return nil, metadata.ErrNotFound
	}

	return &metadata.Author{
		ForeignID: foreignID,
		Name:      authorName,
		Provider:  providerName,
	}, nil
}

// GetBook fetches detailed book information by Google Books volume ID.
func (p *Provider) GetBook(ctx context.Context, foreignID string) (*metadata.Book, error) {
	u := p.buildURL("/volumes/"+foreignID, nil)

	var item volumeItem
	if err := p.doRequest(ctx, u, &item); err != nil {
		return nil, err
	}

	return p.volumeToBook(&item), nil
}

// GetBookByISBN fetches book information by ISBN.
func (p *Provider) GetBookByISBN(ctx context.Context, isbn string) (*metadata.Book, error) {
	isbn = strings.ReplaceAll(isbn, "-", "")
	isbn = strings.ReplaceAll(isbn, " ", "")

	u := p.buildURL("/volumes", map[string]string{
		"q":          "isbn:" + isbn,
		"maxResults": "1",
	})

	var resp volumesResponse
	if err := p.doRequest(ctx, u, &resp); err != nil {
		return nil, err
	}

	if resp.TotalItems == 0 || len(resp.Items) == 0 {
		return nil, metadata.ErrNotFound
	}

	return p.GetBook(ctx, resp.Items[0].ID)
}

func (p *Provider) volumeToBook(item *volumeItem) *metadata.Book {
	book := &metadata.Book{
		ForeignID:     item.ID,
		Title:         item.VolumeInfo.Title,
		Subtitle:      item.VolumeInfo.Subtitle,
		Authors:       item.VolumeInfo.Authors,
		Description:   item.VolumeInfo.Description,
		Publisher:     item.VolumeInfo.Publisher,
		PageCount:     item.VolumeInfo.PageCount,
		Language:      item.VolumeInfo.Language,
		Genres:        item.VolumeInfo.Categories,
		AverageRating: item.VolumeInfo.AverageRating,
		RatingsCount:  item.VolumeInfo.RatingsCount,
		Provider:      providerName,
	}

	if item.VolumeInfo.PublishedDate != "" {
		book.PublishedDate = parseDate(item.VolumeInfo.PublishedDate)
	}

	for _, id := range item.VolumeInfo.IndustryIdentifiers {
		switch id.Type {
		case "ISBN_10":
			book.ISBN10 = id.Identifier
		case "ISBN_13":
			book.ISBN13 = id.Identifier
		}
	}

	links := item.VolumeInfo.ImageLinks
	switch {
	case links.ExtraLarge != "":
		book.CoverURL = links.ExtraLarge
	case links.Large != "":
		book.CoverURL = links.Large
	case links.Medium != "":
		book.CoverURL = links.Medium
	case links.Thumbnail != "":
		book.CoverURL = links.Thumbnail
	case links.SmallThumbnail != "":
		book.CoverURL = links.SmallThumbnail
	}

	if item.VolumeInfo.InfoLink != "" {
		book.Links = append(book.Links, metadata.Link{
			Type: "googlebooks",
			URL:  item.VolumeInfo.InfoLink,
		})
	}

	return book
}

func (p *Provider) buildURL(path string, params map[string]string) string {
	u := baseURL + path
	if params == nil {
		params = make(map[string]string)
	}
	if p.apiKey != "" {
		params["key"] = p.apiKey
	}

	if len(params) > 0 {
		values := url.Values{}
		for k, v := range params {
			values.Set(k, v)
		}
		u += "?" + values.Encode()
	}

	return u
}

func (p *Provider) doRequest(ctx context.Context, url string, v interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", metadata.ErrProviderUnavailable)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		// OK
	case http.StatusNotFound:
		return metadata.ErrNotFound
	case http.StatusTooManyRequests, http.StatusForbidden:
		return metadata.ErrRateLimited
	default:
		return fmt.Errorf("unexpected status %d: %w", resp.StatusCode, metadata.ErrProviderUnavailable)
	}

	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

func parseDate(s string) time.Time {
	formats := []string{"2006-01-02", "2006-01", "2006"}
	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

func parseYear(s string) int {
	t := parseDate(s)
	if t.IsZero() {
		return 0
	}
	return t.Year()
}
