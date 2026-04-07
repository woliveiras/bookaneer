// Package openlibrary implements a metadata provider for the Open Library API.
// https://openlibrary.org/developers/api
package openlibrary

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/woliveiras/bookaneer/internal/metadata"
)

const (
	baseURL      = "https://openlibrary.org"
	coversURL    = "https://covers.openlibrary.org"
	providerName = "openlibrary"

	defaultTimeout = 30 * time.Second
)

// Provider implements metadata.Provider for Open Library.
type Provider struct {
	client    *http.Client
	userAgent string
}

var _ metadata.Provider = (*Provider)(nil)

// New creates a new Open Library provider.
func New(client *http.Client, userAgent string) *Provider {
	if client == nil {
		client = &http.Client{Timeout: defaultTimeout}
	}
	if userAgent == "" {
		userAgent = "Bookaneer/1.0"
	}
	return &Provider{
		client:    client,
		userAgent: userAgent,
	}
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return providerName
}

// SearchAuthors searches for authors by name.
func (p *Provider) SearchAuthors(ctx context.Context, query string) ([]metadata.AuthorResult, error) {
	u := fmt.Sprintf("%s/search/authors.json?q=%s&limit=20", baseURL, url.QueryEscape(query))

	var resp authorSearchResponse
	if err := p.doRequest(ctx, u, &resp); err != nil {
		return nil, err
	}

	results := make([]metadata.AuthorResult, 0, len(resp.Docs))
	for _, doc := range resp.Docs {
		result := metadata.AuthorResult{
			ForeignID:  doc.Key,
			Name:       doc.Name,
			BirthYear:  doc.BirthDate,
			DeathYear:  doc.DeathDate,
			WorksCount: doc.WorkCount,
			Provider:   providerName,
		}
		if doc.Key != "" {
			result.PhotoURL = fmt.Sprintf("%s/a/olid/%s-M.jpg", coversURL, doc.Key)
		}
		results = append(results, result)
	}

	return results, nil
}

// SearchBooks searches for books by title, author, or ISBN.
func (p *Provider) SearchBooks(ctx context.Context, query string) ([]metadata.BookResult, error) {
	fields := "key,title,author_name,author_key,first_publish_year,cover_i,isbn,editions"
	u := fmt.Sprintf("%s/search.json?q=%s&fields=%s&limit=20", baseURL, url.QueryEscape(query), fields)

	var resp bookSearchResponse
	if err := p.doRequest(ctx, u, &resp); err != nil {
		return nil, err
	}

	results := make([]metadata.BookResult, 0, len(resp.Docs))
	for _, doc := range resp.Docs {
		result := metadata.BookResult{
			ForeignID:     strings.TrimPrefix(doc.Key, "/works/"),
			Title:         doc.Title,
			Authors:       doc.AuthorName,
			PublishedYear: doc.FirstPublishYear,
			Provider:      providerName,
		}
		if doc.CoverI > 0 {
			result.CoverURL = fmt.Sprintf("%s/b/id/%d-M.jpg", coversURL, doc.CoverI)
		}
		for _, isbn := range doc.ISBN {
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

// GetAuthor fetches detailed author information by Open Library ID.
func (p *Provider) GetAuthor(ctx context.Context, foreignID string) (*metadata.Author, error) {
	foreignID = strings.TrimPrefix(foreignID, "/authors/")
	u := fmt.Sprintf("%s/authors/%s.json", baseURL, foreignID)

	var resp authorResponse
	if err := p.doRequest(ctx, u, &resp); err != nil {
		return nil, err
	}

	author := &metadata.Author{
		ForeignID: foreignID,
		Name:      resp.Name,
		Provider:  providerName,
		PhotoURL:  fmt.Sprintf("%s/a/olid/%s-L.jpg", coversURL, foreignID),
	}

	if resp.Bio != nil {
		switch bio := resp.Bio.(type) {
		case string:
			author.Bio = bio
		case map[string]interface{}:
			if val, ok := bio["value"].(string); ok {
				author.Bio = val
			}
		}
	}

	if resp.PersonalName != "" {
		author.SortName = resp.PersonalName
	}
	if resp.BirthDate != "" {
		author.BirthDate = parseDate(resp.BirthDate)
	}
	if resp.DeathDate != "" {
		author.DeathDate = parseDate(resp.DeathDate)
	}

	if resp.Wikipedia != "" {
		author.Links = append(author.Links, metadata.Link{Type: "wikipedia", URL: resp.Wikipedia})
	}
	for _, link := range resp.Links {
		author.Links = append(author.Links, metadata.Link{Type: link.Title, URL: link.URL})
	}

	return author, nil
}

// GetBook fetches detailed book information by Open Library work ID.
func (p *Provider) GetBook(ctx context.Context, foreignID string) (*metadata.Book, error) {
	foreignID = strings.TrimPrefix(foreignID, "/works/")
	u := fmt.Sprintf("%s/works/%s.json", baseURL, foreignID)

	var resp workResponse
	if err := p.doRequest(ctx, u, &resp); err != nil {
		return nil, err
	}

	book := &metadata.Book{
		ForeignID: foreignID,
		Title:     resp.Title,
		Subtitle:  resp.Subtitle,
		Subjects:  resp.Subjects,
		Provider:  providerName,
	}

	if resp.Description != nil {
		switch desc := resp.Description.(type) {
		case string:
			book.Description = desc
		case map[string]interface{}:
			if val, ok := desc["value"].(string); ok {
				book.Description = val
			}
		}
	}

	for _, authorRef := range resp.Authors {
		if authorRef.Author.Key != "" {
			authorID := strings.TrimPrefix(authorRef.Author.Key, "/authors/")
			book.AuthorIDs = append(book.AuthorIDs, authorID)
		}
	}

	for _, authorID := range book.AuthorIDs {
		author, err := p.GetAuthor(ctx, authorID)
		if err == nil {
			book.Authors = append(book.Authors, author.Name)
		}
	}

	if len(resp.Covers) > 0 {
		book.CoverURL = fmt.Sprintf("%s/b/id/%d-L.jpg", coversURL, resp.Covers[0])
	}

	if resp.FirstPublishDate != "" {
		book.PublishedDate = parseDate(resp.FirstPublishDate)
	}

	for _, link := range resp.Links {
		book.Links = append(book.Links, metadata.Link{Type: link.Title, URL: link.URL})
	}

	return book, nil
}

// GetBookByISBN fetches book information by ISBN.
func (p *Provider) GetBookByISBN(ctx context.Context, isbn string) (*metadata.Book, error) {
	isbn = strings.ReplaceAll(isbn, "-", "")
	isbn = strings.ReplaceAll(isbn, " ", "")

	if !isValidISBN(isbn) {
		return nil, metadata.ErrInvalidISBN
	}

	u := fmt.Sprintf("%s/isbn/%s.json", baseURL, isbn)

	var resp editionResponse
	if err := p.doRequest(ctx, u, &resp); err != nil {
		return nil, err
	}

	book := &metadata.Book{
		ForeignID: strings.TrimPrefix(resp.Key, "/books/"),
		Title:     resp.Title,
		Publisher: strings.Join(resp.Publishers, ", "),
		PageCount: resp.NumberOfPages,
		Provider:  providerName,
	}

	if len(resp.ISBN10) > 0 {
		book.ISBN10 = resp.ISBN10[0]
	}
	if len(resp.ISBN13) > 0 {
		book.ISBN13 = resp.ISBN13[0]
	}

	if resp.PublishDate != "" {
		book.PublishedDate = parseDate(resp.PublishDate)
	}

	if len(resp.Covers) > 0 {
		book.CoverURL = fmt.Sprintf("%s/b/id/%d-L.jpg", coversURL, resp.Covers[0])
	}

	if len(resp.Languages) > 0 {
		book.Language = strings.TrimPrefix(resp.Languages[0].Key, "/languages/")
	}

	if len(resp.Works) > 0 {
		workID := strings.TrimPrefix(resp.Works[0].Key, "/works/")
		workBook, err := p.GetBook(ctx, workID)
		if err == nil {
			book.Description = workBook.Description
			book.Subjects = workBook.Subjects
			book.Authors = workBook.Authors
			book.AuthorIDs = workBook.AuthorIDs
			if book.CoverURL == "" {
				book.CoverURL = workBook.CoverURL
			}
		}
	}

	return book, nil
}

func (p *Provider) doRequest(ctx context.Context, url string, v interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("User-Agent", p.userAgent)
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
	case http.StatusTooManyRequests:
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
	formats := []string{
		"2006-01-02",
		"January 2, 2006",
		"Jan 2, 2006",
		"2006",
		"January 2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t
		}
	}

	re := regexp.MustCompile(`\b(\d{4})\b`)
	if matches := re.FindStringSubmatch(s); len(matches) > 1 {
		if year, err := strconv.Atoi(matches[1]); err == nil {
			return time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
		}
	}

	return time.Time{}
}

func isValidISBN(isbn string) bool {
	if len(isbn) == 10 {
		return isValidISBN10(isbn)
	}
	if len(isbn) == 13 {
		return isValidISBN13(isbn)
	}
	return false
}

func isValidISBN10(isbn string) bool {
	sum := 0
	for i := 0; i < 9; i++ {
		digit, err := strconv.Atoi(string(isbn[i]))
		if err != nil {
			return false
		}
		sum += digit * (10 - i)
	}

	last := isbn[9]
	if last == 'X' || last == 'x' {
		sum += 10
	} else {
		digit, err := strconv.Atoi(string(last))
		if err != nil {
			return false
		}
		sum += digit
	}

	return sum%11 == 0
}

func isValidISBN13(isbn string) bool {
	sum := 0
	for i := 0; i < 13; i++ {
		digit, err := strconv.Atoi(string(isbn[i]))
		if err != nil {
			return false
		}
		if i%2 == 0 {
			sum += digit
		} else {
			sum += digit * 3
		}
	}
	return sum%10 == 0
}
