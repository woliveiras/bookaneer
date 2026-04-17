package direct

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/woliveiras/bookaneer/internal/bypass"
	"github.com/woliveiras/bookaneer/internal/bypass/challenge"
	"github.com/woliveiras/bookaneer/internal/download"
)

const (
	_previewSize = 4096      // bytes to read for challenge detection
	_downloadBuf = 32 * 1024 // copy buffer size
)

// fetchURL makes an HTTP GET to url, applying optional cookies and a custom
// user-agent (used when retrying after a bypass solve).
func (c *Client) fetchURL(ctx context.Context, url string, headers map[string]string, cookies []*http.Cookie, userAgent string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	ua := "Bookaneer/1.0"
	if userAgent != "" {
		ua = userAgent
	}
	req.Header.Set("User-Agent", ua)

	for k, v := range headers {
		req.Header.Set(k, v)
	}
	for _, ck := range cookies {
		req.AddCookie(ck)
	}

	return c.httpClient.Do(req)
}

// challengeMessage returns a human-readable failure message for a given challenge reason.
func challengeMessage(reason string) string {
	switch reason {
	case "cloudflare":
		return "Cloudflare challenge detected — set flareSolverrUrl in config.yaml or BOOKANEER_FLARESOLVERR_URL"
	case "ddosguard":
		return "DDoS-Guard challenge detected — set flareSolverrUrl in config.yaml or BOOKANEER_FLARESOLVERR_URL"
	default:
		return "login required — set flareSolverrUrl in config.yaml or try a different source"
	}
}

// downloadFile performs the actual HTTP download with optional bypass support.
func (c *Client) downloadFile(ctx context.Context, dl *downloadItem, headers map[string]string) {
	c.updateStatus(dl.id, download.StatusDownloading, "")

	// --- Attempt 1: plain HTTP GET ---
	resp, err := c.fetchURL(ctx, dl.url, headers, nil, "")
	if err != nil {
		c.updateStatus(dl.id, download.StatusFailed, fmt.Sprintf("download failed: %v", err))
		return
	}

	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		c.updateStatus(dl.id, download.StatusFailed, fmt.Sprintf("HTTP %d: %s", resp.StatusCode, resp.Status))
		return
	}

	// Read a small preview to detect challenge/HTML pages before writing to disk.
	preview := make([]byte, _previewSize)
	n, _ := io.ReadFull(resp.Body, preview)
	preview = preview[:n]

	if challenge.IsHTML(resp.Header.Get("Content-Type")) {
		_ = resp.Body.Close()

		found, reason := challenge.Detect(string(preview))
		if !found {
			c.updateStatus(dl.id, download.StatusFailed,
				"received HTML response — source may require login")
			return
		}

		// --- Challenge detected: attempt bypass ---
		if c.cfg.Bypasser == nil || !c.cfg.Bypasser.Enabled() {
			c.updateStatus(dl.id, download.StatusFailed, challengeMessage(reason))
			return
		}

		bypassResult, bypassErr := c.cfg.Bypasser.Solve(ctx, dl.url)
		if bypassErr != nil {
			if errors.Is(bypassErr, bypass.ErrUnsolvable) {
				c.updateStatus(dl.id, download.StatusFailed,
					fmt.Sprintf("bypass could not solve challenge: %v", bypassErr))
			} else {
				c.updateStatus(dl.id, download.StatusFailed,
					fmt.Sprintf("bypass error: %v", bypassErr))
			}
			return
		}

		// --- Attempt 2: retry with bypass session cookies ---
		resp, err = c.fetchURL(ctx, dl.url, headers, bypassResult.Cookies, bypassResult.UserAgent)
		if err != nil {
			c.updateStatus(dl.id, download.StatusFailed,
				fmt.Sprintf("download failed after bypass: %v", err))
			return
		}
		if resp.StatusCode != http.StatusOK {
			_ = resp.Body.Close()
			c.updateStatus(dl.id, download.StatusFailed,
				fmt.Sprintf("HTTP %d after bypass: %s", resp.StatusCode, resp.Status))
			return
		}
		// Reset preview — second response should be the actual file.
		preview = preview[:0]
	}

	defer func() { _ = resp.Body.Close() }()

	// Update file size from whichever response we ended up with.
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

	outFile, err := os.Create(filePath) //nolint:gosec // path is constructed from trusted config root
	if err != nil {
		c.updateStatus(dl.id, download.StatusFailed, fmt.Sprintf("cannot create file: %v", err))
		return
	}
	defer func() { _ = outFile.Close() }()

	// Update the saved path so status queries return the correct path.
	c.mu.Lock()
	dl.savePath = filePath
	c.mu.Unlock()

	// Stream: replay the preview bytes already read, then the remaining body.
	fullBody := io.MultiReader(bytes.NewReader(preview), resp.Body)

	buf := make([]byte, _downloadBuf)
	var downloaded int64
	startTime := time.Now()
	lastUpdate := startTime

	for {
		nr, readErr := fullBody.Read(buf)
		if nr > 0 {
			if _, writeErr := outFile.Write(buf[:nr]); writeErr != nil {
				c.updateStatus(dl.id, download.StatusFailed, fmt.Sprintf("write error: %v", writeErr))
				return
			}
			downloaded += int64(nr)

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
