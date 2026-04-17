// Package libgen provides integration with Library Genesis for ebook downloads.
package libgen

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"

	"github.com/woliveiras/bookaneer/internal/library"
)

// htmlMirrors lists LibGen HTML search base URLs in preferred order.
// libgen.bz uses /index.php?req= for search (JSON API is non-functional on this mirror).
var htmlMirrors = []string{
	"https://libgen.bz",
}

const (
	userAgent  = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	searchPath = "/index.php"
)

var (
	reMD5      = regexp.MustCompile(`/ads\.php\?md5=([0-9a-fA-F]{32})`)
	reSizeNum  = regexp.MustCompile(`([\d.]+)\s*(KB|MB|GB)`)
	reISBNTail = regexp.MustCompile(`\s+\d{10,}.*`)
)

// Provider implements library.Provider for Library Genesis.
type Provider struct {
	httpClient *http.Client
}

// New creates a new LibGen provider.
func New() *Provider {
	return &Provider{
		httpClient: &http.Client{Timeout: 12 * time.Second},
	}
}

func (p *Provider) Name() string {
	return "libgen"
}

func (p *Provider) Search(ctx context.Context, query string) ([]library.SearchResult, error) {
	for _, mirror := range htmlMirrors {
		results, err := p.searchMirror(ctx, mirror, query)
		if err != nil {
			slog.Debug("libgen mirror failed", "mirror", mirror, "error", err)
			continue
		}
		if len(results) > 0 {
			return results, nil
		}
	}
	return nil, nil
}

func (p *Provider) searchMirror(ctx context.Context, baseURL, query string) ([]library.SearchResult, error) {
	params := url.Values{}
	params.Set("req", query)
	params.Set("res", "25")
	for _, c := range []string{"t", "a", "s", "y", "p", "i"} {
		params.Add("columns[]", c)
	}
	for _, o := range []string{"f", "e"} {
		params.Add("objects[]", o)
	}
	params.Add("topics[]", "l")

	reqURL := baseURL + searchPath + "?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	return parseSearchHTML(resp.Body, baseURL)
}

// parseSearchHTML parses the libgen.bz /index.php HTML search result table.
//
// Table columns (0-indexed):
//
//	0: Title (first anchor text)
//	1: Author(s)
//	2: Publisher
//	3: Year
//	4: Language
//	5: Pages (ignored)
//	6: Size
//	7: Extension
//	8: Mirrors (contains /ads.php?md5=MD5 links)
func parseSearchHTML(r io.Reader, baseURL string) ([]library.SearchResult, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("html parse: %w", err)
	}

	table := findResultsTable(doc)
	if table == nil {
		return nil, nil
	}

	var results []library.SearchResult
	headerSkipped := false

	for row := table.FirstChild; row != nil; row = row.NextSibling {
		if row.Type != html.ElementNode || row.Data != "tr" {
			continue
		}

		var cells []*html.Node
		for td := row.FirstChild; td != nil; td = td.NextSibling {
			if td.Type == html.ElementNode && td.Data == "td" {
				cells = append(cells, td)
			}
		}

		if len(cells) < 9 {
			continue
		}

		if !headerSkipped {
			headerSkipped = true
			continue
		}

		if result, ok := parseCells(cells, baseURL); ok {
			results = append(results, result)
		}
	}

	return results, nil
}

// findResultsTable walks the HTML tree and returns the tbody (or table) node
// that contains the book results, identified by the presence of /ads.php?md5= links.
func findResultsTable(n *html.Node) *html.Node {
	if n.Type == html.ElementNode && (n.Data == "tbody" || n.Data == "table") {
		if tableBodyHasResults(n) {
			return n
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if result := findResultsTable(c); result != nil {
			return result
		}
	}
	return nil
}

func tableBodyHasResults(n *html.Node) bool {
	for row := n.FirstChild; row != nil; row = row.NextSibling {
		if row.Type != html.ElementNode || row.Data != "tr" {
			continue
		}
		for td := row.FirstChild; td != nil; td = td.NextSibling {
			if td.Type != html.ElementNode || td.Data != "td" {
				continue
			}
			if findAnchorWithMD5(td) != "" {
				return true
			}
		}
	}
	return false
}

func parseCells(cells []*html.Node, baseURL string) (library.SearchResult, bool) {
	ext := strings.ToLower(strings.TrimSpace(innerText(cells[7])))
	if ext != "epub" && ext != "pdf" && ext != "mobi" && ext != "azw3" {
		return library.SearchResult{}, false
	}

	// Title: text content of the first anchor in cells[0]
	title := strings.TrimSpace(firstAnchorText(cells[0]))
	if title == "" {
		title = strings.TrimSpace(innerText(cells[0]))
	}
	title = cleanTitle(title)
	if title == "" {
		return library.SearchResult{}, false
	}

	md5 := findAnchorWithMD5(cells[8])
	if md5 == "" {
		return library.SearchResult{}, false
	}

	authorRaw := strings.TrimSpace(innerText(cells[1]))
	publisher := strings.TrimSpace(innerText(cells[2]))
	yearRaw := strings.TrimSpace(innerText(cells[3]))
	lang := strings.TrimSpace(innerText(cells[4]))
	sizeRaw := strings.TrimSpace(innerText(cells[6]))

	year, _ := strconv.Atoi(strings.TrimSpace(strings.SplitN(yearRaw, ";", 2)[0]))
	authors := splitAuthors(authorRaw)
	size := parseSize(sizeRaw)

	return library.SearchResult{
		ID:          md5,
		Title:       title,
		Authors:     authors,
		Publisher:   publisher,
		Year:        year,
		Language:    lang,
		Format:      ext,
		Size:        size,
		InfoURL:     fmt.Sprintf("%s/ads.php?md5=%s", baseURL, md5),
		DownloadURL: fmt.Sprintf("%s/ads.php?md5=%s", baseURL, md5),
		Provider:    "libgen",
	}, true
}

// innerText returns the concatenated text content of a node and all descendants.
func innerText(n *html.Node) string {
	if n == nil {
		return ""
	}
	var sb strings.Builder
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.TextNode {
			sb.WriteString(node.Data)
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return sb.String()
}

// firstAnchorText returns the inner text of the first <a> element in n.
func firstAnchorText(n *html.Node) string {
	if n == nil {
		return ""
	}
	if n.Type == html.ElementNode && n.Data == "a" {
		return innerText(n)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if t := firstAnchorText(c); t != "" {
			return t
		}
	}
	return ""
}

// findAnchorWithMD5 returns the MD5 hash from the first /ads.php?md5=... href found in n.
func findAnchorWithMD5(n *html.Node) string {
	if n == nil {
		return ""
	}
	if n.Type == html.ElementNode && n.Data == "a" {
		for _, attr := range n.Attr {
			if attr.Key == "href" {
				if m := reMD5.FindStringSubmatch(attr.Val); len(m) == 2 {
					return m[1]
				}
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if md5 := findAnchorWithMD5(c); md5 != "" {
			return md5
		}
	}
	return ""
}

// cleanTitle strips ISBNs and trailing libgen IDs from the title cell text.
func cleanTitle(s string) string {
	// Strip from the first newline (libgen appends \r\n followed by bookmark/ID tags)
	if idx := strings.IndexByte(s, '\n'); idx != -1 {
		s = s[:idx]
	}
	// Strip from the first long numeric sequence (ISBN, ≥10 digits)
	s = reISBNTail.ReplaceAllString(s, "")
	return strings.TrimRight(strings.TrimSpace(s), ",;.")
}

// splitAuthors splits an author string by commas into individual names.
func splitAuthors(s string) []string {
	var authors []string
	for _, a := range strings.Split(s, ",") {
		if a = strings.TrimSpace(a); a != "" {
			authors = append(authors, a)
		}
	}
	if len(authors) == 0 {
		return []string{s}
	}
	return authors
}

// parseSize converts a human-readable size string (e.g. "5 MB") to bytes.
func parseSize(s string) int64 {
	m := reSizeNum.FindStringSubmatch(s)
	if len(m) < 3 {
		return 0
	}
	val, _ := strconv.ParseFloat(m[1], 64)
	switch strings.ToUpper(m[2]) {
	case "KB":
		return int64(val * 1024)
	case "MB":
		return int64(val * 1024 * 1024)
	case "GB":
		return int64(val * 1024 * 1024 * 1024)
	}
	return 0
}

func (p *Provider) GetDownloadLink(_ context.Context, id string) (string, error) {
	return fmt.Sprintf("%s/ads.php?md5=%s", htmlMirrors[0], id), nil
}

var _ library.Provider = (*Provider)(nil)
