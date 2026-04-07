package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/woliveiras/bookaneer/internal/core/author"
)

// AuthorHandler handles author-related HTTP requests.
type AuthorHandler struct {
	svc *author.Service
}

// NewAuthorHandler creates a new author handler.
func NewAuthorHandler(svc *author.Service) *AuthorHandler {
	return &AuthorHandler{svc: svc}
}

// Register registers the author routes.
func (h *AuthorHandler) Register(g *echo.Group) {
	g.GET("/author", h.List)
	g.GET("/author/:id", h.GetByID)
	g.POST("/author", h.Create)
	g.PUT("/author/:id", h.Update)
	g.DELETE("/author/:id", h.Delete)
	g.GET("/author/:id/stats", h.GetStats)
}

// List returns a list of authors.
func (h *AuthorHandler) List(c echo.Context) error {
	filter := author.ListAuthorsFilter{
		Status:  c.QueryParam("status"),
		Search:  c.QueryParam("search"),
		SortBy:  c.QueryParam("sortBy"),
		SortDir: c.QueryParam("sortDir"),
	}

	if m := c.QueryParam("monitored"); m != "" {
		monitored := m == "true"
		filter.Monitored = &monitored
	}
	if l := c.QueryParam("limit"); l != "" {
		if limit, err := strconv.Atoi(l); err == nil {
			filter.Limit = limit
		}
	}
	if o := c.QueryParam("offset"); o != "" {
		if offset, err := strconv.Atoi(o); err == nil {
			filter.Offset = offset
		}
	}

	authors, total, err := h.svc.List(c.Request().Context(), filter)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list authors")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"records":      authors,
		"totalRecords": total,
	})
}

// GetByID returns an author by ID.
func (h *AuthorHandler) GetByID(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid author id")
	}

	a, err := h.svc.FindByID(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, author.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "author not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get author")
	}

	return c.JSON(http.StatusOK, a)
}

// Create creates a new author.
func (h *AuthorHandler) Create(c echo.Context) error {
	var input author.CreateAuthorInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	a, err := h.svc.Create(c.Request().Context(), input)
	if err != nil {
		if errors.Is(err, author.ErrInvalidInput) {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid author data")
		}
		if errors.Is(err, author.ErrDuplicate) {
			return echo.NewHTTPError(http.StatusConflict, "author already exists")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create author")
	}

	return c.JSON(http.StatusCreated, a)
}

// Update updates an existing author.
func (h *AuthorHandler) Update(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid author id")
	}

	var input author.UpdateAuthorInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	a, err := h.svc.Update(c.Request().Context(), id, input)
	if err != nil {
		if errors.Is(err, author.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "author not found")
		}
		if errors.Is(err, author.ErrDuplicate) {
			return echo.NewHTTPError(http.StatusConflict, "author already exists")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update author")
	}

	return c.JSON(http.StatusOK, a)
}

// Delete deletes an author.
func (h *AuthorHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid author id")
	}

	if err := h.svc.Delete(c.Request().Context(), id); err != nil {
		if errors.Is(err, author.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "author not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete author")
	}

	return c.NoContent(http.StatusNoContent)
}

// GetStats returns statistics for an author.
func (h *AuthorHandler) GetStats(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid author id")
	}

	stats, err := h.svc.GetStats(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get author stats")
	}

	return c.JSON(http.StatusOK, stats)
}
