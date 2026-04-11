package blackhole

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/woliveiras/bookaneer/internal/download"
)

func init() {
	download.RegisterFactory(download.ClientTypeBlackhole, func(cfg download.ClientConfig) (download.Client, error) {
		return &Client{
			cfg: Config{
				Name:          cfg.Name,
				NzbFolder:     cfg.NzbFolder,
				TorrentFolder: cfg.TorrentFolder,
				WatchFolder:   cfg.WatchFolder,
			},
			client: &http.Client{Timeout: 60 * time.Second},
		}, nil
	})
}

// Config holds Blackhole client configuration.
type Config struct {
	Name          string
	NzbFolder     string
	TorrentFolder string
	WatchFolder   string
}

// Client is a Blackhole download client that saves files to a folder.
type Client struct {
	cfg    Config
	client *http.Client
}

// New creates a new Blackhole client.
func New(cfg Config) *Client {
	return &Client{
		cfg: cfg,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Name returns the client name.
func (c *Client) Name() string {
	return c.cfg.Name
}

// Type returns the client type.
func (c *Client) Type() string {
	return download.ClientTypeBlackhole
}

// Test tests the blackhole folders exist and are writable.
func (c *Client) Test(ctx context.Context) error {
	folders := []string{c.cfg.NzbFolder, c.cfg.TorrentFolder, c.cfg.WatchFolder}
	for _, folder := range folders {
		if folder == "" {
			continue
		}
		if err := os.MkdirAll(folder, 0755); err != nil {
			return fmt.Errorf("cannot create folder %s: %w", folder, err)
		}
		testFile := filepath.Join(folder, ".bookaneer_test")
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			return fmt.Errorf("cannot write to folder %s: %w", folder, err)
		}
		_ = os.Remove(testFile)
	}
	return nil
}

// Add downloads a file and saves it to the appropriate folder.
func (c *Client) Add(ctx context.Context, item download.AddItem) (string, error) {
	// Determine the target folder based on URL
	var targetFolder string
	if strings.HasSuffix(strings.ToLower(item.DownloadURL), ".nzb") || strings.Contains(item.DownloadURL, "nzb") {
		targetFolder = c.cfg.NzbFolder
		if targetFolder == "" {
			targetFolder = c.cfg.WatchFolder
		}
	} else if strings.HasSuffix(strings.ToLower(item.DownloadURL), ".torrent") || strings.Contains(item.DownloadURL, "torrent") {
		targetFolder = c.cfg.TorrentFolder
		if targetFolder == "" {
			targetFolder = c.cfg.WatchFolder
		}
	} else {
		targetFolder = c.cfg.WatchFolder
	}

	if targetFolder == "" {
		return "", fmt.Errorf("no target folder configured")
	}

	// Download the file
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, item.DownloadURL, nil)
	if err != nil {
		return "", err
	}
	for k, v := range item.Headers {
		req.Header.Set(k, v)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %v", download.ErrConnectionFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed: status %d", resp.StatusCode)
	}

	// Determine filename
	filename := item.Name
	if filename == "" {
		filename = filepath.Base(item.DownloadURL)
	}
	// Ensure proper extension
	if strings.Contains(item.DownloadURL, "nzb") && !strings.HasSuffix(filename, ".nzb") {
		filename += ".nzb"
	} else if strings.Contains(item.DownloadURL, "torrent") && !strings.HasSuffix(filename, ".torrent") {
		filename += ".torrent"
	}

	// Sanitize filename
	filename = sanitizeFilename(filename)

	targetPath := filepath.Join(targetFolder, filename)

	// Create target folder if needed
	if err := os.MkdirAll(targetFolder, 0755); err != nil {
		return "", fmt.Errorf("create folder: %w", err)
	}

	// Write file
	f, err := os.Create(targetPath)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer func() { _ = f.Close() }()

	if _, err := io.Copy(f, resp.Body); err != nil {
		_ = os.Remove(targetPath)
		return "", fmt.Errorf("write file: %w", err)
	}

	return fmt.Sprintf("bh_%d_%s", time.Now().UnixNano(), filename), nil
}

// Remove is a no-op for blackhole - files are managed by other applications.
func (c *Client) Remove(ctx context.Context, id string, deleteData bool) error {
	// Blackhole doesn't track files, so this is a no-op
	return nil
}

// GetStatus always returns not found for blackhole.
func (c *Client) GetStatus(ctx context.Context, id string) (*download.ItemStatus, error) {
	// Blackhole doesn't track downloads
	return nil, download.ErrNotFound
}

// GetQueue returns an empty queue for blackhole.
func (c *Client) GetQueue(ctx context.Context) ([]download.ItemStatus, error) {
	// Blackhole doesn't maintain a queue - files are dropped and forgotten
	return nil, nil
}

func sanitizeFilename(name string) string {
	// Replace invalid characters
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := name
	for _, char := range invalid {
		result = strings.ReplaceAll(result, char, "_")
	}
	// Limit length
	if len(result) > 200 {
		ext := filepath.Ext(result)
		result = result[:200-len(ext)] + ext
	}
	return result
}
