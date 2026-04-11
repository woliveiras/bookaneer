// Package libgen provides integration with Library Genesis for ebook downloads.
package libgen

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/woliveiras/bookaneer/internal/library"
)

const (
	defaultSearchURL = "https://libgen.is/search.php"
	userAgent        = "Bookaneer/1.0"
)

type Provider struct {
	httpClient *http.Client
}

func New() *Provider {
	return &Provider{
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (p *Provider) Name() string {
	return "libgen"
}

type libgenEntry struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Author    string `json:"author"`
	Year      string `json:"year"`
	Publisher string `json:"publisher"`
	Language  string `json:"language"`
	Extension string `json:"extension"`
	Filesize  string `json:"filesize"`
	MD5       string `json:"md5"`
	CoverURL  string `json:"coverurl"`
	ISBN      string `json:"identifier"`
}

func (p *Provider) Search(ctx context.Context, query string) ([]library.SearchResult, error) {
	apiURL := fmt.Sprintf("https://libgen.is/json.php?fields=id,title,author,year,publisher,language,extension,filesize,md5,coverurl,identifier&limit1=25&mode=last&req=%s",
		url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return []library.SearchResult{}, nil
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return []library.SearchResult{}, nil
	}

	var entries []libgenEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return []library.SearchResult{}, nil
	}

	results := make([]library.SearchResult, 0, len(entries))
	for _, entry := range entries {
		ext := strings.ToLower(entry.Extension)
		if ext != "epub" && ext != "pdf" && ext != "mobi" && ext != "azw3" {
			continue
		}

		size, _ := strconv.ParseInt(entry.Filesize, 10, 64)
		year, _ := strconv.Atoi(entry.Year)
		authors := strings.Split(entry.Author, ",")
		for i := range authors {
			authors[i] = strings.TrimSpace(authors[i])
		}

		result := library.SearchResult{
			ID:          entry.MD5,
			Title:       entry.Title,
			Authors:     authors,
			Publisher:   entry.Publisher,
			Year:        year,
			Language:    entry.Language,
			Format:      ext,
			Size:        size,
			ISBN:        entry.ISBN,
			InfoURL:     fmt.Sprintf("https://libgen.is/book/index.php?md5=%s", entry.MD5),
			DownloadURL: fmt.Sprintf("https://library.lol/main/%s", entry.MD5),
			Provider:    "libgen",
		}

		if entry.CoverURL != "" {
			result.CoverURL = fmt.Sprintf("https://libgen.is/covers/%s", entry.CoverURL)
		}

		results = append(results, result)
	}

	return results, nil
}

func (p *Provider) GetDownloadLink(ctx context.Context, id string) (string, error) {
	return fmt.Sprintf("https://library.lol/main/%s", id), nil
}

var _ library.Provider = (*Provider)(nil)
