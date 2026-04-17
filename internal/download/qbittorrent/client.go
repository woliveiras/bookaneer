package qbittorrent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/woliveiras/bookaneer/internal/download"
)

func init() {
	download.RegisterFactory(download.ClientTypeQBittorrent, func(cfg download.ClientConfig) (download.Client, error) {
		jar, _ := cookiejar.New(nil)
		httpClient := &http.Client{
			Timeout: 30 * time.Second,
			Jar:     jar,
		}
		return &Client{
			cfg: Config{
				Name:     cfg.Name,
				Host:     cfg.Host,
				Port:     cfg.Port,
				UseTLS:   cfg.UseTLS,
				Username: cfg.Username,
				Password: cfg.Password,
				Category: cfg.Category,
			},
			client: httpClient,
		}, nil
	})
}

// Config holds qBittorrent client configuration.
type Config struct {
	Name     string
	Host     string
	Port     int
	UseTLS   bool
	Username string
	Password string
	Category string
}

// Client is a qBittorrent Web API client.
type Client struct {
	cfg    Config
	client *http.Client
}

// Name returns the client name.
func (c *Client) Name() string {
	return c.cfg.Name
}

// Type returns the client type.
func (c *Client) Type() string {
	return download.ClientTypeQBittorrent
}

// Test tests the connection to qBittorrent.
func (c *Client) Test(ctx context.Context) error {
	return c.login(ctx)
}

// Add adds a torrent to qBittorrent.
func (c *Client) Add(ctx context.Context, item download.AddItem) (string, error) {
	if err := c.login(ctx); err != nil {
		return "", err
	}

	data := url.Values{}
	data.Set("urls", item.DownloadURL)
	if item.Category != "" {
		data.Set("category", item.Category)
	}
	if item.SavePath != "" {
		data.Set("savepath", item.SavePath)
	}
	if len(item.Tags) > 0 {
		data.Set("tags", strings.Join(item.Tags, ","))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/api/v2/torrents/add"), strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %v", download.ErrConnectionFailed, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("add torrent failed: status %d", resp.StatusCode)
	}

	// qBittorrent doesn't return the hash on add, we'd need to query for it
	// For now, return a timestamp-based ID
	return fmt.Sprintf("qbt_%d", time.Now().UnixNano()), nil
}

// Remove removes a torrent from qBittorrent.
func (c *Client) Remove(ctx context.Context, id string, deleteData bool) error {
	if err := c.login(ctx); err != nil {
		return err
	}

	data := url.Values{}
	data.Set("hashes", id)
	data.Set("deleteFiles", fmt.Sprintf("%t", deleteData))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/api/v2/torrents/delete"), strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", download.ErrConnectionFailed, err)
	}
	defer func() { _ = resp.Body.Close() }()

	return nil
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

// GetQueue returns all torrents in qBittorrent.
func (c *Client) GetQueue(ctx context.Context) ([]download.ItemStatus, error) {
	if err := c.login(ctx); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url("/api/v2/torrents/info"), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", download.ErrConnectionFailed, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get queue failed: status %d", resp.StatusCode)
	}

	var torrents []torrentInfo
	if err := json.NewDecoder(resp.Body).Decode(&torrents); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	items := make([]download.ItemStatus, 0, len(torrents))
	for _, t := range torrents {
		items = append(items, download.ItemStatus{
			ID:             t.Hash,
			Name:           t.Name,
			Status:         mapQbtStatus(t.State),
			Progress:       t.Progress * 100,
			Size:           t.Size,
			DownloadedSize: t.Downloaded,
			Speed:          t.DlSpeed,
			ETA:            time.Duration(t.ETA) * time.Second,
			Seeders:        t.NumSeeds,
			Leechers:       t.NumLeechs,
			Ratio:          t.Ratio,
			SavePath:       t.SavePath,
			Category:       t.Category,
			AddedAt:        time.Unix(t.AddedOn, 0),
		})
	}

	return items, nil
}

func (c *Client) login(ctx context.Context) error {
	data := url.Values{}
	data.Set("username", c.cfg.Username)
	data.Set("password", c.cfg.Password)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/api/v2/auth/login"), strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", download.ErrConnectionFailed, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK || !strings.Contains(string(body), "Ok") {
		return download.ErrAuthFailed
	}

	return nil
}

func (c *Client) url(path string) string {
	scheme := "http"
	if c.cfg.UseTLS {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s:%d%s", scheme, c.cfg.Host, c.cfg.Port, path)
}

type torrentInfo struct {
	Hash       string  `json:"hash"`
	Name       string  `json:"name"`
	State      string  `json:"state"`
	Progress   float64 `json:"progress"`
	Size       int64   `json:"size"`
	Downloaded int64   `json:"downloaded"`
	DlSpeed    int64   `json:"dlspeed"`
	ETA        int64   `json:"eta"`
	NumSeeds   int     `json:"num_seeds"`
	NumLeechs  int     `json:"num_leechs"`
	Ratio      float64 `json:"ratio"`
	SavePath   string  `json:"save_path"`
	Category   string  `json:"category"`
	AddedOn    int64   `json:"added_on"`
}

func mapQbtStatus(state string) download.Status {
	switch state {
	case "downloading", "metaDL", "forcedDL":
		return download.StatusDownloading
	case "pausedDL", "pausedUP":
		return download.StatusPaused
	case "uploading", "forcedUP", "stalledUP":
		return download.StatusSeeding
	case "stalledDL", "checkingDL", "checkingUP", "queuedDL", "queuedUP":
		return download.StatusQueued
	case "error", "missingFiles":
		return download.StatusFailed
	default:
		return download.StatusQueued
	}
}
