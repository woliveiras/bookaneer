// Package flaresolverr implements bypass.Bypasser using a FlareSolverr sidecar.
// See https://github.com/FlareSolverr/FlareSolverr for setup instructions.
package flaresolverr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/woliveiras/bookaneer/internal/bypass"
)

// compile-time interface check
var _ bypass.Bypasser = (*Client)(nil)

type solveRequest struct {
	CMD        string `json:"cmd"`
	URL        string `json:"url"`
	MaxTimeout int    `json:"maxTimeout"`
}

type cookieEntry struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type solveResponse struct {
	Status   string `json:"status"`
	Message  string `json:"message"`
	Solution struct {
		URL       string            `json:"url"`
		Status    int               `json:"status"`
		Headers   map[string]string `json:"headers"`
		Response  string            `json:"response"`
		Cookies   []cookieEntry     `json:"cookies"`
		UserAgent string            `json:"userAgent"`
	} `json:"solution"`
}

// Client sends requests through a FlareSolverr proxy to bypass anti-bot challenges.
type Client struct {
	solverrURL string
	httpClient *http.Client
}

// New creates a Client that proxies through the FlareSolverr instance at solverrURL.
func New(solverrURL string) *Client {
	return &Client{
		solverrURL: solverrURL,
		// FlareSolverr can take up to 60 s to solve a challenge; allow headroom.
		httpClient: &http.Client{Timeout: 90 * time.Second},
	}
}

// Enabled always returns true — this client is only constructed when configured.
func (c *Client) Enabled() bool { return true }

// Solve sends targetURL to FlareSolverr, waits for the challenge to be resolved,
// and returns the resulting cookies and user-agent for subsequent requests.
func (c *Client) Solve(ctx context.Context, targetURL string) (*bypass.Result, error) {
	payload := solveRequest{
		CMD:        "request.get",
		URL:        targetURL,
		MaxTimeout: 60000,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal solve request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.solverrURL+"/v1", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create solve request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute solve request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("flaresolverr returned HTTP %d", resp.StatusCode)
	}

	var result solveResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode solve response: %w", err)
	}

	if result.Status != "ok" {
		return nil, fmt.Errorf("%w: %s", bypass.ErrUnsolvable, result.Message)
	}

	cookies := make([]*http.Cookie, 0, len(result.Solution.Cookies))
	for _, ce := range result.Solution.Cookies {
		cookies = append(cookies, &http.Cookie{Name: ce.Name, Value: ce.Value})
	}

	return &bypass.Result{
		Cookies:   cookies,
		UserAgent: result.Solution.UserAgent,
	}, nil
}
