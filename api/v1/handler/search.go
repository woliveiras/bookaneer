package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"

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

	// Prowlarr-compatible schema endpoint
	g.GET("/indexer/schema", h.IndexerSchema)

	// Indexers CRUD
	g.GET("/indexer", h.ListIndexers)
	g.GET("/indexer/:id", h.GetIndexer)
	g.POST("/indexer", h.CreateIndexer)
	g.PUT("/indexer/:id", h.UpdateIndexer)
	g.DELETE("/indexer/:id", h.DeleteIndexer)

	// Search
	g.GET("/search", h.Search)
}

// IndexerSchemaField represents a field in the indexer schema for Prowlarr compatibility.
type IndexerSchemaField struct {
	Order    int         `json:"order"`
	Name     string      `json:"name"`
	Label    string      `json:"label"`
	Value    interface{} `json:"value"`
	Type     string      `json:"type"`
	Advanced bool        `json:"advanced"`
	Privacy  string      `json:"privacy,omitempty"`
	IsFloat  bool        `json:"isFloat,omitempty"`
	HelpText string      `json:"helpText,omitempty"`
}

// IndexerSchema returns the schema that Prowlarr uses to sync indexers.
// This enables Prowlarr to add Bookaneer as an application and sync indexers automatically.
func (h *SearchHandler) IndexerSchema(c *echo.Context) error {
	schemas := []map[string]interface{}{
		{
			"id":                      0,
			"name":                    "",
			"implementation":          "Newznab",
			"implementationName":      "Newznab",
			"configContract":          "NewznabSettings",
			"infoLink":                "https://wiki.servarr.com/readarr/supported#newznab",
			"enableRss":               true,
			"enableAutomaticSearch":   true,
			"enableInteractiveSearch": true,
			"supportsRss":             true,
			"supportsSearch":          true,
			"protocol":                "usenet",
			"priority":                25,
			"fields": []IndexerSchemaField{
				{Order: 0, Name: "baseUrl", Label: "URL", Type: "textbox", Value: ""},
				{Order: 1, Name: "apiPath", Label: "API Path", Type: "textbox", Value: "/api", Advanced: true},
				{Order: 2, Name: "apiKey", Label: "API Key", Type: "textbox", Value: "", Privacy: "apiKey"},
				{Order: 3, Name: "categories", Label: "Categories", Type: "textbox", Value: "7000,7010,7020,7030,7040,7050,7060", Advanced: true, HelpText: "Comma-separated list of categories"},
				{Order: 4, Name: "additionalParameters", Label: "Additional Parameters", Type: "textbox", Advanced: true},
			},
		},
		{
			"id":                      0,
			"name":                    "",
			"implementation":          "Torznab",
			"implementationName":      "Torznab",
			"configContract":          "TorznabSettings",
			"infoLink":                "https://wiki.servarr.com/readarr/supported#torznab",
			"enableRss":               true,
			"enableAutomaticSearch":   true,
			"enableInteractiveSearch": true,
			"supportsRss":             true,
			"supportsSearch":          true,
			"protocol":                "torrent",
			"priority":                25,
			"fields": []IndexerSchemaField{
				{Order: 0, Name: "baseUrl", Label: "URL", Type: "textbox", Value: ""},
				{Order: 1, Name: "apiPath", Label: "API Path", Type: "textbox", Value: "/api", Advanced: true},
				{Order: 2, Name: "apiKey", Label: "API Key", Type: "textbox", Value: "", Privacy: "apiKey"},
				{Order: 3, Name: "categories", Label: "Categories", Type: "textbox", Value: "7000,7010,7020,7030,7040,7050,7060", Advanced: true, HelpText: "Comma-separated list of categories"},
				{Order: 4, Name: "minimumSeeders", Label: "Minimum Seeders", Type: "number", Value: 1},
				{Order: 5, Name: "seedCriteria.seedRatio", Label: "Seed Ratio", Type: "textbox", Value: "", Advanced: true, IsFloat: true},
				{Order: 6, Name: "seedCriteria.seedTime", Label: "Seed Time", Type: "number", Value: nil, Advanced: true, HelpText: "Time in minutes"},
				{Order: 7, Name: "additionalParameters", Label: "Additional Parameters", Type: "textbox", Advanced: true},
			},
		},
	}

	return c.JSON(http.StatusOK, schemas)
}

// ListIndexers returns all indexers.
func (h *SearchHandler) ListIndexers(c *echo.Context) error {
	indexers, err := h.svc.ListIndexers(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list indexers")
	}
	return c.JSON(http.StatusOK, indexers)
}

// GetIndexer returns an indexer by ID.
func (h *SearchHandler) GetIndexer(c *echo.Context) error {
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
// Supports both Bookaneer native format and Prowlarr sync format.
type CreateIndexerRequest struct {
	// Native Bookaneer format
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

	// Prowlarr sync format (alternative way to provide fields)
	Implementation string                   `json:"implementation,omitempty"`
	Fields         []map[string]interface{} `json:"fields,omitempty"`
}

// parseFieldValue safely extracts a value from a Prowlarr field map.
func parseFieldValue(fields []map[string]interface{}, name string) interface{} {
	for _, f := range fields {
		if n, ok := f["name"].(string); ok && n == name {
			return f["value"]
		}
	}
	return nil
}

// CreateIndexer creates a new indexer.
func (h *SearchHandler) CreateIndexer(c *echo.Context) error {
	var req CreateIndexerRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	// Handle Prowlarr format (with Implementation and Fields)
	if req.Implementation != "" && len(req.Fields) > 0 {
		// Determine type from implementation
		switch req.Implementation {
		case "Newznab":
			req.Type = "newznab"
		case "Torznab":
			req.Type = "torznab"
		default:
			req.Type = "torznab"
		}

		// Extract fields
		if v, ok := parseFieldValue(req.Fields, "baseUrl").(string); ok {
			req.BaseURL = v
		}
		if v, ok := parseFieldValue(req.Fields, "apiPath").(string); ok {
			req.APIPath = v
		}
		if v, ok := parseFieldValue(req.Fields, "apiKey").(string); ok {
			req.APIKey = v
		}
		if v, ok := parseFieldValue(req.Fields, "categories").(string); ok {
			req.Categories = v
		}
		if v, ok := parseFieldValue(req.Fields, "minimumSeeders").(float64); ok {
			req.MinimumSeeders = int(v)
		}
		if v, ok := parseFieldValue(req.Fields, "seedCriteria.seedRatio").(float64); ok {
			req.SeedRatio = &v
		}
		if v, ok := parseFieldValue(req.Fields, "seedCriteria.seedTime").(float64); ok {
			t := int(v)
			req.SeedTime = &t
		}
		if v, ok := parseFieldValue(req.Fields, "additionalParameters").(string); ok {
			req.AdditionalParameters = v
		}

		// Default enabled to true for Prowlarr sync
		req.Enabled = true
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
func (h *SearchHandler) UpdateIndexer(c *echo.Context) error {
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
func (h *SearchHandler) DeleteIndexer(c *echo.Context) error {
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
func (h *SearchHandler) TestIndexer(c *echo.Context) error {
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
func (h *SearchHandler) Search(c *echo.Context) error {
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
func (h *SearchHandler) GetOptions(c *echo.Context) error {
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
func (h *SearchHandler) UpdateOptions(c *echo.Context) error {
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
