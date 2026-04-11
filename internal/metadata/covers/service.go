// Package covers provides cover image fetching and caching.
package covers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const defaultTimeout = 30 * time.Second

// Service handles cover image fetching and caching.
type Service struct {
	client   *http.Client
	cacheDir string
}

// CoverType represents the type of cover.
type CoverType string

const (
	CoverTypeAuthor CoverType = "authors"
	CoverTypeBook   CoverType = "books"
)

// New creates a new cover service.
func New(cacheDir string, client *http.Client) (*Service, error) {
	if client == nil {
		client = &http.Client{Timeout: defaultTimeout}
	}

	for _, dir := range []string{"authors", "books"} {
		path := filepath.Join(cacheDir, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return nil, fmt.Errorf("create cache dir %s: %w", path, err)
		}
	}

	return &Service{
		client:   client,
		cacheDir: cacheDir,
	}, nil
}

// FetchAndCache fetches a cover image and caches it locally.
func (s *Service) FetchAndCache(ctx context.Context, coverType CoverType, id string, url string) (string, error) {
	if url == "" {
		return "", fmt.Errorf("empty cover URL")
	}

	ext := getExtension(url)
	hash := hashURL(url)
	filename := fmt.Sprintf("%s_%s%s", sanitizeID(id), hash, ext)
	cachePath := filepath.Join(s.cacheDir, string(coverType), filename)

	if _, err := os.Stat(cachePath); err == nil {
		return cachePath, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("User-Agent", "Bookaneer/1.0")
	req.Header.Set("Accept", "image/*")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch cover: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetch cover: status %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return "", fmt.Errorf("invalid content type: %s", contentType)
	}

	tmpPath := cachePath + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}

	_, err = io.Copy(f, resp.Body)
	if closeErr := f.Close(); closeErr != nil && err == nil {
		err = closeErr
	}
	if err != nil {
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("write cover: %w", err)
	}

	if err := os.Rename(tmpPath, cachePath); err != nil {
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("rename temp file: %w", err)
	}

	return cachePath, nil
}

// GetCachedPath returns the cached path if it exists.
func (s *Service) GetCachedPath(coverType CoverType, id string, url string) (string, bool) {
	if url == "" {
		return "", false
	}

	ext := getExtension(url)
	hash := hashURL(url)
	filename := fmt.Sprintf("%s_%s%s", sanitizeID(id), hash, ext)
	cachePath := filepath.Join(s.cacheDir, string(coverType), filename)

	if _, err := os.Stat(cachePath); err == nil {
		return cachePath, true
	}

	return "", false
}

// Delete removes a cached cover.
func (s *Service) Delete(coverType CoverType, id string, url string) error {
	if url == "" {
		return nil
	}

	ext := getExtension(url)
	hash := hashURL(url)
	filename := fmt.Sprintf("%s_%s%s", sanitizeID(id), hash, ext)
	cachePath := filepath.Join(s.cacheDir, string(coverType), filename)

	return os.Remove(cachePath)
}

func hashURL(url string) string {
	h := sha256.Sum256([]byte(url))
	return hex.EncodeToString(h[:8])
}

func getExtension(url string) string {
	if idx := strings.Index(url, "?"); idx != -1 {
		url = url[:idx]
	}

	ext := filepath.Ext(url)
	if ext == "" {
		ext = ".jpg"
	}

	ext = strings.ToLower(ext)
	switch ext {
	case ".jpeg":
		ext = ".jpg"
	case ".png", ".gif", ".webp":
		// Keep as-is
	default:
		ext = ".jpg"
	}

	return ext
}

func sanitizeID(id string) string {
	safe := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, id)

	if len(safe) > 64 {
		safe = safe[:64]
	}

	return safe
}
