// Package gutendex provides integration with Gutendex (Project Gutenberg metadata API).
package gutendex

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/woliveiras/bookaneer/internal/library"
)

const (
	baseURL   = "https://gutendex.com"
	userAgent = "Bookaneer/1.0 (+https://github.com/woliveiras/bookaneer)"
)

type Provider struct {
	httpClient *http.Client
}

func New() *Provider {
	return &Provider{httpClient: &http.Client{Timeout: 20 * time.Second}}
}

func (p *Provider) Name() string {
	return "gutendex"
}

type searchResponse struct {
	Results []struct {
		ID      int    `json:"id"`
		Title   string `json:"title"`
		Authors []struct {
			Name string `json:"name"`
		} `json:"authors"`
		Languages []string          `json:"languages"`
		Formats   map[string]string `json:"formats"`
	} `json:"results"`
}

func (p *Provider) Search(ctx context.Context, query string) ([]library.SearchResult, error) {
	searchURL := fmt.Sprintf("%s/books?search=%s", baseURL, url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var sr searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return nil, err
	}

	results := make([]library.SearchResult, 0, len(sr.Results))
	for _, book := range sr.Results {
		format, downloadURL := bestFormatURL(book.Formats)
		if format == "" || downloadURL == "" {
			continue
		}

		authors := make([]string, 0, len(book.Authors))
		for _, a := range book.Authors {
			if strings.TrimSpace(a.Name) != "" {
				authors = append(authors, a.Name)
			}
		}

		language := ""
		if len(book.Languages) > 0 {
			language = book.Languages[0]
		}

		results = append(results, library.SearchResult{
			ID:          fmt.Sprintf("%d", book.ID),
			Title:       book.Title,
			Authors:     authors,
			Language:    language,
			Format:      format,
			DownloadURL: downloadURL,
			InfoURL:     fmt.Sprintf("https://www.gutenberg.org/ebooks/%d", book.ID),
			Provider:    "gutendex",
		})
	}

	return results, nil
}

func bestFormatURL(formats map[string]string) (string, string) {
	if len(formats) == 0 {
		return "", ""
	}

	priority := []struct {
		mime   string
		format string
	}{
		{mime: "application/epub+zip", format: "epub"},
		{mime: "application/pdf", format: "pdf"},
		{mime: "application/x-mobipocket-ebook", format: "mobi"},
	}

	for _, p := range priority {
		for mime, u := range formats {
			if strings.HasPrefix(strings.ToLower(mime), strings.ToLower(p.mime)) && strings.TrimSpace(u) != "" {
				return p.format, u
			}
		}
	}

	return "", ""
}

func (p *Provider) GetDownloadLink(ctx context.Context, id string) (string, error) {
	lookupURL := fmt.Sprintf("%s/books/%s", baseURL, url.PathEscape(id))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, lookupURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var book struct {
		Formats map[string]string `json:"formats"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&book); err != nil {
		return "", err
	}

	_, downloadURL := bestFormatURL(book.Formats)
	if downloadURL == "" {
		return "", fmt.Errorf("no supported download format")
	}

	return downloadURL, nil
}

var _ library.Provider = (*Provider)(nil)
