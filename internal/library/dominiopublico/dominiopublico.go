// Package dominiopublico provides integration with Portal Domínio Público search.
package dominiopublico

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/woliveiras/bookaneer/internal/library"
)

const (
	baseURL        = "http://www.dominiopublico.gov.br"
	searchEndpoint = baseURL + "/pesquisa/PesquisaObraForm.do"
	detailPath     = "/pesquisa/DetalheObraForm.do"
	userAgent      = "Bookaneer/1.0 (+https://github.com/woliveiras/bookaneer)"
)

var (
	tagRe        = regexp.MustCompile(`<[^>]+>`)
	detailLinkRe = regexp.MustCompile(`(?is)href=["']([^"']*DetalheObraForm\.do[^"']*co_obra=(\d+)[^"']*)["'][^>]*>(.*?)</a>`)
	fileLinkRe   = regexp.MustCompile(`(?is)href=["']([^"']+)["']`)
)

type Provider struct {
	httpClient *http.Client
}

func New() *Provider {
	return &Provider{httpClient: &http.Client{Timeout: 25 * time.Second}}
}

func (p *Provider) Name() string {
	return "dominio-publico"
}

func (p *Provider) Search(ctx context.Context, query string) ([]library.SearchResult, error) {
	q := strings.TrimSpace(query)
	if q == "" {
		return []library.SearchResult{}, nil
	}

	form := url.Values{}
	form.Set("co_obra", "")
	form.Set("co_flag_periodico", "0")
	form.Set("co_midia", "3")
	form.Set("co_categoria", "22")
	form.Set("co_autor", "")
	form.Set("no_autor", "")
	form.Set("ds_titulo", q)
	form.Set("co_idioma", "")
	form.Set("select_action", "Submit")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, searchEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

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

	if strings.Contains(body, "__cf_chl_") || strings.Contains(strings.ToLower(body), "just a moment") {
		return nil, fmt.Errorf("dominio publico blocked by cloudflare challenge")
	}

	results := p.parseSearchResults(body)
	if len(results) > 20 {
		results = results[:20]
	}

	return results, nil
}

func (p *Provider) parseSearchResults(body string) []library.SearchResult {
	matches := detailLinkRe.FindAllStringSubmatch(body, -1)
	results := make([]library.SearchResult, 0, len(matches))
	seen := map[string]struct{}{}

	for _, m := range matches {
		if len(m) < 4 {
			continue
		}

		detailURL := resolveURL(strings.TrimSpace(html.UnescapeString(m[1])))
		id := strings.TrimSpace(m[2])
		title := normalizeText(m[3])
		if id == "" || title == "" || detailURL == "" {
			continue
		}

		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}

		results = append(results, library.SearchResult{
			ID:          id,
			Title:       title,
			Language:    "pt",
			Format:      "pdf",
			InfoURL:     detailURL,
			DownloadURL: detailURL,
			Provider:    "dominio-publico",
		})
	}

	return results
}

func (p *Provider) GetDownloadLink(ctx context.Context, id string) (string, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return "", fmt.Errorf("empty id")
	}

	detailURL := id
	if _, err := strconv.Atoi(id); err == nil {
		detailURL = fmt.Sprintf("%s%s?select_action=&co_obra=%s", baseURL, detailPath, url.QueryEscape(id))
	}

	resolved, err := p.resolveDownloadFromDetail(ctx, detailURL)
	if err != nil {
		return "", err
	}

	return resolved, nil
}

func (p *Provider) resolveDownloadFromDetail(ctx context.Context, detailURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, detailURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	body := string(bodyBytes)

	if strings.Contains(body, "__cf_chl_") || strings.Contains(strings.ToLower(body), "just a moment") {
		return "", fmt.Errorf("dominio publico blocked by cloudflare challenge")
	}

	if dl := extractBestFileURL(body); dl != "" {
		return resolveURL(dl), nil
	}

	return detailURL, nil
}

func extractBestFileURL(body string) string {
	matches := fileLinkRe.FindAllStringSubmatch(body, -1)
	best := ""
	bestScore := -1

	for _, m := range matches {
		if len(m) < 2 {
			continue
		}

		u := strings.TrimSpace(html.UnescapeString(m[1]))
		if u == "" {
			continue
		}

		l := strings.ToLower(u)
		score := -1
		switch {
		case strings.Contains(l, ".epub"):
			score = 5
		case strings.Contains(l, ".pdf"):
			score = 4
		case strings.Contains(l, ".mobi"), strings.Contains(l, ".azw"):
			score = 3
		case strings.Contains(l, ".txt"):
			score = 2
		case strings.Contains(l, "download"), strings.Contains(l, "baixar"):
			score = 1
		}

		if score > bestScore {
			best = u
			bestScore = score
		}
	}

	return best
}

func resolveURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		return raw
	}
	if strings.HasPrefix(raw, "//") {
		return "http:" + raw
	}
	if strings.HasPrefix(raw, "/") {
		return baseURL + raw
	}

	return baseURL + "/" + raw
}

func normalizeText(s string) string {
	plain := tagRe.ReplaceAllString(html.UnescapeString(s), " ")
	plain = strings.TrimSpace(strings.Join(strings.Fields(plain), " "))
	return plain
}

var _ library.Provider = (*Provider)(nil)
