// Package annas provides integration with Anna's Archive for ebook downloads.
package annas

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/woliveiras/bookaneer/internal/library"
	"github.com/woliveiras/bookaneer/internal/library/flaresolverr"
)

// mirrors lists Anna's Archive domains in preferred order.
var mirrors = []string{
	"https://annas-archive.gl",
	"https://annas-archive.org",
	"https://annas-archive.se",
	"https://annas-archive.li",
	"https://annas-archive.gs",
}

const userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

// Provider implements library.Provider for Anna's Archive.
type Provider struct {
	httpClient *http.Client
	flare      *flaresolverr.Client // optional; nil when not configured
}

// New creates a new Anna's Archive provider.
func New() *Provider {
	return &Provider{
		httpClient: &http.Client{
			Timeout: 8 * time.Second,
		},
	}
}

// NewWithFlareSolverr creates a provider that routes through FlareSolverr.
func NewWithFlareSolverr(solverrURL string) *Provider {
	p := New()
	p.flare = flaresolverr.New(solverrURL)
	return p
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return "annas-archive"
}

// Search searches Anna's Archive for ebooks, trying all mirrors in parallel and
// returning the first non-empty result set.
func (p *Provider) Search(ctx context.Context, query string) ([]library.SearchResult, error) {
	// Cap total search time across all mirrors
	ctx, cancel := context.WithTimeout(ctx, 12*time.Second)
	defer cancel()

	type result struct {
		results []library.SearchResult
		err     error
	}

	resultCh := make(chan result, len(mirrors))
	var wg sync.WaitGroup

	for _, mirror := range mirrors {
		wg.Add(1)
		go func(m string) {
			defer wg.Done()
			res, err := p.searchMirror(ctx, m, query)
			if err != nil {
				slog.Debug("annas-archive mirror failed", "mirror", m, "error", err)
			}
			resultCh <- result{results: res, err: err}
		}(mirror)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	var best []library.SearchResult
	for r := range resultCh {
		if len(r.results) > len(best) {
			best = r.results
		}
		// Return early if we already have good results
		if len(best) >= 5 {
			cancel()
			return best, nil
		}
	}
	return best, nil
}

func (p *Provider) searchMirror(ctx context.Context, baseURL, query string) ([]library.SearchResult, error) {
	searchURL := fmt.Sprintf("%s/search?index=&q=%s&ext=epub&ext=pdf&ext=mobi&sort=&lang=",
		baseURL, url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	// Detect bot/Cloudflare challenge pages — they won't contain our search anchors
	bodyStr := string(body)
	if strings.Contains(strings.ToLower(bodyStr), "just a moment") ||
		strings.Contains(strings.ToLower(bodyStr), "verify you are human") ||
		strings.Contains(strings.ToLower(bodyStr), "robot") {
		return nil, fmt.Errorf("bot challenge page detected")
	}

	results := p.parseSearchResults(baseURL, bodyStr)
	return results, nil
}

// parseSearchResults extracts results from HTML response.
func (p *Provider) parseSearchResults(baseURL, html string) []library.SearchResult {
	var results []library.SearchResult

	entries := strings.Split(html, `href="/md5/`)

	for i, entry := range entries[1:] {
		if i >= 25 {
			break
		}

		endIdx := strings.Index(entry, `"`)
		if endIdx == -1 {
			continue
		}
		md5 := entry[:endIdx]

		titleStart := strings.Index(entry, `<h3`)
		if titleStart == -1 {
			titleStart = strings.Index(entry, `class="line-clamp`)
		}
		if titleStart == -1 {
			continue
		}

		contentStart := strings.Index(entry[titleStart:], ">")
		if contentStart == -1 {
			continue
		}
		titleStart += contentStart + 1

		contentEnd := strings.Index(entry[titleStart:], "<")
		if contentEnd == -1 {
			continue
		}

		title := strings.TrimSpace(entry[titleStart : titleStart+contentEnd])
		if title == "" {
			continue
		}

		format := "unknown"
		for _, ext := range []string{"epub", "pdf", "mobi", "azw3", "djvu"} {
			if strings.Contains(strings.ToLower(entry), ext) {
				format = ext
				break
			}
		}

		result := library.SearchResult{
			ID:          md5,
			Title:       cleanHTML(title),
			Format:      format,
			DownloadURL: fmt.Sprintf("%s/md5/%s", baseURL, md5),
			InfoURL:     fmt.Sprintf("%s/md5/%s", baseURL, md5),
			Provider:    "annas-archive",
		}

		authorIdx := strings.Index(entry, `class="italic"`)
		if authorIdx != -1 && authorIdx < 500 {
			authorStart := strings.Index(entry[authorIdx:], ">") + authorIdx + 1
			authorEnd := strings.Index(entry[authorStart:], "<")
			if authorEnd != -1 {
				author := strings.TrimSpace(entry[authorStart : authorStart+authorEnd])
				if author != "" && author != title {
					result.Authors = []string{cleanHTML(author)}
				}
			}
		}

		results = append(results, result)
	}

	return results
}

func cleanHTML(s string) string {
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&quot;", `"`)
	s = strings.ReplaceAll(s, "&#39;", "'")
	s = strings.ReplaceAll(s, "&nbsp;", " ")
	return strings.TrimSpace(s)
}

// GetDownloadLink returns a download link for the given MD5 hash.
func (p *Provider) GetDownloadLink(ctx context.Context, id string) (string, error) {
	return fmt.Sprintf("%s/md5/%s", mirrors[0], id), nil
}

var _ library.Provider = (*Provider)(nil)
