// Package aozora provides integration with Aozora Bunko's public catalog.
package aozora

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/woliveiras/bookaneer/internal/library"
)

const (
	catalogZIPURL = "https://www.aozora.gr.jp/index_pages/list_person_all_extended_utf8.zip"
	userAgent     = "Bookaneer/1.0 (+https://github.com/woliveiras/bookaneer)"
)

type catalogEntry struct {
	workID      string
	title       string
	author      string
	language    string
	format      string
	downloadURL string
	infoURL     string
}

type Provider struct {
	httpClient *http.Client

	mu      sync.RWMutex
	loaded  time.Time
	entries []catalogEntry
	byID    map[string]catalogEntry
}

func New() *Provider {
	return &Provider{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		byID:       make(map[string]catalogEntry),
	}
}

func (p *Provider) Name() string {
	return "aozora"
}

func (p *Provider) Search(ctx context.Context, query string) ([]library.SearchResult, error) {
	if err := p.ensureCatalog(ctx); err != nil {
		return nil, err
	}

	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return []library.SearchResult{}, nil
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	results := make([]library.SearchResult, 0, 25)
	for _, entry := range p.entries {
		haystackTitle := strings.ToLower(entry.title)
		haystackAuthor := strings.ToLower(entry.author)
		if !strings.Contains(haystackTitle, q) && !strings.Contains(haystackAuthor, q) {
			continue
		}

		authors := []string{}
		if strings.TrimSpace(entry.author) != "" {
			authors = []string{entry.author}
		}

		results = append(results, library.SearchResult{
			ID:          entry.workID,
			Title:       entry.title,
			Authors:     authors,
			Language:    entry.language,
			Format:      entry.format,
			DownloadURL: entry.downloadURL,
			InfoURL:     entry.infoURL,
			Provider:    "aozora",
		})

		if len(results) >= 25 {
			break
		}
	}

	return results, nil
}

func (p *Provider) GetDownloadLink(ctx context.Context, id string) (string, error) {
	if err := p.ensureCatalog(ctx); err != nil {
		return "", err
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	entry, ok := p.byID[strings.TrimSpace(id)]
	if !ok || strings.TrimSpace(entry.downloadURL) == "" {
		return "", fmt.Errorf("download not found for id %q", id)
	}

	return entry.downloadURL, nil
}

func (p *Provider) ensureCatalog(ctx context.Context) error {
	p.mu.RLock()
	fresh := len(p.entries) > 0 && time.Since(p.loaded) < 24*time.Hour
	p.mu.RUnlock()
	if fresh {
		return nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.entries) > 0 && time.Since(p.loaded) < 24*time.Hour {
		return nil
	}

	entries, err := p.fetchCatalog(ctx)
	if err != nil {
		return err
	}

	byID := make(map[string]catalogEntry, len(entries))
	for _, entry := range entries {
		byID[entry.workID] = entry
	}

	p.entries = entries
	p.byID = byID
	p.loaded = time.Now()
	return nil
}

func (p *Provider) fetchCatalog(ctx context.Context) ([]catalogEntry, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, catalogZIPURL, nil)
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

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	zr, err := zip.NewReader(bytes.NewReader(payload), int64(len(payload)))
	if err != nil {
		return nil, err
	}
	if len(zr.File) == 0 {
		return nil, fmt.Errorf("empty catalog archive")
	}

	f, err := zr.File[0].Open()
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	reader := csv.NewReader(f)
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1

	// Skip header.
	if _, err := reader.Read(); err != nil {
		return nil, err
	}

	entries := make([]catalogEntry, 0, 5000)
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		entry, ok := parseRecord(record)
		if !ok {
			continue
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func parseRecord(record []string) (catalogEntry, bool) {
	const (
		idxWorkID      = 0
		idxTitle       = 1
		idxInfoURL     = 13
		idxAuthorLast  = 16
		idxAuthorFirst = 17
		idxTextZIPURL  = 43
		idxHTMLURL     = 48
	)

	if len(record) <= idxHTMLURL {
		return catalogEntry{}, false
	}

	workID := strings.TrimSpace(record[idxWorkID])
	title := strings.TrimSpace(record[idxTitle])
	infoURL := strings.TrimSpace(record[idxInfoURL])
	textURL := strings.TrimSpace(record[idxTextZIPURL])
	htmlURL := strings.TrimSpace(record[idxHTMLURL])

	if workID == "" || title == "" || infoURL == "" {
		return catalogEntry{}, false
	}

	format := ""
	downloadURL := ""
	switch {
	case textURL != "":
		format = "zip"
		downloadURL = textURL
	case htmlURL != "":
		format = "html"
		downloadURL = htmlURL
	default:
		return catalogEntry{}, false
	}

	author := strings.TrimSpace(strings.TrimSpace(record[idxAuthorLast]) + " " + strings.TrimSpace(record[idxAuthorFirst]))

	return catalogEntry{
		workID:      workID,
		title:       title,
		author:      author,
		language:    "ja",
		format:      format,
		downloadURL: downloadURL,
		infoURL:     infoURL,
	}, true
}

var _ library.Provider = (*Provider)(nil)
