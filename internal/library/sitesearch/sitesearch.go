// Package sitesearch provides a generic provider that searches a specific site
// using DuckDuckGo's HTML endpoint and returns normalized results.
package sitesearch

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/woliveiras/bookaneer/internal/library"
)

const (
	duckURL   = "https://duckduckgo.com/html/"
	userAgent = "Bookaneer/1.0 (+https://github.com/woliveiras/bookaneer)"
)

var (
	resultLinkRe = regexp.MustCompile(`(?s)<a[^>]*class="[^"]*result__a[^"]*"[^>]*href="([^"]+)"[^>]*>(.*?)</a>`)
	tagRe        = regexp.MustCompile(`<[^>]+>`)
)

type Provider struct {
	name       string
	domain     string
	formatHint string
	httpClient *http.Client
}

func New(name, domain, formatHint string) *Provider {
	return &Provider{
		name:       name,
		domain:     strings.ToLower(strings.TrimSpace(domain)),
		formatHint: strings.TrimSpace(formatHint),
		httpClient: &http.Client{Timeout: 20 * time.Second},
	}
}

func (p *Provider) Name() string {
	return p.name
}

func (p *Provider) Search(ctx context.Context, query string) ([]library.SearchResult, error) {
	q := strings.TrimSpace(query)
	if q == "" {
		return []library.SearchResult{}, nil
	}

	params := url.Values{}
	params.Set("q", fmt.Sprintf("site:%s %s", p.domain, q))
	endpoint := duckURL + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
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

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	body := string(bodyBytes)

	matches := resultLinkRe.FindAllStringSubmatch(body, 60)
	results := make([]library.SearchResult, 0, len(matches))
	seen := map[string]struct{}{}

	for _, m := range matches {
		if len(m) < 3 {
			continue
		}

		rawHref := html.UnescapeString(strings.TrimSpace(m[1]))
		title := strings.TrimSpace(tagRe.ReplaceAllString(html.UnescapeString(m[2]), ""))
		if rawHref == "" || title == "" {
			continue
		}

		href := resolveDuckRedirect(rawHref)
		u, err := url.Parse(href)
		if err != nil || u.Host == "" {
			continue
		}

		host := strings.ToLower(u.Host)
		if !strings.Contains(host, p.domain) {
			continue
		}

		if _, ok := seen[href]; ok {
			continue
		}
		seen[href] = struct{}{}

		format := p.formatHint
		if format == "" {
			format = detectFormat(href)
		}
		if format == "" {
			format = "html"
		}

		results = append(results, library.SearchResult{
			ID:          href,
			Title:       title,
			Format:      format,
			InfoURL:     href,
			DownloadURL: href,
			Provider:    p.name,
		})

		if len(results) >= 20 {
			break
		}
	}

	return results, nil
}

func (p *Provider) GetDownloadLink(_ context.Context, id string) (string, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return "", fmt.Errorf("empty id")
	}
	return id, nil
}

func resolveDuckRedirect(href string) string {
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}

	if strings.HasPrefix(href, "//") {
		return "https:" + href
	}

	u, err := url.Parse(href)
	if err != nil {
		return href
	}

	if u.Path == "/l/" || strings.HasPrefix(u.Path, "/l/") {
		if target := u.Query().Get("uddg"); strings.TrimSpace(target) != "" {
			decoded, err := url.QueryUnescape(target)
			if err == nil {
				return decoded
			}
			return target
		}
	}

	return href
}

func detectFormat(href string) string {
	l := strings.ToLower(href)
	switch {
	case strings.Contains(l, ".epub"):
		return "epub"
	case strings.Contains(l, ".pdf"):
		return "pdf"
	case strings.Contains(l, ".mobi") || strings.Contains(l, ".azw"):
		return "mobi"
	default:
		return ""
	}
}

var _ library.Provider = (*Provider)(nil)
