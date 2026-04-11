// Package wikisource provides integration with Wikisource public domain texts.
package wikisource

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
	baseURL   = "https://en.wikisource.org"
	exportURL = "https://ws-export.wmcloud.org/"
	userAgent = "Bookaneer/1.0 (+https://github.com/woliveiras/bookaneer)"
)

var sites = []struct {
	lang string
	api  string
}{
	{lang: "en", api: "https://en.wikisource.org/w/api.php"},
	{lang: "pt", api: "https://pt.wikisource.org/w/api.php"},
	{lang: "es", api: "https://es.wikisource.org/w/api.php"},
	{lang: "fr", api: "https://fr.wikisource.org/w/api.php"},
}

type Provider struct {
	httpClient *http.Client
}

func New() *Provider {
	return &Provider{httpClient: &http.Client{Timeout: 20 * time.Second}}
}

func (p *Provider) Name() string {
	return "wikisource"
}

func (p *Provider) Search(ctx context.Context, query string) ([]library.SearchResult, error) {
	q := strings.TrimSpace(query)
	if q == "" {
		return []library.SearchResult{}, nil
	}

	results := make([]library.SearchResult, 0, 40)
	seen := map[string]struct{}{}

	for _, site := range sites {
		siteResults, err := p.searchSite(ctx, site.api, site.lang, q)
		if err != nil {
			continue
		}
		for _, r := range siteResults {
			if _, ok := seen[r.ID]; ok {
				continue
			}
			seen[r.ID] = struct{}{}
			results = append(results, r)
		}
	}

	return results, nil
}

func (p *Provider) searchSite(ctx context.Context, apiURL, lang, query string) ([]library.SearchResult, error) {
	params := url.Values{}
	params.Set("action", "opensearch")
	params.Set("search", query)
	params.Set("limit", "20")
	params.Set("namespace", "0")
	params.Set("format", "json")

	endpoint := fmt.Sprintf("%s?%s", apiURL, params.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
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

	var payload []json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if len(payload) < 4 {
		return []library.SearchResult{}, nil
	}

	var titles []string
	if err := json.Unmarshal(payload[1], &titles); err != nil {
		return nil, err
	}

	var links []string
	if err := json.Unmarshal(payload[3], &links); err != nil {
		return nil, err
	}

	count := min(len(titles), len(links))
	results := make([]library.SearchResult, 0, count)
	for i := 0; i < count; i++ {
		title := strings.TrimSpace(titles[i])
		infoURL := strings.TrimSpace(links[i])
		if title == "" || infoURL == "" {
			continue
		}

		results = append(results, library.SearchResult{
			ID:          fmt.Sprintf("%s:%s", lang, title),
			Title:       title,
			Language:    lang,
			Format:      "epub",
			DownloadURL: p.epubExportURL(lang, title),
			InfoURL:     infoURL,
			Provider:    "wikisource",
		})
	}

	return results, nil
}

func (p *Provider) GetDownloadLink(_ context.Context, id string) (string, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return "", fmt.Errorf("empty id")
	}

	parts := strings.SplitN(id, ":", 2)
	if len(parts) == 2 && strings.TrimSpace(parts[0]) != "" && strings.TrimSpace(parts[1]) != "" {
		return p.epubExportURL(parts[0], parts[1]), nil
	}

	return fmt.Sprintf("%s/w/index.php?title=%s&action=raw", baseURL, url.QueryEscape(id)), nil
}

func (p *Provider) epubExportURL(lang, title string) string {
	params := url.Values{}
	params.Set("format", "epub")
	params.Set("lang", strings.TrimSpace(lang))
	params.Set("page", strings.TrimSpace(title))
	return exportURL + "?" + params.Encode()
}

var _ library.Provider = (*Provider)(nil)
