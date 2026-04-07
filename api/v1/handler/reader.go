package handler

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/woliveiras/bookaneer/internal/auth"
	"github.com/woliveiras/bookaneer/internal/core/reader"
)

// ReaderHandler handles reader-related endpoints.
type ReaderHandler struct {
	svc *reader.Service
}

// NewReaderHandler creates a new reader handler.
func NewReaderHandler(svc *reader.Service) *ReaderHandler {
	return &ReaderHandler{svc: svc}
}

// GetBookFile returns book file metadata for the reader.
// GET /api/v1/reader/:id
func (h *ReaderHandler) GetBookFile(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid book file id")
	}

	bf, err := h.svc.GetBookFile(c.Request().Context(), id)
	if err == reader.ErrBookFileNotFound {
		return echo.NewHTTPError(http.StatusNotFound, "book file not found")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get book file")
	}

	return c.JSON(http.StatusOK, bf)
}

// ServeContent streams the book file content.
// GET /api/v1/reader/:id/content
func (h *ReaderHandler) ServeContent(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid book file id")
	}

	bf, err := h.svc.GetBookFile(c.Request().Context(), id)
	if err == reader.ErrBookFileNotFound {
		return echo.NewHTTPError(http.StatusNotFound, "book file not found")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get book file")
	}

	// Check if file exists
	if _, err := os.Stat(bf.Path); os.IsNotExist(err) {
		return echo.NewHTTPError(http.StatusNotFound, "file not found on disk")
	}

	// Open file
	file, err := os.Open(bf.Path)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to open file")
	}
	defer file.Close()

	// Get file info for size
	stat, err := file.Stat()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get file info")
	}

	// Set appropriate content type based on format
	contentType := "application/octet-stream"
	switch bf.Format {
	case "epub":
		contentType = "application/epub+zip"
	case "pdf":
		contentType = "application/pdf"
	case "mobi":
		contentType = "application/x-mobipocket-ebook"
	case "azw3":
		contentType = "application/vnd.amazon.ebook"
	case "fb2":
		contentType = "application/x-fictionbook+xml"
	case "cbz":
		contentType = "application/vnd.comicbook+zip"
	case "cbr":
		contentType = "application/vnd.comicbook-rar"
	}

	// Set headers
	c.Response().Header().Set("Content-Type", contentType)
	c.Response().Header().Set("Content-Length", strconv.FormatInt(stat.Size(), 10))
	c.Response().Header().Set("Content-Disposition", "inline; filename=\""+filepath.Base(bf.Path)+"\"")
	c.Response().Header().Set("Accept-Ranges", "bytes")

	// Stream the file (supports HTTP Range requests via ServeContent)
	http.ServeContent(c.Response(), c.Request(), bf.Path, stat.ModTime(), file)
	return nil
}

// GetProgress returns reading progress for the current user.
// GET /api/v1/reader/:id/progress
func (h *ReaderHandler) GetProgress(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid book file id")
	}

	// Get user from context
	user, ok := c.Get("user").(*auth.User)
	userID := int64(0)
	if ok && user != nil {
		userID = user.ID
	}
	// If no user (system API key), use ID 0
	if userID == 0 {
		// Check for apiKey context (system key auth)
		if _, ok := c.Get("apiKey").(string); ok {
			userID = 0 // System user
		}
	}

	progress, err := h.svc.GetProgress(c.Request().Context(), id, userID)
	if err == reader.ErrProgressNotFound {
		// Return empty progress instead of error
		return c.JSON(http.StatusOK, map[string]interface{}{
			"bookFileId": id,
			"position":   "",
			"percentage": 0,
		})
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get progress")
	}

	return c.JSON(http.StatusOK, progress)
}

// SaveProgressRequest is the request body for saving progress.
type SaveProgressRequest struct {
	Position   string  `json:"position"`
	Percentage float64 `json:"percentage"`
}

// SaveProgress saves reading progress for the current user.
// PUT /api/v1/reader/:id/progress
func (h *ReaderHandler) SaveProgress(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid book file id")
	}

	var req SaveProgressRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	// Get user from context
	user, ok := c.Get("user").(*auth.User)
	userID := int64(0)
	if ok && user != nil {
		userID = user.ID
	}

	progress, err := h.svc.SaveProgress(c.Request().Context(), id, userID, req.Position, req.Percentage)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to save progress")
	}

	return c.JSON(http.StatusOK, progress)
}

// Bookmark handlers

// ListBookmarks returns all bookmarks for a book file.
// GET /api/v1/reader/:id/bookmarks
func (h *ReaderHandler) ListBookmarks(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid book file id")
	}

	user, ok := c.Get("user").(*auth.User)
	userID := int64(0)
	if ok && user != nil {
		userID = user.ID
	}

	bookmarks, err := h.svc.ListBookmarks(c.Request().Context(), id, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list bookmarks")
	}

	if bookmarks == nil {
		bookmarks = []reader.Bookmark{}
	}

	return c.JSON(http.StatusOK, bookmarks)
}

// CreateBookmarkRequest is the request body for creating a bookmark.
type CreateBookmarkRequest struct {
	Position string `json:"position"`
	Title    string `json:"title"`
	Note     string `json:"note"`
}

// CreateBookmark creates a new bookmark.
// POST /api/v1/reader/:id/bookmarks
func (h *ReaderHandler) CreateBookmark(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid book file id")
	}

	var req CreateBookmarkRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if req.Position == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "position is required")
	}

	user, ok := c.Get("user").(*auth.User)
	userID := int64(0)
	if ok && user != nil {
		userID = user.ID
	}

	bookmark, err := h.svc.CreateBookmark(c.Request().Context(), id, userID, req.Position, req.Title, req.Note)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create bookmark")
	}

	return c.JSON(http.StatusCreated, bookmark)
}

// DeleteBookmark deletes a bookmark.
// DELETE /api/v1/reader/:id/bookmarks/:bookmarkId
func (h *ReaderHandler) DeleteBookmark(c echo.Context) error {
	bookmarkID, err := strconv.ParseInt(c.Param("bookmarkId"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid bookmark id")
	}

	user, ok := c.Get("user").(*auth.User)
	userID := int64(0)
	if ok && user != nil {
		userID = user.ID
	}

	err = h.svc.DeleteBookmark(c.Request().Context(), bookmarkID, userID)
	if err == reader.ErrBookmarkNotFound {
		return echo.NewHTTPError(http.StatusNotFound, "bookmark not found")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete bookmark")
	}

	return c.NoContent(http.StatusNoContent)
}

// Register registers the reader routes.
func (h *ReaderHandler) Register(g *echo.Group) {
	g.GET("/reader/:id", h.GetBookFile)
	g.GET("/reader/:id/content", h.ServeContent)
	g.GET("/reader/:id/progress", h.GetProgress)
	g.PUT("/reader/:id/progress", h.SaveProgress)
	g.GET("/reader/:id/bookmarks", h.ListBookmarks)
	g.POST("/reader/:id/bookmarks", h.CreateBookmark)
	g.DELETE("/reader/:id/bookmarks/:bookmarkId", h.DeleteBookmark)
}

// Discard implements io.Writer to discard bytes (used when ServeContent is called).
type discardResponseWriter struct {
	io.Writer
}
