// Package openlibrarypublic provides integration with Open Library public scans.
package openlibrarypublic

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
	searchURL  = "https://openlibrary.org/search.json"
	archiveURL = "https://archive.org"
	userAgent  = "Bookaneer/1.0 (+https://github.com/woliveiras/bookaneer)"
)

type Provider struct {
	httpClient *http.Client
}

func New() *Provider {
	return &Provider{httpClient: &http.Client{Timeout: 20 * time.Second}}
}

func (p *Provider) Name() string {
	return "openlibrary-public"
}

type searchResponse struct {
	Docs []struct {
		Key              string   `json:"key"`
		Title            string   `json:"title"`
		AuthorName       []string `json:"author_name"`
		FirstPublishYear int      `json:"first_publish_year"`
		Language         []string `json:"language"`
		IA               []string `json:"ia"`
		PublicScan       bool     `json:"public_scan_b"`
		HasFulltext      bool     `json:"has_fulltext"`
		CoverI           int      `json:"cover_i"`
		Availability     struct {
			Identifier string `json:"identifier"`
			IsReadable bool   `json:"is_readable"`
		} `json:"availability"`
	} `json:"docs"`
}

func (p *Provider) Search(ctx context.Context, query string) ([]library.SearchResult, error) {
	params := url.Values{}
	params.Set("q", query)
	params.Set("limit", "20")
	params.Set("fields", "key,title,author_name,first_publish_year,language,ia,public_scan_b,has_fulltext,cover_i,availability")

	endpoint := fmt.Sprintf("%s?%s", searchURL, params.Encode())
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

	var sr searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return nil, err
	}

	results := make([]library.SearchResult, 0, len(sr.Docs))
	for _, doc := range sr.Docs {
		identifier := strings.TrimSpace(doc.Availability.Identifier)
		if identifier == "" && len(doc.IA) > 0 {
			identifier = strings.TrimSpace(doc.IA[0])
		}
		if identifier == "" {
			continue
		}

		if !doc.PublicScan && !doc.Availability.IsReadable {
			continue
		}

		format, downloadURL, err := p.resolveArchiveDownload(ctx, identifier)
		if err != nil || format == "" || downloadURL == "" {
			continue
		}

		language := ""
		if len(doc.Language) > 0 {
			language = doc.Language[0]
		}

		res := library.SearchResult{
			ID:          identifier,
			Title:       doc.Title,
			Authors:     doc.AuthorName,
			Year:        doc.FirstPublishYear,
			Language:    language,
			Format:      format,
			Provider:    "openlibrary-public",
			InfoURL:     fmt.Sprintf("https://openlibrary.org%s", doc.Key),
			DownloadURL: downloadURL,
		}
		if doc.CoverI > 0 {
			res.CoverURL = fmt.Sprintf("https://covers.openlibrary.org/b/id/%d-M.jpg", doc.CoverI)
		}
		results = append(results, res)
	}

	return results, nil
}

func (p *Provider) resolveArchiveDownload(ctx context.Context, identifier string) (string, string, error) {
	metaURL := fmt.Sprintf("%s/metadata/%s/files", archiveURL, url.PathEscape(identifier))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, metaURL, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var filesResp struct {
		Result []struct {
			Name string `json:"name"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&filesResp); err != nil {
		return "", "", err
	}

	for _, ext := range []string{"epub", "pdf", "mobi"} {
		suffix := "." + ext
		for _, f := range filesResp.Result {
			name := strings.TrimSpace(f.Name)
			if strings.HasSuffix(strings.ToLower(name), suffix) {
				return ext, fmt.Sprintf("%s/download/%s/%s", archiveURL, identifier, url.PathEscape(name)), nil
			}
		}
	}

	return "", "", fmt.Errorf("no supported files")
}

func (p *Provider) GetDownloadLink(ctx context.Context, id string) (string, error) {
	_, u, err := p.resolveArchiveDownload(ctx, id)
	if err != nil {
		return "", err
	}
	return u, nil
}

var _ library.Provider = (*Provider)(nil)
