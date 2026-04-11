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
			Identifier string          `json:"identifier"`
			Title      string          `json:"title"`
			Creator    json.RawMessage `json:"creator"` // can be string or []string
			Date       string          `json:"date"`
			Language   json.RawMessage `json:"language"` // can be string or []string
			Format     json.RawMessage `json:"format"`   // can be string or []string
		} `json:"docs"`
	} `json:"response"`
}

// parseStringOrSlice parses a JSON field that can be either a string or []string.
func parseStringOrSlice(raw json.RawMessage) []string {
	if len(raw) == 0 {
		return nil
	}
	// Try as []string first
	var arr []string
	if err := json.Unmarshal(raw, &arr); err == nil {
		return arr
	}
	// Try as single string
	var s string
	if err := json.Unmarshal(raw, &s); err == nil && s != "" {
		return []string{s}
	}
	return nil
}

func (p *Provider) Search(ctx context.Context, query string) ([]library.SearchResult, error) {
	// Search by title for better results
	searchQuery := fmt.Sprintf("title:(%s) AND mediatype:texts", query)
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
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var searchResp searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, err
	}

	results := make([]library.SearchResult, 0, len(searchResp.Response.Docs))
	for _, doc := range searchResp.Response.Docs {
		formats := parseStringOrSlice(doc.Format)
		format := p.bestFormat(formats)
		if format == "" {
			continue
		}

		year := 0
		if len(doc.Date) >= 4 {
			_, _ = fmt.Sscanf(doc.Date[:4], "%d", &year)
		}

		languages := parseStringOrSlice(doc.Language)
		lang := ""
		if len(languages) > 0 {
			lang = languages[0]
		}

		authors := parseStringOrSlice(doc.Creator)

		// Get the actual file name from metadata
		downloadURL := p.getDownloadURL(doc.Identifier, format)

		result := library.SearchResult{
			ID:          doc.Identifier,
			Title:       doc.Title,
			Authors:     authors,
			Year:        year,
			Language:    lang,
			Format:      format,
			InfoURL:     fmt.Sprintf("%s/details/%s", p.baseURL, doc.Identifier),
			DownloadURL: downloadURL,
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

// getDownloadURL gets the actual download URL by fetching item metadata to find the real filename.
func (p *Provider) getDownloadURL(identifier, format string) string {
	// Fetch metadata to find the actual file name
	metadataURL := fmt.Sprintf("%s/metadata/%s/files", p.baseURL, identifier)

	req, err := http.NewRequest(http.MethodGet, metadataURL, nil)
	if err != nil {
		// Fallback to default naming
		return fmt.Sprintf("%s/download/%s/%s.%s", p.baseURL, identifier, identifier, format)
	}
	req.Header.Set("User-Agent", userAgent)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Sprintf("%s/download/%s/%s.%s", p.baseURL, identifier, identifier, format)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("%s/download/%s/%s.%s", p.baseURL, identifier, identifier, format)
	}

	var filesResp struct {
		Result []struct {
			Name   string `json:"name"`
			Format string `json:"format"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&filesResp); err != nil {
		return fmt.Sprintf("%s/download/%s/%s.%s", p.baseURL, identifier, identifier, format)
	}

	// Find the file matching our desired format
	formatLower := strings.ToLower(format)
	for _, f := range filesResp.Result {
		nameLower := strings.ToLower(f.Name)
		// Check if file ends with .epub, .pdf, or .mobi
		if strings.HasSuffix(nameLower, "."+formatLower) {
			return fmt.Sprintf("%s/download/%s/%s", p.baseURL, identifier, url.PathEscape(f.Name))
		}
	}

	// Fallback to default
	return fmt.Sprintf("%s/download/%s/%s.%s", p.baseURL, identifier, identifier, format)
}

func (p *Provider) GetDownloadLink(ctx context.Context, id string) (string, error) {
	return fmt.Sprintf("%s/details/%s", p.baseURL, id), nil
}

var _ library.Provider = (*Provider)(nil)
