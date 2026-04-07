package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/woliveiras/bookaneer/internal/core/book"
)

// BookHandler handles book-related HTTP requests.
type BookHandler struct {
	svc *book.Service
}

// NewBookHandler creates a new book handler.
func NewBookHandler(svc *book.Service) *BookHandler {
	return &BookHandler{svc: svc}
}

// Register registers the book routes.
func (h *BookHandler) Register(g *echo.Group) {
	g.GET("/book", h.List)
	g.GET("/book/:id", h.GetByID)
	g.POST("/book", h.Create)
	g.PUT("/book/:id", h.Update)
	g.DELETE("/book/:id", h.Delete)

	// Editions
	g.POST("/book/:id/edition", h.CreateEdition)
	g.DELETE("/edition/:id", h.DeleteEdition)
}

// List returns a list of books.
func (h *BookHandler) List(c echo.Context) error {
	filter := book.ListBooksFilter{
		Search:  c.QueryParam("search"),
		SortBy:  c.QueryParam("sortBy"),
		SortDir: c.QueryParam("sortDir"),
		Missing: c.QueryParam("missing") == "true",
	}

	if a := c.QueryParam("authorId"); a != "" {
		if authorID, err := strconv.ParseInt(a, 10, 64); err == nil {
			filter.AuthorID = &authorID
		}
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

	books, total, err := h.svc.List(c.Request().Context(), filter)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list books")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"records":      books,
		"totalRecords": total,
	})
}

// GetByID returns a book by ID with its editions.
func (h *BookHandler) GetByID(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid book id")
	}

	b, err := h.svc.GetWithEditions(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, book.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "book not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get book")
	}

	return c.JSON(http.StatusOK, b)
}

// Create creates a new book.
func (h *BookHandler) Create(c echo.Context) error {
	var input book.CreateBookInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	b, err := h.svc.Create(c.Request().Context(), input)
	if err != nil {
		if errors.Is(err, book.ErrInvalidInput) {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid book data")
		}
		if errors.Is(err, book.ErrAuthorNotFound) {
			return echo.NewHTTPError(http.StatusBadRequest, "author not found")
		}
		if errors.Is(err, book.ErrDuplicate) {
			return echo.NewHTTPError(http.StatusConflict, "book already exists")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create book")
	}

	return c.JSON(http.StatusCreated, b)
}

// Update updates an existing book.
func (h *BookHandler) Update(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid book id")
	}

	var input book.UpdateBookInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	b, err := h.svc.Update(c.Request().Context(), id, input)
	if err != nil {
		if errors.Is(err, book.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "book not found")
		}
		if errors.Is(err, book.ErrAuthorNotFound) {
			return echo.NewHTTPError(http.StatusBadRequest, "author not found")
		}
		if errors.Is(err, book.ErrDuplicate) {
			return echo.NewHTTPError(http.StatusConflict, "book already exists")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update book")
	}

	return c.JSON(http.StatusOK, b)
}

// Delete deletes a book.
func (h *BookHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid book id")
	}

	if err := h.svc.Delete(c.Request().Context(), id); err != nil {
		if errors.Is(err, book.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "book not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete book")
	}

	return c.NoContent(http.StatusNoContent)
}

// CreateEdition creates a new edition for a book.
func (h *BookHandler) CreateEdition(c echo.Context) error {
	bookID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid book id")
	}

	var input book.CreateEditionInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	input.BookID = bookID

	e, err := h.svc.CreateEdition(c.Request().Context(), input)
	if err != nil {
		if errors.Is(err, book.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "book not found")
		}
		if errors.Is(err, book.ErrInvalidInput) {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid edition data")
		}
		if errors.Is(err, book.ErrDuplicate) {
			return echo.NewHTTPError(http.StatusConflict, "edition already exists")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create edition")
	}

	return c.JSON(http.StatusCreated, e)
}

// DeleteEdition deletes an edition.
func (h *BookHandler) DeleteEdition(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid edition id")
	}

	if err := h.svc.DeleteEdition(c.Request().Context(), id); err != nil {
		if errors.Is(err, book.ErrEditionNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "edition not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete edition")
	}

	return c.NoContent(http.StatusNoContent)
}
