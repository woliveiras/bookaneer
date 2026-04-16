// Package flaresolverr provides an HTTP client that routes requests through
// a FlareSolverr instance to bypass Cloudflare challenges.
// See https://github.com/FlareSolverr/FlareSolverr
package flaresolverr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type request struct {
	CMD        string `json:"cmd"`
	URL        string `json:"url"`
	MaxTimeout int    `json:"maxTimeout"`
}

type solutionResponse struct {
	Status   string `json:"status"`
	Message  string `json:"message"`
	Solution struct {
		URL      string            `json:"url"`
		Status   int               `json:"status"`
		Headers  map[string]string `json:"headers"`
		Response string            `json:"response"`
	} `json:"solution"`
}

// Client sends requests through a FlareSolverr proxy.
type Client struct {
	solverrURL string
	httpClient *http.Client
}

// New creates a client that proxies through the given FlareSolverr URL.
func New(solverrURL string) *Client {
	return &Client{
		solverrURL: solverrURL,
		httpClient: &http.Client{Timeout: 90 * time.Second},
	}
}

// Get fetches a URL via FlareSolverr and returns the response body.
func (c *Client) Get(ctx context.Context, targetURL string) ([]byte, error) {
	payload := request{
		CMD:        "request.get",
		URL:        targetURL,
		MaxTimeout: 60000,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.solverrURL+"/v1", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("flaresolverr returned status %d", resp.StatusCode)
	}

	var result solutionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if result.Status != "ok" {
		return nil, fmt.Errorf("flaresolverr error: %s", result.Message)
	}

	return io.ReadAll(bytes.NewBufferString(result.Solution.Response))
}
