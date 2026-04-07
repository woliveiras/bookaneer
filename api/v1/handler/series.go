package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/woliveiras/bookaneer/internal/core/series"
)

// SeriesHandler handles series-related HTTP requests.
type SeriesHandler struct {
	svc *series.Service
}

// NewSeriesHandler creates a new series handler.
func NewSeriesHandler(svc *series.Service) *SeriesHandler {
	return &SeriesHandler{svc: svc}
}

// Register registers the series routes.
func (h *SeriesHandler) Register(g *echo.Group) {
	g.GET("/series", h.List)
	g.GET("/series/:id", h.GetByID)
	g.POST("/series", h.Create)
	g.PUT("/series/:id", h.Update)
	g.DELETE("/series/:id", h.Delete)

	// Series books
	g.POST("/series/:id/book", h.AddBook)
	g.DELETE("/series/:id/book/:bookId", h.RemoveBook)
	g.PUT("/series/:id/book/:bookId", h.UpdateBookPosition)
}

// List returns a list of series.
func (h *SeriesHandler) List(c echo.Context) error {
	filter := series.ListSeriesFilter{
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

	seriesList, total, err := h.svc.List(c.Request().Context(), filter)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list series")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"records":      seriesList,
		"totalRecords": total,
	})
}

// GetByID returns a series by ID with its books.
func (h *SeriesHandler) GetByID(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid series id")
	}

	s, err := h.svc.GetWithBooks(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, series.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "series not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get series")
	}

	return c.JSON(http.StatusOK, s)
}

// Create creates a new series.
func (h *SeriesHandler) Create(c echo.Context) error {
	var input series.CreateSeriesInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	s, err := h.svc.Create(c.Request().Context(), input)
	if err != nil {
		if errors.Is(err, series.ErrInvalidInput) {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid series data")
		}
		if errors.Is(err, series.ErrDuplicate) {
			return echo.NewHTTPError(http.StatusConflict, "series already exists")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create series")
	}

	return c.JSON(http.StatusCreated, s)
}

// Update updates an existing series.
func (h *SeriesHandler) Update(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid series id")
	}

	var input series.UpdateSeriesInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	s, err := h.svc.Update(c.Request().Context(), id, input)
	if err != nil {
		if errors.Is(err, series.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "series not found")
		}
		if errors.Is(err, series.ErrDuplicate) {
			return echo.NewHTTPError(http.StatusConflict, "series already exists")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update series")
	}

	return c.JSON(http.StatusOK, s)
}

// Delete deletes a series.
func (h *SeriesHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid series id")
	}

	if err := h.svc.Delete(c.Request().Context(), id); err != nil {
		if errors.Is(err, series.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "series not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete series")
	}

	return c.NoContent(http.StatusNoContent)
}

// AddBook adds a book to a series.
func (h *SeriesHandler) AddBook(c echo.Context) error {
	seriesID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid series id")
	}

	var input series.AddBookInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if err := h.svc.AddBook(c.Request().Context(), seriesID, input); err != nil {
		if errors.Is(err, series.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "series not found")
		}
		if errors.Is(err, series.ErrBookNotFound) {
			return echo.NewHTTPError(http.StatusBadRequest, "book not found")
		}
		if errors.Is(err, series.ErrBookAlreadyInSeries) {
			return echo.NewHTTPError(http.StatusConflict, "book already in series")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to add book to series")
	}

	return c.NoContent(http.StatusCreated)
}

// RemoveBook removes a book from a series.
func (h *SeriesHandler) RemoveBook(c echo.Context) error {
	seriesID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid series id")
	}

	bookID, err := strconv.ParseInt(c.Param("bookId"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid book id")
	}

	if err := h.svc.RemoveBook(c.Request().Context(), seriesID, bookID); err != nil {
		if errors.Is(err, series.ErrBookNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "book not in series")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to remove book from series")
	}

	return c.NoContent(http.StatusNoContent)
}

// UpdateBookPosition updates the position of a book in a series.
func (h *SeriesHandler) UpdateBookPosition(c echo.Context) error {
	seriesID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid series id")
	}

	bookID, err := strconv.ParseInt(c.Param("bookId"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid book id")
	}

	var input struct {
		Position string `json:"position"`
	}
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if err := h.svc.UpdateBookPosition(c.Request().Context(), seriesID, bookID, input.Position); err != nil {
		if errors.Is(err, series.ErrBookNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "book not in series")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update book position")
	}

	return c.NoContent(http.StatusOK)
}
