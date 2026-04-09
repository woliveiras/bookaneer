// Package direct provides an internal HTTP download client for direct downloads.
// Unlike external clients (qBittorrent, SABnzbd), this downloads files directly
// using HTTP GET requests, suitable for digital library sources like Internet Archive,
// LibGen, and Anna's Archive.
package direct

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/woliveiras/bookaneer/internal/download"
)

func init() {
	download.RegisterFactory(download.ClientTypeDirect, func(cfg download.ClientConfig) (download.Client, error) {
		return New(Config{
			Name:        cfg.Name,
			DownloadDir: cfg.DownloadDir,
		}), nil
	})
}

// Config holds Direct downloader configuration.
type Config struct {
	Name        string
	DownloadDir string // Directory to save downloaded files
}

// downloadItem tracks an active download.
type downloadItem struct {
	id           string
	name         string
	url          string
	status       download.Status
	progress     float64
	size         int64
	downloaded   int64
	speed        int64
	savePath     string
	errorMessage string
	addedAt      time.Time
	completedAt  *time.Time
}

// Client is a direct HTTP download client.
type Client struct {
	cfg        Config
	httpClient *http.Client

	mu        sync.RWMutex
	downloads map[string]*downloadItem
}

// New creates a new Direct download client.
func New(cfg Config) *Client {
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Minute, // Long timeout for large files
			Transport: &http.Transport{
				MaxIdleConns:        10,
				MaxIdleConnsPerHost: 5,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		downloads: make(map[string]*downloadItem),
	}
}

// Name returns the client name.
func (c *Client) Name() string {
	return c.cfg.Name
}

// Type returns the client type.
func (c *Client) Type() string {
	return download.ClientTypeDirect
}

// Test verifies the download directory exists and is writable.
func (c *Client) Test(ctx context.Context) error {
	if c.cfg.DownloadDir == "" {
		return fmt.Errorf("download directory not configured")
	}

	if err := os.MkdirAll(c.cfg.DownloadDir, 0755); err != nil {
		return fmt.Errorf("cannot create download directory: %w", err)
	}

	testFile := filepath.Join(c.cfg.DownloadDir, ".bookaneer_test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("cannot write to download directory: %w", err)
	}
	os.Remove(testFile)

	return nil
}

// Add starts downloading a file from the given URL.
func (c *Client) Add(ctx context.Context, item download.AddItem) (string, error) {
	// Generate unique ID
	hash := sha256.Sum256([]byte(item.DownloadURL + time.Now().String()))
	id := hex.EncodeToString(hash[:8])

	// Determine save path
	savePath := c.cfg.DownloadDir
	if item.SavePath != "" {
		savePath = item.SavePath
	}

	// Create download entry
	dl := &downloadItem{
		id:       id,
		name:     item.Name,
		url:      item.DownloadURL,
		status:   download.StatusQueued,
		savePath: savePath,
		addedAt:  time.Now(),
	}

	c.mu.Lock()
	c.downloads[id] = dl
	c.mu.Unlock()

	// Start download in background
	go c.downloadFile(context.Background(), dl, item.Headers)

	return id, nil
}

// downloadFile performs the actual HTTP download.
func (c *Client) downloadFile(ctx context.Context, dl *downloadItem, headers map[string]string) {
	c.updateStatus(dl.id, download.StatusDownloading, "")

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, dl.url, nil)
	if err != nil {
		c.updateStatus(dl.id, download.StatusFailed, fmt.Sprintf("invalid URL: %v", err))
		return
	}

	// Add custom headers
	req.Header.Set("User-Agent", "Bookaneer/1.0")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.updateStatus(dl.id, download.StatusFailed, fmt.Sprintf("download failed: %v", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.updateStatus(dl.id, download.StatusFailed, fmt.Sprintf("HTTP %d: %s", resp.StatusCode, resp.Status))
		return
	}

	// Get file size
	c.mu.Lock()
	dl.size = resp.ContentLength
	c.mu.Unlock()

	// Determine filename
	filename := c.extractFilename(resp, dl.name)
	if filename == "" {
		filename = dl.name + ".epub" // Default to epub
	}

	// Create output file
	filePath := filepath.Join(dl.savePath, filename)
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		c.updateStatus(dl.id, download.StatusFailed, fmt.Sprintf("cannot create directory: %v", err))
		return
	}

	outFile, err := os.Create(filePath)
	if err != nil {
		c.updateStatus(dl.id, download.StatusFailed, fmt.Sprintf("cannot create file: %v", err))
		return
	}
	defer outFile.Close()

	// Download with progress tracking
	c.mu.Lock()
	dl.savePath = filePath
	c.mu.Unlock()

	buf := make([]byte, 32*1024) // 32KB buffer
	var downloaded int64
	startTime := time.Now()
	lastUpdate := startTime

	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := outFile.Write(buf[:n]); writeErr != nil {
				c.updateStatus(dl.id, download.StatusFailed, fmt.Sprintf("write error: %v", writeErr))
				return
			}
			downloaded += int64(n)

			// Update progress periodically
			if time.Since(lastUpdate) > 500*time.Millisecond {
				c.mu.Lock()
				dl.downloaded = downloaded
				if dl.size > 0 {
					dl.progress = float64(downloaded) / float64(dl.size) * 100
				}
				elapsed := time.Since(startTime).Seconds()
				if elapsed > 0 {
					dl.speed = int64(float64(downloaded) / elapsed)
				}
				c.mu.Unlock()
				lastUpdate = time.Now()
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			c.updateStatus(dl.id, download.StatusFailed, fmt.Sprintf("read error: %v", readErr))
			return
		}
	}

	// Mark as completed
	now := time.Now()
	c.mu.Lock()
	dl.status = download.StatusCompleted
	dl.progress = 100
	dl.downloaded = downloaded
	dl.size = downloaded
	dl.completedAt = &now
	c.mu.Unlock()
}

// extractFilename extracts filename from Content-Disposition header or URL.
// If a meaningful fallback name is provided (contains ebook extension), use it directly.
// This ensures books are saved with semantic names like "Author - Title.epub"
// instead of server-side IDs like "bv00180a.epub".
func (c *Client) extractFilename(resp *http.Response, fallback string) string {
	// Check if fallback already has a valid ebook extension - use it directly
	lowerFallback := strings.ToLower(fallback)
	validExtensions := []string{".epub", ".pdf", ".mobi", ".azw", ".azw3", ".fb2", ".cbz", ".cbr"}
	for _, ext := range validExtensions {
		if strings.HasSuffix(lowerFallback, ext) {
			return sanitizeFilename(fallback)
		}
	}

	// Fallback doesn't have extension, try to get extension from response
	ext := ""

	// Try Content-Disposition header
	cd := resp.Header.Get("Content-Disposition")
	if cd != "" {
		_, params, err := mime.ParseMediaType(cd)
		if err == nil && params["filename"] != "" {
			serverFilename := params["filename"]
			if idx := strings.LastIndex(serverFilename, "."); idx != -1 {
				ext = serverFilename[idx:]
			}
		}
	}

	// Try URL path if no extension found
	if ext == "" {
		urlPath := resp.Request.URL.Path
		if idx := strings.LastIndex(urlPath, "."); idx != -1 {
			ext = urlPath[idx:]
		}
	}

	// If we got extension, add it to fallback
	if ext != "" && fallback != "" {
		return sanitizeFilename(fallback + ext)
	}

	// Last resort: use URL filename
	urlPath := resp.Request.URL.Path
	if idx := strings.LastIndex(urlPath, "/"); idx != -1 {
		filename := urlPath[idx+1:]
		if filename != "" && strings.Contains(filename, ".") {
			return sanitizeFilename(filename)
		}
	}

	// Default to fallback with .epub
	return sanitizeFilename(fallback + ".epub")
}

// sanitizeFilename removes unsafe characters from filename.
func sanitizeFilename(name string) string {
	// Remove path separators and other unsafe characters
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	return replacer.Replace(name)
}

// updateStatus updates the download status.
func (c *Client) updateStatus(id string, status download.Status, errorMsg string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if dl, ok := c.downloads[id]; ok {
		dl.status = status
		dl.errorMessage = errorMsg
	}
}

// Remove cancels and removes a download.
func (c *Client) Remove(ctx context.Context, id string, deleteData bool) error {
	c.mu.Lock()
	dl, ok := c.downloads[id]
	if !ok {
		c.mu.Unlock()
		return fmt.Errorf("download %s not found", id)
	}

	// If download is in progress, we should cancel it (future: context cancellation)
	savePath := dl.savePath
	delete(c.downloads, id)
	c.mu.Unlock()

	// Delete the file if requested
	if deleteData && savePath != "" {
		if err := os.Remove(savePath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("cannot delete file: %w", err)
		}
	}

	return nil
}

// GetStatus returns the status of a specific download.
func (c *Client) GetStatus(ctx context.Context, id string) (*download.ItemStatus, error) {
	c.mu.RLock()
	dl, ok := c.downloads[id]
	c.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("download %s not found", id)
	}

	return c.itemToStatus(dl), nil
}

// GetQueue returns all downloads.
func (c *Client) GetQueue(ctx context.Context) ([]download.ItemStatus, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]download.ItemStatus, 0, len(c.downloads))
	for _, dl := range c.downloads {
		result = append(result, *c.itemToStatus(dl))
	}
	return result, nil
}

// itemToStatus converts a downloadItem to ItemStatus.
func (c *Client) itemToStatus(dl *downloadItem) *download.ItemStatus {
	var eta time.Duration
	if dl.speed > 0 && dl.size > 0 && dl.downloaded < dl.size {
		remaining := dl.size - dl.downloaded
		eta = time.Duration(remaining/dl.speed) * time.Second
	}

	return &download.ItemStatus{
		ID:             dl.id,
		Name:           dl.name,
		Status:         dl.status,
		Progress:       dl.progress,
		Size:           dl.size,
		DownloadedSize: dl.downloaded,
		Speed:          dl.speed,
		ETA:            eta,
		SavePath:       dl.savePath,
		ErrorMessage:   dl.errorMessage,
		AddedAt:        dl.addedAt,
		CompletedAt:    dl.completedAt,
	}
}

// CompletedDownloads returns paths of completed downloads and removes them from tracking.
func (c *Client) CompletedDownloads() []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	var paths []string
	for id, dl := range c.downloads {
		if dl.status == download.StatusCompleted && dl.savePath != "" {
			paths = append(paths, dl.savePath)
			delete(c.downloads, id)
		}
	}
	return paths
}
