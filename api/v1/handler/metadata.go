package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/woliveiras/bookaneer/internal/metadata"
)

// MetadataHandler handles metadata search endpoints.
type MetadataHandler struct {
	aggregator *metadata.Aggregator
}

// NewMetadataHandler creates a new metadata handler.
func NewMetadataHandler(aggregator *metadata.Aggregator) *MetadataHandler {
	return &MetadataHandler{aggregator: aggregator}
}

// SearchAuthorsRequest represents the request for author search.
type SearchAuthorsRequest struct {
	Query string `query:"q" validate:"required,min=1"`
}

// SearchAuthors searches for authors across all metadata providers.
// GET /api/v1/metadata/authors?q=query
func (h *MetadataHandler) SearchAuthors(c echo.Context) error {
	var req SearchAuthorsRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	if req.Query == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "query parameter 'q' is required")
	}

	results, err := h.aggregator.SearchAuthors(c.Request().Context(), req.Query)
	if err != nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "metadata search failed")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"results": results,
		"total":   len(results),
	})
}

// SearchBooksRequest represents the request for book search.
type SearchBooksRequest struct {
	Query string `query:"q" validate:"required,min=1"`
}

// SearchBooks searches for books across all metadata providers.
// GET /api/v1/metadata/books?q=query
func (h *MetadataHandler) SearchBooks(c echo.Context) error {
	var req SearchBooksRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	if req.Query == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "query parameter 'q' is required")
	}

	results, err := h.aggregator.SearchBooks(c.Request().Context(), req.Query)
	if err != nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "metadata search failed")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"results": results,
		"total":   len(results),
	})
}

// GetAuthorRequest represents the request for author details.
type GetAuthorRequest struct {
	Provider  string `query:"provider"`
	ForeignID string `param:"foreignId"`
}

// GetAuthor fetches author details from a metadata provider.
// GET /api/v1/metadata/authors/:foreignId?provider=openlibrary
func (h *MetadataHandler) GetAuthor(c echo.Context) error {
	foreignID := c.Param("foreignId")
	if foreignID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "foreignId is required")
	}

	provider := c.QueryParam("provider")

	author, err := h.aggregator.GetAuthor(c.Request().Context(), provider, foreignID)
	if err != nil {
		if err == metadata.ErrNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "author not found")
		}
		return echo.NewHTTPError(http.StatusServiceUnavailable, "metadata fetch failed")
	}

	return c.JSON(http.StatusOK, author)
}

// GetBook fetches book details from a metadata provider.
// GET /api/v1/metadata/books/:foreignId?provider=openlibrary
func (h *MetadataHandler) GetBook(c echo.Context) error {
	foreignID := c.Param("foreignId")
	if foreignID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "foreignId is required")
	}

	provider := c.QueryParam("provider")

	book, err := h.aggregator.GetBook(c.Request().Context(), provider, foreignID)
	if err != nil {
		if err == metadata.ErrNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "book not found")
		}
		return echo.NewHTTPError(http.StatusServiceUnavailable, "metadata fetch failed")
	}

	return c.JSON(http.StatusOK, book)
}

// LookupISBN fetches book details by ISBN.
// GET /api/v1/metadata/isbn/:isbn
func (h *MetadataHandler) LookupISBN(c echo.Context) error {
	isbn := c.Param("isbn")
	if isbn == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "isbn is required")
	}

	book, err := h.aggregator.GetBookByISBN(c.Request().Context(), isbn)
	if err != nil {
		if err == metadata.ErrNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "book not found")
		}
		if err == metadata.ErrInvalidISBN {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid ISBN")
		}
		return echo.NewHTTPError(http.StatusServiceUnavailable, "metadata fetch failed")
	}

	return c.JSON(http.StatusOK, book)
}

// ListProviders returns the list of available metadata providers.
// GET /api/v1/metadata/providers
func (h *MetadataHandler) ListProviders(c echo.Context) error {
	providers := h.aggregator.Providers()
	names := make([]string, 0, len(providers))
	for _, p := range providers {
		names = append(names, p.Name())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"providers": names,
	})
}
