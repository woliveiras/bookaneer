package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"

	"github.com/woliveiras/bookaneer/internal/core/library"
)

// LibraryHandler handles library-related HTTP requests.
type LibraryHandler struct {
	scanner *library.Scanner
}

// NewLibraryHandler creates a new library handler.
func NewLibraryHandler(scanner *library.Scanner) *LibraryHandler {
	return &LibraryHandler{scanner: scanner}
}

// Register registers the library routes.
func (h *LibraryHandler) Register(g *echo.Group) {
	g.POST("/command/libraryscan/rootfolder/:id", h.ScanRootFolder)
	g.POST("/command/libraryscan/author/:id", h.ScanAuthor)
}

// ScanRootFolder triggers a scan of a specific root folder.
func (h *LibraryHandler) ScanRootFolder(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid root folder id")
	}

	result, err := h.scanner.ScanRootFolder(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

// ScanAuthor triggers a scan of a specific author's folder.
func (h *LibraryHandler) ScanAuthor(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid author id")
	}

	result, err := h.scanner.ScanAuthorFolder(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}
