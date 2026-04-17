// Package direct provides an internal HTTP download client for direct downloads.
// Unlike external clients (qBittorrent, SABnzbd), this downloads files directly
// using HTTP GET requests, suitable for digital library sources like Internet Archive,
// LibGen, and Anna's Archive.
package direct

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/woliveiras/bookaneer/internal/bypass"
	"github.com/woliveiras/bookaneer/internal/download"
)

func init() {
	download.RegisterFactory(download.ClientTypeDirect, func(cfg download.ClientConfig) (download.Client, error) {
		return New(Config{
			Name:                  cfg.Name,
			DownloadDir:           cfg.DownloadDir,
			Bypasser:              cfg.Bypasser,
			CertificateValidation: cfg.CertificateValidation,
		}), nil
	})
}

// Config holds Direct downloader configuration.
type Config struct {
	Name                  string
	DownloadDir           string          // Directory to save downloaded files
	Bypasser              bypass.Bypasser // Optional; nil means no bypass
	// CertificateValidation controls TLS verification: "enabled" (default), "disabled_local", "disabled".
	CertificateValidation string
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
	cfg             Config
	httpClient      *http.Client // TLS-verified client
	httpClientNoTLS *http.Client // TLS-skipped client (nil when not needed)

	mu        sync.RWMutex
	downloads map[string]*downloadItem
}

// New creates a new Direct download client.
func New(cfg Config) *Client {
	newTransport := func(skipTLS bool) *http.Transport {
		t := &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 5,
			IdleConnTimeout:     90 * time.Second,
		}
		if skipTLS {
			t.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // user-configured opt-in
		}
		return t
	}

	c := &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout:   30 * time.Minute,
			Transport: newTransport(false),
		},
		downloads: make(map[string]*downloadItem),
	}

	if cfg.CertificateValidation == "disabled" || cfg.CertificateValidation == "disabled_local" {
		c.httpClientNoTLS = &http.Client{
			Timeout:   30 * time.Minute,
			Transport: newTransport(true),
		}
	}

	return c
}

// isLocalURL reports whether the URL's host resolves to a private/local address.
func isLocalURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	host := u.Hostname()
	ip := net.ParseIP(host)
	if ip == nil {
		// Treat common local hostnames as local.
		if host == "localhost" {
			return true
		}
		addrs, err := net.LookupHost(host)
		if err != nil || len(addrs) == 0 {
			return false
		}
		ip = net.ParseIP(addrs[0])
		if ip == nil {
			return false
		}
	}
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
		"::1/128",
		"fc00::/7",
	}
	for _, cidr := range privateRanges {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(ip) {
			return true
		}
	}
	return false
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
	_ = os.Remove(testFile)

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

