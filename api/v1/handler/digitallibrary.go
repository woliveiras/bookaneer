package handler

import (
	"net/http"

	"github.com/labstack/echo/v5"

	"github.com/woliveiras/bookaneer/internal/library"
	"github.com/woliveiras/bookaneer/internal/search"
)

// DigitalLibraryHandler handles digital library search HTTP requests.
// It searches across Anna's Archive, LibGen, Internet Archive, etc.
type DigitalLibraryHandler struct {
	agg *library.Aggregator
}

// NewDigitalLibraryHandler creates a new digital library handler.
func NewDigitalLibraryHandler(agg *library.Aggregator) *DigitalLibraryHandler {
	return &DigitalLibraryHandler{agg: agg}
}

// Register registers the digital library routes.
func (h *DigitalLibraryHandler) Register(g *echo.Group) {
	g.GET("/digitallibrary/search", h.Search)
	g.GET("/digitallibrary/providers", h.Providers)
}

// DigitalLibrarySearchResponse represents the library search response.
type DigitalLibrarySearchResponse struct {
	Results      []library.SearchResult `json:"results"`
	Total        int                    `json:"total"`
	ColumnConfig search.ColumnConfig    `json:"columnConfig"`
}

// Search searches all digital library providers.
func (h *DigitalLibraryHandler) Search(c *echo.Context) error {
	query := c.QueryParam("q")
	if query == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "query parameter 'q' is required")
	}

	results, err := h.agg.Search(c.Request().Context(), query)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "search failed")
	}

	return c.JSON(http.StatusOK, DigitalLibrarySearchResponse{
		Results:      results,
		Total:        len(results),
		ColumnConfig: search.LibraryColumnConfig(),
	})
}

// DigitalLibraryProvidersResponse lists available digital library providers.
type DigitalLibraryProvidersResponse struct {
	Providers []string `json:"providers"`
}

// Providers returns the list of configured digital library providers.
func (h *DigitalLibraryHandler) Providers(c *echo.Context) error {
	return c.JSON(http.StatusOK, DigitalLibraryProvidersResponse{
		Providers: h.agg.Providers(),
	})
}
