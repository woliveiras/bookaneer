package direct

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/woliveiras/bookaneer/internal/download"
)

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
