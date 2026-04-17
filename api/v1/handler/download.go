package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"

	"github.com/woliveiras/bookaneer/internal/download"
)

// DownloadHandler handles download client related HTTP requests.
type DownloadHandler struct {
	svc *download.Service
}

// NewDownloadHandler creates a new download handler.
func NewDownloadHandler(svc *download.Service) *DownloadHandler {
	return &DownloadHandler{svc: svc}
}

// Register registers the download routes.
func (h *DownloadHandler) Register(g *echo.Group) {
	// Download clients CRUD
	g.GET("/downloadclient", h.ListClients)
	g.GET("/downloadclient/:id", h.GetClient)
	g.POST("/downloadclient", h.CreateClient)
	g.PUT("/downloadclient/:id", h.UpdateClient)
	g.DELETE("/downloadclient/:id", h.DeleteClient)
	g.POST("/downloadclient/test", h.TestClient)

	// Queue
	g.GET("/downloadclient/queue", h.GetQueue)
	g.GET("/downloadclient/queue/:clientId", h.GetClientQueue)
}

// ListClients returns all download clients.
func (h *DownloadHandler) ListClients(c *echo.Context) error {
	clients, err := h.svc.ListClients(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list download clients")
	}
	return c.JSON(http.StatusOK, clients)
}

// GetClient returns a download client by ID.
func (h *DownloadHandler) GetClient(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid client id")
	}

	client, err := h.svc.GetClient(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, download.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "download client not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get download client")
	}
	return c.JSON(http.StatusOK, client)
}

// CreateClientRequest is the request body for creating a download client.
type CreateClientRequest struct {
	Name                 string            `json:"name"`
	Type                 string            `json:"type"`
	Host                 string            `json:"host"`
	Port                 int               `json:"port"`
	UseTLS               bool              `json:"useTls"`
	Username             string            `json:"username"`
	Password             string            `json:"password"`
	APIKey               string            `json:"apiKey"`
	Category             string            `json:"category"`
	RecentPriority       download.Priority `json:"recentPriority"`
	OlderPriority        download.Priority `json:"olderPriority"`
	RemoveCompletedAfter int               `json:"removeCompletedAfter"`
	Enabled              bool              `json:"enabled"`
	Priority             int               `json:"priority"`
	NzbFolder            string            `json:"nzbFolder"`
	TorrentFolder        string            `json:"torrentFolder"`
	WatchFolder          string            `json:"watchFolder"`
}

// CreateClient creates a new download client.
func (h *DownloadHandler) CreateClient(c *echo.Context) error {
	var req CreateClientRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if req.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name is required")
	}
	if req.Type == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "type is required")
	}

	cfg := &download.ClientConfig{
		Name:                 req.Name,
		Type:                 req.Type,
		Host:                 req.Host,
		Port:                 req.Port,
		UseTLS:               req.UseTLS,
		Username:             req.Username,
		Password:             req.Password,
		APIKey:               req.APIKey,
		Category:             req.Category,
		RecentPriority:       req.RecentPriority,
		OlderPriority:        req.OlderPriority,
		RemoveCompletedAfter: req.RemoveCompletedAfter,
		Enabled:              req.Enabled,
		Priority:             req.Priority,
		NzbFolder:            req.NzbFolder,
		TorrentFolder:        req.TorrentFolder,
		WatchFolder:          req.WatchFolder,
	}

	if err := h.svc.CreateClient(c.Request().Context(), cfg); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create download client")
	}
	return c.JSON(http.StatusCreated, cfg)
}

// UpdateClient updates an existing download client.
func (h *DownloadHandler) UpdateClient(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid client id")
	}

	var req CreateClientRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	cfg := &download.ClientConfig{
		ID:                   id,
		Name:                 req.Name,
		Type:                 req.Type,
		Host:                 req.Host,
		Port:                 req.Port,
		UseTLS:               req.UseTLS,
		Username:             req.Username,
		Password:             req.Password,
		APIKey:               req.APIKey,
		Category:             req.Category,
		RecentPriority:       req.RecentPriority,
		OlderPriority:        req.OlderPriority,
		RemoveCompletedAfter: req.RemoveCompletedAfter,
		Enabled:              req.Enabled,
		Priority:             req.Priority,
		NzbFolder:            req.NzbFolder,
		TorrentFolder:        req.TorrentFolder,
		WatchFolder:          req.WatchFolder,
	}

	if err := h.svc.UpdateClient(c.Request().Context(), cfg); err != nil {
		if errors.Is(err, download.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "download client not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update download client")
	}
	return c.JSON(http.StatusOK, cfg)
}

// DeleteClient deletes a download client.
func (h *DownloadHandler) DeleteClient(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid client id")
	}

	if err := h.svc.DeleteClient(c.Request().Context(), id); err != nil {
		if errors.Is(err, download.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "download client not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete download client")
	}
	return c.NoContent(http.StatusNoContent)
}

// TestClientRequest is the request body for testing a download client.
type TestClientRequest struct {
	Name          string `json:"name"`
	Type          string `json:"type"`
	Host          string `json:"host"`
	Port          int    `json:"port"`
	UseTLS        bool   `json:"useTls"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	APIKey        string `json:"apiKey"`
	NzbFolder     string `json:"nzbFolder"`
	TorrentFolder string `json:"torrentFolder"`
	WatchFolder   string `json:"watchFolder"`
}

// TestClient tests connectivity to a download client.
func (h *DownloadHandler) TestClient(c *echo.Context) error {
	var req TestClientRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	cfg := &download.ClientConfig{
		Name:          req.Name,
		Type:          req.Type,
		Host:          req.Host,
		Port:          req.Port,
		UseTLS:        req.UseTLS,
		Username:      req.Username,
		Password:      req.Password,
		APIKey:        req.APIKey,
		NzbFolder:     req.NzbFolder,
		TorrentFolder: req.TorrentFolder,
		WatchFolder:   req.WatchFolder,
	}

	if err := h.svc.TestClient(c.Request().Context(), cfg); err != nil {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Connection successful",
	})
}

// GetQueue returns the combined queue from all enabled clients.
func (h *DownloadHandler) GetQueue(c *echo.Context) error {
	items, err := h.svc.GetQueue(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get queue")
	}
	return c.JSON(http.StatusOK, items)
}

// GetClientQueue returns the queue from a specific client.
func (h *DownloadHandler) GetClientQueue(c *echo.Context) error {
	clientID, err := strconv.ParseInt(c.Param("clientId"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid client id")
	}

	items, err := h.svc.GetClientQueue(c.Request().Context(), clientID)
	if err != nil {
		if errors.Is(err, download.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "download client not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get client queue")
	}
	return c.JSON(http.StatusOK, items)
}

