package flaresolverr_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/woliveiras/bookaneer/internal/bypass"
	"github.com/woliveiras/bookaneer/internal/bypass/flaresolverr"
)

func TestClient_Enabled(t *testing.T) {
	c := flaresolverr.New("http://localhost:8191")
	assert.True(t, c.Enabled())
}

func TestClient_Solve_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		resp := map[string]any{
			"status": "ok",
			"solution": map[string]any{
				"url":    "https://example.com/file.epub",
				"status": 200,
				"cookies": []map[string]string{
					{"name": "cf_clearance", "value": "abc123"},
					{"name": "session_id", "value": "xyz789"},
				},
				"userAgent": "Mozilla/5.0 (FlareSolverr)",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := flaresolverr.New(srv.URL)
	result, err := c.Solve(context.Background(), "https://example.com/file.epub")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "Mozilla/5.0 (FlareSolverr)", result.UserAgent)
	require.Len(t, result.Cookies, 2)
	assert.Equal(t, "cf_clearance", result.Cookies[0].Name)
	assert.Equal(t, "abc123", result.Cookies[0].Value)
	assert.Equal(t, "session_id", result.Cookies[1].Name)
}

func TestClient_Solve_FlareSolverrError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"status":  "error",
			"message": "Challenge timed out after 60s",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := flaresolverr.New(srv.URL)
	_, err := c.Solve(context.Background(), "https://example.com/file.epub")

	require.Error(t, err)
	assert.ErrorIs(t, err, bypass.ErrUnsolvable)
	assert.Contains(t, err.Error(), "Challenge timed out after 60s")
}

func TestClient_Solve_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	c := flaresolverr.New(srv.URL)
	_, err := c.Solve(context.Background(), "https://example.com/file.epub")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "503")
}

func TestClient_Solve_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("not valid json"))
	}))
	defer srv.Close()

	c := flaresolverr.New(srv.URL)
	_, err := c.Solve(context.Background(), "https://example.com/file.epub")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "decode solve response")
}
