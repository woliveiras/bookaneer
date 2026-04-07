package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/woliveiras/bookaneer/internal/search"
)

// SearchHandler handles search and indexer-related HTTP requests.
type SearchHandler struct {
	svc *search.Service
}

// NewSearchHandler creates a new search handler.
func NewSearchHandler(svc *search.Service) *SearchHandler {
	return &SearchHandler{svc: svc}
}

// Register registers the search routes.
func (h *SearchHandler) Register(g *echo.Group) {
	// Indexer options (must be before :id routes to avoid conflicts)
	g.GET("/indexer/options", h.GetOptions)
	g.PUT("/indexer/options", h.UpdateOptions)
	g.POST("/indexer/test", h.TestIndexer)

	// Indexers CRUD
	g.GET("/indexer", h.ListIndexers)
	g.GET("/indexer/:id", h.GetIndexer)
	g.POST("/indexer", h.CreateIndexer)
	g.PUT("/indexer/:id", h.UpdateIndexer)
	g.DELETE("/indexer/:id", h.DeleteIndexer)

	// Search
	g.GET("/search", h.Search)
}

// ListIndexers returns all indexers.
func (h *SearchHandler) ListIndexers(c echo.Context) error {
	indexers, err := h.svc.ListIndexers(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list indexers")
	}
	return c.JSON(http.StatusOK, indexers)
}

// GetIndexer returns an indexer by ID.
func (h *SearchHandler) GetIndexer(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid indexer id")
	}
	indexer, err := h.svc.GetIndexer(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, search.ErrIndexerNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "indexer not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get indexer")
	}
	return c.JSON(http.StatusOK, indexer)
}

// CreateIndexerRequest is the request body for creating an indexer.
type CreateIndexerRequest struct {
	Name                    string   `json:"name"`
	Type                    string   `json:"type"`
	BaseURL                 string   `json:"baseUrl"`
	APIPath                 string   `json:"apiPath"`
	APIKey                  string   `json:"apiKey"`
	Categories              string   `json:"categories"`
	Priority                int      `json:"priority"`
	Enabled                 bool     `json:"enabled"`
	EnableRSS               bool     `json:"enableRss"`
	EnableAutomaticSearch   bool     `json:"enableAutomaticSearch"`
	EnableInteractiveSearch bool     `json:"enableInteractiveSearch"`
	AdditionalParameters    string   `json:"additionalParameters"`
	MinimumSeeders          int      `json:"minimumSeeders"`
	SeedRatio               *float64 `json:"seedRatio,omitempty"`
	SeedTime                *int     `json:"seedTime,omitempty"`
}

// CreateIndexer creates a new indexer.
func (h *SearchHandler) CreateIndexer(c echo.Context) error {
	var req CreateIndexerRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.Name == "" || req.Type == "" || req.BaseURL == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name, type, and baseUrl are required")
	}
	// Set defaults
	if req.APIPath == "" {
		req.APIPath = "/api"
	}
	if req.Priority == 0 {
		req.Priority = 25
	}

	cfg := search.IndexerConfig{
		Name:                    req.Name,
		Type:                    req.Type,
		BaseURL:                 req.BaseURL,
		APIPath:                 req.APIPath,
		APIKey:                  req.APIKey,
		Categories:              req.Categories,
		Priority:                req.Priority,
		Enabled:                 req.Enabled,
		EnableRSS:               req.EnableRSS,
		EnableAutomaticSearch:   req.EnableAutomaticSearch,
		EnableInteractiveSearch: req.EnableInteractiveSearch,
		AdditionalParameters:    req.AdditionalParameters,
		MinimumSeeders:          req.MinimumSeeders,
		SeedRatio:               req.SeedRatio,
		SeedTime:                req.SeedTime,
	}

	id, err := h.svc.CreateIndexer(c.Request().Context(), cfg)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create indexer")
	}
	cfg.ID = id

	return c.JSON(http.StatusCreated, cfg)
}

// UpdateIndexer updates an existing indexer.
func (h *SearchHandler) UpdateIndexer(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid indexer id")
	}

	var req CreateIndexerRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.Name == "" || req.Type == "" || req.BaseURL == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name, type, and baseUrl are required")
	}
	// Set defaults
	if req.APIPath == "" {
		req.APIPath = "/api"
	}

	cfg := search.IndexerConfig{
		ID:                      id,
		Name:                    req.Name,
		Type:                    req.Type,
		BaseURL:                 req.BaseURL,
		APIPath:                 req.APIPath,
		APIKey:                  req.APIKey,
		Categories:              req.Categories,
		Priority:                req.Priority,
		Enabled:                 req.Enabled,
		EnableRSS:               req.EnableRSS,
		EnableAutomaticSearch:   req.EnableAutomaticSearch,
		EnableInteractiveSearch: req.EnableInteractiveSearch,
		AdditionalParameters:    req.AdditionalParameters,
		MinimumSeeders:          req.MinimumSeeders,
		SeedRatio:               req.SeedRatio,
		SeedTime:                req.SeedTime,
	}

	if err := h.svc.UpdateIndexer(c.Request().Context(), cfg); err != nil {
		if errors.Is(err, search.ErrIndexerNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "indexer not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update indexer")
	}

	return c.JSON(http.StatusOK, cfg)
}

// DeleteIndexer deletes an indexer.
func (h *SearchHandler) DeleteIndexer(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid indexer id")
	}
	if err := h.svc.DeleteIndexer(c.Request().Context(), id); err != nil {
		if errors.Is(err, search.ErrIndexerNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "indexer not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete indexer")
	}
	return c.NoContent(http.StatusNoContent)
}

// TestIndexer tests an indexer configuration.
func (h *SearchHandler) TestIndexer(c echo.Context) error {
	var req CreateIndexerRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.Type == "" || req.BaseURL == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "type and baseUrl are required")
	}
	// Set defaults
	if req.APIPath == "" {
		req.APIPath = "/api"
	}

	cfg := search.IndexerConfig{
		Name:    req.Name,
		Type:    req.Type,
		BaseURL: req.BaseURL,
		APIPath: req.APIPath,
		APIKey:  req.APIKey,
	}

	if err := h.svc.TestIndexer(c.Request().Context(), cfg); err != nil {
		return c.JSON(http.StatusOK, map[string]any{
			"success": false,
			"message": err.Error(),
		})
	}
	return c.JSON(http.StatusOK, map[string]any{
		"success": true,
		"message": "Connection successful",
	})
}

// Search searches all enabled indexers.
func (h *SearchHandler) Search(c echo.Context) error {
	query := search.SearchQuery{
		Query:  c.QueryParam("q"),
		Author: c.QueryParam("author"),
		Title:  c.QueryParam("title"),
		ISBN:   c.QueryParam("isbn"),
	}
	if cat := c.QueryParam("category"); cat != "" {
		query.Category = []string{cat}
	}
	if l := c.QueryParam("limit"); l != "" {
		if limit, err := strconv.Atoi(l); err == nil {
			query.Limit = limit
		}
	}
	if o := c.QueryParam("offset"); o != "" {
		if offset, err := strconv.Atoi(o); err == nil {
			query.Offset = offset
		}
	}

	results, err := h.svc.Search(c.Request().Context(), query)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]any{
		"results": results,
		"total":   len(results),
	})
}

// GetOptions returns global indexer options.
func (h *SearchHandler) GetOptions(c echo.Context) error {
	opts, err := h.svc.GetOptions(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get indexer options")
	}
	return c.JSON(http.StatusOK, opts)
}

// UpdateOptionsRequest is the request body for updating indexer options.
type UpdateOptionsRequest struct {
	MinimumAge         int  `json:"minimumAge"`
	Retention          int  `json:"retention"`
	MaximumSize        int  `json:"maximumSize"`
	RSSSyncInterval    int  `json:"rssSyncInterval"`
	PreferIndexerFlags bool `json:"preferIndexerFlags"`
	AvailabilityDelay  int  `json:"availabilityDelay"`
}

// UpdateOptions updates global indexer options.
func (h *SearchHandler) UpdateOptions(c echo.Context) error {
	var req UpdateOptionsRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	opts := search.IndexerOptions{
		MinimumAge:         req.MinimumAge,
		Retention:          req.Retention,
		MaximumSize:        req.MaximumSize,
		RSSSyncInterval:    req.RSSSyncInterval,
		PreferIndexerFlags: req.PreferIndexerFlags,
		AvailabilityDelay:  req.AvailabilityDelay,
	}

	if err := h.svc.UpdateOptions(c.Request().Context(), opts); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update indexer options")
	}

	// Return updated options
	updated, err := h.svc.GetOptions(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get updated options")
	}
	return c.JSON(http.StatusOK, updated)
}
