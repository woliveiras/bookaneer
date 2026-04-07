// Package archive provides integration with Internet Archive for ebook downloads.
package archive

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
	defaultBaseURL = "https://archive.org"
	userAgent      = "Bookaneer/1.0"
)

type Provider struct {
	baseURL    string
	httpClient *http.Client
}

func New() *Provider {
	return &Provider{
		baseURL:    defaultBaseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (p *Provider) Name() string {
	return "internet-archive"
}

type searchResponse struct {
	Response struct {
		NumFound int `json:"numFound"`
		Docs     []struct {
			Identifier  string   `json:"identifier"`
			Title       string   `json:"title"`
			Creator     []string `json:"creator"`
			Date        string   `json:"date"`
			Language    []string `json:"language"`
			Format      []string `json:"format"`
		} `json:"docs"`
	} `json:"response"`
}

func (p *Provider) Search(ctx context.Context, query string) ([]library.SearchResult, error) {
	searchQuery := fmt.Sprintf("(%s) AND mediatype:texts AND format:(PDF OR EPUB)", query)
	searchURL := fmt.Sprintf("%s/advancedsearch.php?q=%s&fl[]=identifier&fl[]=title&fl[]=creator&fl[]=date&fl[]=language&fl[]=format&sort[]=downloads+desc&rows=25&page=1&output=json",
		p.baseURL, url.QueryEscape(searchQuery))

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

	var searchResp searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, err
	}

	results := make([]library.SearchResult, 0, len(searchResp.Response.Docs))
	for _, doc := range searchResp.Response.Docs {
		format := p.bestFormat(doc.Format)
		if format == "" {
			continue
		}

		year := 0
		if len(doc.Date) >= 4 {
			fmt.Sscanf(doc.Date[:4], "%d", &year)
		}

		lang := ""
		if len(doc.Language) > 0 {
			lang = doc.Language[0]
		}

		result := library.SearchResult{
			ID:          doc.Identifier,
			Title:       doc.Title,
			Authors:     doc.Creator,
			Year:        year,
			Language:    lang,
			Format:      format,
			InfoURL:     fmt.Sprintf("%s/details/%s", p.baseURL, doc.Identifier),
			DownloadURL: fmt.Sprintf("%s/download/%s/%s.%s", p.baseURL, doc.Identifier, doc.Identifier, format),
			CoverURL:    fmt.Sprintf("%s/services/img/%s", p.baseURL, doc.Identifier),
			Provider:    "internet-archive",
		}

		results = append(results, result)
	}

	return results, nil
}

func (p *Provider) bestFormat(formats []string) string {
	priorities := map[string]int{"epub": 1, "pdf": 2, "mobi": 3}
	best := ""
	bestPriority := 999

	for _, f := range formats {
		lower := strings.ToLower(f)
		for format, priority := range priorities {
			if strings.Contains(lower, format) && priority < bestPriority {
				best = format
				bestPriority = priority
			}
		}
	}

	return best
}

func (p *Provider) GetDownloadLink(ctx context.Context, id string) (string, error) {
	return fmt.Sprintf("%s/details/%s", p.baseURL, id), nil
}

var _ library.Provider = (*Provider)(nil)
