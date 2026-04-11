package sabnzbd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/woliveiras/bookaneer/internal/download"
)

func init() {
	download.RegisterFactory(download.ClientTypeSABnzbd, func(cfg download.ClientConfig) (download.Client, error) {
		return &Client{
			cfg: Config{
				Name:     cfg.Name,
				Host:     cfg.Host,
				Port:     cfg.Port,
				UseTLS:   cfg.UseTLS,
				APIKey:   cfg.APIKey,
				Category: cfg.Category,
			},
			client: &http.Client{Timeout: 30 * time.Second},
		}, nil
	})
}

// Config holds SABnzbd client configuration.
type Config struct {
	Name     string
	Host     string
	Port     int
	UseTLS   bool
	APIKey   string
	Category string
}

// Client is a SABnzbd API client.
type Client struct {
	cfg    Config
	client *http.Client
}

// New creates a new SABnzbd client.
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
	return download.ClientTypeSABnzbd
}

// Test tests the connection to SABnzbd.
func (c *Client) Test(ctx context.Context) error {
	_, err := c.api(ctx, "version", nil)
	return err
}

// Add adds an NZB to SABnzbd.
func (c *Client) Add(ctx context.Context, item download.AddItem) (string, error) {
	params := url.Values{}
	params.Set("name", item.DownloadURL)
	params.Set("nzbname", item.Name)
	if item.Category != "" {
		params.Set("cat", item.Category)
	} else if c.cfg.Category != "" {
		params.Set("cat", c.cfg.Category)
	}

	switch item.Priority {
	case download.PriorityLow:
		params.Set("priority", "-1")
	case download.PriorityNormal:
		params.Set("priority", "0")
	case download.PriorityHigh:
		params.Set("priority", "1")
	case download.PriorityForced:
		params.Set("priority", "2")
	}

	result, err := c.api(ctx, "addurl", params)
	if err != nil {
		return "", err
	}

	// SABnzbd returns the nzo_id in the response
	if nzoIDs, ok := result["nzo_ids"].([]interface{}); ok && len(nzoIDs) > 0 {
		if nzoID, ok := nzoIDs[0].(string); ok {
			return nzoID, nil
		}
	}

	return fmt.Sprintf("sab_%d", time.Now().UnixNano()), nil
}

// Remove removes an NZB from SABnzbd.
func (c *Client) Remove(ctx context.Context, id string, deleteData bool) error {
	params := url.Values{}
	params.Set("value", id)
	if deleteData {
		params.Set("del_files", "1")
	}

	// Try both queue and history
	_, _ = c.api(ctx, "queue", params)

	params.Set("name", "delete")
	params.Set("value", id)
	_, _ = c.api(ctx, "history", params)

	return nil
}

// GetStatus returns the status of an NZB.
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

// GetQueue returns all items in SABnzbd queue and history.
func (c *Client) GetQueue(ctx context.Context) ([]download.ItemStatus, error) {
	var items []download.ItemStatus

	// Get queue
	queueParams := url.Values{}
	queueParams.Set("start", "0")
	queueParams.Set("limit", "100")
	queueResult, err := c.api(ctx, "queue", queueParams)
	if err != nil {
		return nil, err
	}

	if queue, ok := queueResult["queue"].(map[string]interface{}); ok {
		if slots, ok := queue["slots"].([]interface{}); ok {
			for _, s := range slots {
				slot, ok := s.(map[string]interface{})
				if !ok {
					continue
				}

				item := download.ItemStatus{
					ID:   getString(slot, "nzo_id"),
					Name: getString(slot, "filename"),
				}

				status := getString(slot, "status")
				switch status {
				case "Downloading":
					item.Status = download.StatusDownloading
				case "Paused":
					item.Status = download.StatusPaused
				case "Queued":
					item.Status = download.StatusQueued
				case "Extracting", "Verifying", "Repairing":
					item.Status = download.StatusProcessing
				default:
					item.Status = download.StatusQueued
				}

				if pct, ok := slot["percentage"].(string); ok {
					_, _ = fmt.Sscanf(pct, "%f", &item.Progress)
				}
				if mb, ok := slot["mb"].(string); ok {
					var mbFloat float64
					_, _ = fmt.Sscanf(mb, "%f", &mbFloat)
					item.Size = int64(mbFloat * 1024 * 1024)
				}
				if mbleft, ok := slot["mbleft"].(string); ok {
					var mbleftFloat float64
					_, _ = fmt.Sscanf(mbleft, "%f", &mbleftFloat)
					item.DownloadedSize = item.Size - int64(mbleftFloat*1024*1024)
				}
				item.Category = getString(slot, "cat")

				items = append(items, item)
			}
		}
	}

	// Get history
	histParams := url.Values{}
	histParams.Set("start", "0")
	histParams.Set("limit", "50")
	histResult, err := c.api(ctx, "history", histParams)
	if err != nil {
		return items, nil // Return what we have from queue
	}

	if history, ok := histResult["history"].(map[string]interface{}); ok {
		if slots, ok := history["slots"].([]interface{}); ok {
			for _, s := range slots {
				slot, ok := s.(map[string]interface{})
				if !ok {
					continue
				}

				item := download.ItemStatus{
					ID:       getString(slot, "nzo_id"),
					Name:     getString(slot, "name"),
					Progress: 100,
				}

				status := getString(slot, "status")
				switch status {
				case "Completed":
					item.Status = download.StatusCompleted
				case "Failed":
					item.Status = download.StatusFailed
					item.ErrorMessage = getString(slot, "fail_message")
				case "Extracting":
					item.Status = download.StatusExtracted
				default:
					item.Status = download.StatusCompleted
				}

				if bytes, ok := slot["bytes"].(float64); ok {
					item.Size = int64(bytes)
					item.DownloadedSize = item.Size
				}
				item.Category = getString(slot, "category")
				item.SavePath = getString(slot, "storage")

				if completed, ok := slot["completed"].(float64); ok && completed > 0 {
					t := time.Unix(int64(completed), 0)
					item.CompletedAt = &t
				}

				items = append(items, item)
			}
		}
	}

	return items, nil
}

func (c *Client) api(ctx context.Context, mode string, params url.Values) (map[string]interface{}, error) {
	if params == nil {
		params = url.Values{}
	}
	params.Set("apikey", c.cfg.APIKey)
	params.Set("mode", mode)
	params.Set("output", "json")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url()+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", download.ErrConnectionFailed, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, download.ErrAuthFailed
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API failed: status %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	// Check for API error
	if errMsg, ok := result["error"].(string); ok && errMsg != "" {
		if errMsg == "API Key Incorrect" || errMsg == "API Key Required" {
			return nil, download.ErrAuthFailed
		}
		return nil, fmt.Errorf("API error: %s", errMsg)
	}

	return result, nil
}

func (c *Client) url() string {
	scheme := "http"
	if c.cfg.UseTLS {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s:%d/sabnzbd/api", scheme, c.cfg.Host, c.cfg.Port)
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
