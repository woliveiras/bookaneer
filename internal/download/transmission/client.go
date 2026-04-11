package transmission

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/woliveiras/bookaneer/internal/download"
)

func init() {
	download.RegisterFactory(download.ClientTypeTransmission, func(cfg download.ClientConfig) (download.Client, error) {
		return &Client{
			cfg: Config{
				Name:     cfg.Name,
				Host:     cfg.Host,
				Port:     cfg.Port,
				UseTLS:   cfg.UseTLS,
				Username: cfg.Username,
				Password: cfg.Password,
			},
			client: &http.Client{Timeout: 30 * time.Second},
		}, nil
	})
}

// Config holds Transmission client configuration.
type Config struct {
	Name     string
	Host     string
	Port     int
	UseTLS   bool
	Username string
	Password string
}

// Client is a Transmission RPC client.
type Client struct {
	cfg       Config
	client    *http.Client
	sessionID string
	mu        sync.RWMutex
}

// New creates a new Transmission client.
func New(cfg Config, httpClient *http.Client) *Client {
	return &Client{
		cfg:    cfg,
		client: httpClient,
	}
}

// Name returns the client name.
func (c *Client) Name() string {
	return c.cfg.Name
}

// Type returns the client type.
func (c *Client) Type() string {
	return download.ClientTypeTransmission
}

// Test tests the connection to Transmission.
func (c *Client) Test(ctx context.Context) error {
	_, err := c.rpc(ctx, "session-get", nil)
	return err
}

// Add adds a torrent to Transmission.
func (c *Client) Add(ctx context.Context, item download.AddItem) (string, error) {
	args := map[string]interface{}{
		"filename": item.DownloadURL,
	}
	if item.SavePath != "" {
		args["download-dir"] = item.SavePath
	}

	result, err := c.rpc(ctx, "torrent-add", args)
	if err != nil {
		return "", err
	}

	if torrentAdded, ok := result["torrent-added"].(map[string]interface{}); ok {
		if hashString, ok := torrentAdded["hashString"].(string); ok {
			return hashString, nil
		}
		if id, ok := torrentAdded["id"].(float64); ok {
			return fmt.Sprintf("%d", int64(id)), nil
		}
	}

	if torrentDupe, ok := result["torrent-duplicate"].(map[string]interface{}); ok {
		if hashString, ok := torrentDupe["hashString"].(string); ok {
			return hashString, nil
		}
	}

	return fmt.Sprintf("tr_%d", time.Now().UnixNano()), nil
}

// Remove removes a torrent from Transmission.
func (c *Client) Remove(ctx context.Context, id string, deleteData bool) error {
	args := map[string]interface{}{
		"ids":               []string{id},
		"delete-local-data": deleteData,
	}
	_, err := c.rpc(ctx, "torrent-remove", args)
	return err
}

// GetStatus returns the status of a torrent.
func (c *Client) GetStatus(ctx context.Context, id string) (*download.ItemStatus, error) {
	queue, err := c.GetQueue(ctx)
	if err != nil {
		return nil, err
	}

	for _, item := range queue {
		if item.ID == id {
			return &item, nil
		}
	}

	return nil, download.ErrNotFound
}

// GetQueue returns all torrents in Transmission.
func (c *Client) GetQueue(ctx context.Context) ([]download.ItemStatus, error) {
	args := map[string]interface{}{
		"fields": []string{
			"id", "hashString", "name", "status", "percentDone", "totalSize",
			"downloadedEver", "rateDownload", "eta", "seeders", "leechers",
			"uploadRatio", "downloadDir", "addedDate", "errorString",
		},
	}

	result, err := c.rpc(ctx, "torrent-get", args)
	if err != nil {
		return nil, err
	}

	torrentsRaw, ok := result["torrents"].([]interface{})
	if !ok {
		return nil, nil
	}

	items := make([]download.ItemStatus, 0, len(torrentsRaw))
	for _, t := range torrentsRaw {
		torrent, ok := t.(map[string]interface{})
		if !ok {
			continue
		}

		item := download.ItemStatus{
			ID:   getStringOrInt(torrent, "hashString", "id"),
			Name: getString(torrent, "name"),
		}

		if status, ok := torrent["status"].(float64); ok {
			item.Status = mapTrStatus(int(status))
		}
		if pct, ok := torrent["percentDone"].(float64); ok {
			item.Progress = pct * 100
		}
		if size, ok := torrent["totalSize"].(float64); ok {
			item.Size = int64(size)
		}
		if dl, ok := torrent["downloadedEver"].(float64); ok {
			item.DownloadedSize = int64(dl)
		}
		if rate, ok := torrent["rateDownload"].(float64); ok {
			item.Speed = int64(rate)
		}
		if eta, ok := torrent["eta"].(float64); ok && eta > 0 {
			item.ETA = time.Duration(eta) * time.Second
		}
		if ratio, ok := torrent["uploadRatio"].(float64); ok {
			item.Ratio = ratio
		}
		item.SavePath = getString(torrent, "downloadDir")
		item.ErrorMessage = getString(torrent, "errorString")

		if added, ok := torrent["addedDate"].(float64); ok && added > 0 {
			item.AddedAt = time.Unix(int64(added), 0)
		}

		items = append(items, item)
	}

	return items, nil
}

func (c *Client) rpc(ctx context.Context, method string, args map[string]interface{}) (map[string]interface{}, error) {
	body := map[string]interface{}{
		"method": method,
	}
	if args != nil {
		body["arguments"] = args
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	// First request might fail with 409, which gives us the session ID
	for tries := 0; tries < 2; tries++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url(), bytes.NewReader(data))
		if err != nil {
			return nil, err
		}

		req.Header.Set("Content-Type", "application/json")
		if c.cfg.Username != "" {
			req.SetBasicAuth(c.cfg.Username, c.cfg.Password)
		}

		c.mu.RLock()
		sessionID := c.sessionID
		c.mu.RUnlock()

		if sessionID != "" {
			req.Header.Set("X-Transmission-Session-Id", sessionID)
		}

		resp, err := c.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", download.ErrConnectionFailed, err)
		}

		if resp.StatusCode == http.StatusConflict {
			// Get new session ID
			newSessionID := resp.Header.Get("X-Transmission-Session-Id")
			_ = resp.Body.Close()
			if newSessionID != "" {
				c.mu.Lock()
				c.sessionID = newSessionID
				c.mu.Unlock()
				continue
			}
			return nil, download.ErrAuthFailed
		}

		if resp.StatusCode == http.StatusUnauthorized {
			_ = resp.Body.Close()
			return nil, download.ErrAuthFailed
		}

		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("RPC failed: status %d", resp.StatusCode)
		}

		var result struct {
			Arguments map[string]interface{} `json:"arguments"`
			Result    string                 `json:"result"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, fmt.Errorf("decode response: %w", err)
		}

		if result.Result != "success" {
			return nil, fmt.Errorf("RPC error: %s", result.Result)
		}

		return result.Arguments, nil
	}

	return nil, download.ErrConnectionFailed
}

func (c *Client) url() string {
	scheme := "http"
	if c.cfg.UseTLS {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s:%d/transmission/rpc", scheme, c.cfg.Host, c.cfg.Port)
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getStringOrInt(m map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if v, ok := m[key].(string); ok && v != "" {
			return v
		}
		if v, ok := m[key].(float64); ok {
			return fmt.Sprintf("%d", int64(v))
		}
	}
	return ""
}

// Transmission status codes
const (
	trStopped       = 0
	trCheckPending  = 1
	trChecking      = 2
	trDownloadQueue = 3
	trDownloading   = 4
	trSeedQueue     = 5
	trSeeding       = 6
)

func mapTrStatus(status int) download.Status {
	switch status {
	case trStopped:
		return download.StatusPaused
	case trCheckPending, trChecking:
		return download.StatusQueued
	case trDownloadQueue:
		return download.StatusQueued
	case trDownloading:
		return download.StatusDownloading
	case trSeedQueue:
		return download.StatusCompleted
	case trSeeding:
		return download.StatusSeeding
	default:
		return download.StatusQueued
	}
}
